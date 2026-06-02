package agents

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	jxlog "github.com/Inkedi9/jarvinx/jxlog"
	"github.com/Inkedi9/jarvinx/llm"
	"github.com/Inkedi9/jarvinx/memory"
)

const qdrantCollection = "jarvinx_decisions"

// QdrantAgent vectorise chaque décision LLM via Ollama et la stocke dans Qdrant.
// Actif seulement si JARVINX_QDRANT_URL est défini (opt-in).
// Pattern N-1 : lit la décision du cycle précédent, l'embeddise, stocke dans Qdrant.
// Deux circuit breakers indépendants : un pour Ollama embedding, un pour Qdrant HTTP.
// Toute erreur est fail-silent : le cycle 15s n'est jamais bloqué.
type QdrantAgent struct {
	BaseAgent
	qdrantURL     string
	embedder      *llm.OllamaEmbedder
	httpClient    *http.Client
	qdrantCircuit *llm.CircuitBreaker

	// collMu + collReady remplacent sync.Once : retentative à chaque cycle si Qdrant était down.
	collMu    sync.Mutex
	collReady bool
}

func NewQdrantAgent(qdrantURL, ollamaURL, embedModel string) *QdrantAgent {
	return &QdrantAgent{
		BaseAgent:     NewBaseAgent("qdrant", 15*time.Second),
		qdrantURL:     qdrantURL,
		embedder:      llm.NewOllamaEmbedder(ollamaURL, embedModel),
		httpClient:    &http.Client{Timeout: 10 * time.Second},
		qdrantCircuit: llm.DefaultCircuitBreaker(),
	}
}

func (a *QdrantAgent) Run(ctx context.Context, actx AgentContext) error {
	cycles := actx.State.LastCycles(1)
	if len(cycles) == 0 {
		jxlog.Debug("QDRANT AGENT", "aucune décision disponible — cycle ignoré")
		a.recordSuccess()
		return nil
	}

	last := cycles[0]
	text := fmt.Sprintf("[%s] %s. %s", last.Action, last.Analysis, last.Reason)

	// Embedding Ollama — circuit breaker interne à OllamaEmbedder (3 échecs → open 30s)
	vector, err := a.embedder.Embed(ctx, text)
	if err != nil {
		jxlog.Warn("QDRANT AGENT", fmt.Sprintf("embedding échoué (cycle %d) : %v", last.CycleNum, err))
		a.recordSuccess() // fail-silent
		return nil
	}

	// Qdrant collection — retentative à chaque cycle si init précédente a échoué
	if err := a.ensureCollectionOnce(ctx, len(vector)); err != nil {
		jxlog.Warn("QDRANT AGENT", fmt.Sprintf("init collection échouée : %v", err))
		a.recordSuccess() // fail-silent
		return nil
	}

	// Qdrant upsert — circuit breaker Qdrant (3 échecs → open 30s)
	if err := a.upsert(ctx, last, vector); err != nil {
		jxlog.Warn("QDRANT AGENT", fmt.Sprintf("upsert échoué (cycle %d) : %v", last.CycleNum, err))
		a.recordSuccess() // fail-silent
		return nil
	}

	jxlog.Debug("QDRANT AGENT", fmt.Sprintf(
		"cycle %d vectorisé — action=%s confidence=%.2f dim=%d",
		last.CycleNum, last.Action, last.Confidence, len(vector),
	))

	a.recordSuccess()
	return nil
}

// ensureCollectionOnce tente de créer la collection si pas encore confirmée.
// Contrairement à sync.Once, elle retente à chaque cycle jusqu'au premier succès.
func (a *QdrantAgent) ensureCollectionOnce(ctx context.Context, dim int) error {
	a.collMu.Lock()
	defer a.collMu.Unlock()
	if a.collReady {
		return nil
	}
	if err := a.ensureCollection(ctx, dim); err != nil {
		return err // retentative au prochain cycle
	}
	a.collReady = true
	return nil
}

// ensureCollection vérifie l'existence de la collection et la crée si besoin.
// Protégée par le circuit breaker Qdrant.
func (a *QdrantAgent) ensureCollection(ctx context.Context, dim int) error {
	if err := a.qdrantCircuit.Allow(); err != nil {
		return err
	}

	url := fmt.Sprintf("%s/collections/%s", a.qdrantURL, qdrantCollection)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("check collection request: %w", err)
	}
	resp, err := a.httpClient.Do(req)
	if err != nil {
		a.qdrantCircuit.RecordFailure()
		return fmt.Errorf("check collection: %w", err)
	}
	_ = resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		a.qdrantCircuit.RecordSuccess()
		return nil // existe déjà
	}
	if resp.StatusCode != http.StatusNotFound {
		a.qdrantCircuit.RecordFailure()
		return fmt.Errorf("check collection: status inattendu %d", resp.StatusCode)
	}

	// Crée la collection
	body, _ := json.Marshal(map[string]any{
		"vectors": map[string]any{
			"size":     dim,
			"distance": "Cosine",
		},
	})
	req, err = http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("create collection request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err = a.httpClient.Do(req)
	if err != nil {
		a.qdrantCircuit.RecordFailure()
		return fmt.Errorf("create collection: %w", err)
	}
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		a.qdrantCircuit.RecordFailure()
		return fmt.Errorf("create collection: status %d", resp.StatusCode)
	}

	a.qdrantCircuit.RecordSuccess()
	jxlog.Info("QDRANT AGENT", fmt.Sprintf("collection '%s' créée (dim=%d, distance=Cosine)", qdrantCollection, dim))
	return nil
}

// upsert envoie un point vectorisé dans Qdrant avec les métadonnées de la décision.
// Protégée par le circuit breaker Qdrant.
func (a *QdrantAgent) upsert(ctx context.Context, record memory.CycleRecord, vector []float32) error {
	if err := a.qdrantCircuit.Allow(); err != nil {
		return err
	}

	type point struct {
		ID      uint64         `json:"id"`
		Vector  []float32      `json:"vector"`
		Payload map[string]any `json:"payload"`
	}
	type upsertBody struct {
		Points []point `json:"points"`
	}

	p := point{
		ID:     uint64(record.CycleNum),
		Vector: vector,
		Payload: map[string]any{
			"action":       record.Action,
			"analysis":     record.Analysis,
			"reason":       record.Reason,
			"confidence":   record.Confidence,
			"cycle_num":    record.CycleNum,
			"trigger_cpu":  record.TriggerCPU,
			"trigger_ram":  record.TriggerRAM,
			"trigger_disk": record.TriggerDisk,
			"timestamp":    record.Timestamp.UTC().Format(time.RFC3339),
		},
	}

	body, err := json.Marshal(upsertBody{Points: []point{p}})
	if err != nil {
		return fmt.Errorf("marshal upsert: %w", err)
	}

	url := fmt.Sprintf("%s/collections/%s/points", a.qdrantURL, qdrantCollection)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("create upsert request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		a.qdrantCircuit.RecordFailure()
		return fmt.Errorf("qdrant upsert: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		a.qdrantCircuit.RecordFailure()
		return fmt.Errorf("qdrant upsert: status %d", resp.StatusCode)
	}

	a.qdrantCircuit.RecordSuccess()
	return nil
}
