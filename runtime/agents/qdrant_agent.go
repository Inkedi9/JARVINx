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
// Toute erreur embedding/réseau est fail-silent : le cycle 15s n'est jamais bloqué.
type QdrantAgent struct {
	BaseAgent
	qdrantURL  string
	embedder   *llm.OllamaEmbedder
	httpClient *http.Client
	initOnce   sync.Once
	initErr    error
}

func NewQdrantAgent(qdrantURL, ollamaURL, embedModel string) *QdrantAgent {
	return &QdrantAgent{
		BaseAgent:  NewBaseAgent("qdrant", 15*time.Second),
		qdrantURL:  qdrantURL,
		embedder:   llm.NewOllamaEmbedder(ollamaURL, embedModel),
		httpClient: &http.Client{Timeout: 10 * time.Second},
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

	vector, err := a.embedder.Embed(ctx, text)
	if err != nil {
		jxlog.Warn("QDRANT AGENT", fmt.Sprintf("embedding échoué (cycle %d) : %v", last.CycleNum, err))
		a.recordSuccess() // fail-silent
		return nil
	}

	// Crée la collection Qdrant au premier appel réussi — on connaît la dimension ici
	a.initOnce.Do(func() {
		a.initErr = a.ensureCollection(ctx, len(vector))
	})
	if a.initErr != nil {
		jxlog.Warn("QDRANT AGENT", fmt.Sprintf("init collection échouée : %v", a.initErr))
		a.recordSuccess() // fail-silent
		return nil
	}

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

// ensureCollection crée la collection Qdrant si elle n'existe pas encore.
func (a *QdrantAgent) ensureCollection(ctx context.Context, dim int) error {
	url := fmt.Sprintf("%s/collections/%s", a.qdrantURL, qdrantCollection)

	// Vérifie si la collection existe
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("check collection request: %w", err)
	}
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("check collection: %w", err)
	}
	_ = resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil // existe déjà
	}
	if resp.StatusCode != http.StatusNotFound {
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
		return fmt.Errorf("create collection: %w", err)
	}
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("create collection: status %d", resp.StatusCode)
	}

	jxlog.Info("QDRANT AGENT", fmt.Sprintf("collection '%s' créée (dim=%d, distance=Cosine)", qdrantCollection, dim))
	return nil
}

// upsert envoie un point vectorisé dans Qdrant avec les métadonnées de la décision.
func (a *QdrantAgent) upsert(ctx context.Context, record memory.CycleRecord, vector []float32) error {
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
		return fmt.Errorf("qdrant upsert: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("qdrant upsert: status %d", resp.StatusCode)
	}

	return nil
}
