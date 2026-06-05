package agents

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"net/http"
	"sync"
	"time"

	jxlog "github.com/Inkedi9/jarvinx/jxlog"
	"github.com/Inkedi9/jarvinx/llm"
	"github.com/Inkedi9/jarvinx/memory"
)

const (
	qdrantCollection   = "jarvinx_decisions"
	similarDecisionsK  = 3   // top-K résultats retournés à chaque recherche
	minSimilarityScore = 0.6 // seuil cosine minimum — en dessous, ignoré
)

// QdrantAgent vectorise chaque décision LLM (N-1) et interroge la mémoire sémantique.
// Actif seulement si JARVINX_QDRANT_URL est défini (opt-in).
// Implémente SimilarDecisionsProvider : l'Orchestrateur lit LastSimilarDecisions()
// pour injecter les décisions passées pertinentes dans le prompt du cycle suivant.
type QdrantAgent struct {
	BaseAgent
	qdrantURL     string
	embedder      *llm.OllamaEmbedder
	httpClient    *http.Client
	qdrantCircuit *llm.CircuitBreaker
	instanceID    string // unique per process — prevents point ID collisions on cycle counter reset

	// collMu + collReady : retentative à chaque cycle jusqu'au premier succès Qdrant
	collMu    sync.Mutex
	collReady bool

	// sdMu + lastSimilar : décisions similaires du cycle précédent, lues par l'Orchestrateur
	sdMu        sync.RWMutex
	lastSimilar []string
}

func NewQdrantAgent(qdrantURL, ollamaURL, embedModel string) *QdrantAgent {
	return &QdrantAgent{
		BaseAgent:     NewBaseAgent("qdrant", 15*time.Second),
		qdrantURL:     qdrantURL,
		embedder:      llm.NewOllamaEmbedder(ollamaURL, embedModel),
		httpClient:    &http.Client{Timeout: 10 * time.Second},
		qdrantCircuit: llm.DefaultCircuitBreaker(),
		instanceID:    newInstanceID(),
	}
}

// newInstanceID generates a random 8-byte hex string to uniquely identify this runtime instance.
func newInstanceID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// cyclePointID hashes instance_id:cycle_num into a uint64 Qdrant point ID.
// Prevents collisions when the cycle counter resets (e.g. SQLite wipe + restart).
func cyclePointID(instanceID string, cycleNum int) uint64 {
	h := fnv.New64a()
	_, _ = fmt.Fprintf(h, "%s:%d", instanceID, cycleNum)
	return h.Sum64()
}

// LastSimilarDecisions implémente SimilarDecisionsProvider.
// Retourne les décisions du dernier cycle de recherche réussi — nil si aucune.
func (a *QdrantAgent) LastSimilarDecisions() []string {
	a.sdMu.RLock()
	defer a.sdMu.RUnlock()
	return a.lastSimilar
}

func (a *QdrantAgent) Run(ctx context.Context, actx AgentContext) error {
	// Phase 1 : vectoriser et stocker la décision du cycle précédent (N-1)
	a.storeDecision(ctx, actx)

	// Phase 2 : rechercher des décisions similaires pour le prochain cycle
	a.updateSimilarDecisions(ctx, actx)

	a.recordSuccess()
	return nil
}

// storeDecision embed la dernière décision LLM et l'upserte dans Qdrant.
func (a *QdrantAgent) storeDecision(ctx context.Context, actx AgentContext) {
	cycles := actx.State.LastCycles(1)
	if len(cycles) == 0 {
		jxlog.Debug("QDRANT AGENT", "aucune décision disponible — upsert ignoré")
		return
	}

	last := cycles[0]
	text := fmt.Sprintf("[%s] %s. %s", last.Action, last.Analysis, last.Reason)

	vector, err := a.embedder.Embed(ctx, text)
	if err != nil {
		jxlog.Warn("QDRANT AGENT", fmt.Sprintf("embedding échoué (cycle %d) : %v", last.CycleNum, err))
		return
	}

	if err := a.ensureCollectionOnce(ctx, len(vector)); err != nil {
		jxlog.Warn("QDRANT AGENT", fmt.Sprintf("init collection échouée : %v", err))
		return
	}

	if err := a.upsert(ctx, last, vector); err != nil {
		jxlog.Warn("QDRANT AGENT", fmt.Sprintf("upsert échoué (cycle %d) : %v", last.CycleNum, err))
		return
	}

	jxlog.Debug("QDRANT AGENT", fmt.Sprintf(
		"cycle %d vectorisé — action=%s confidence=%.2f dim=%d",
		last.CycleNum, last.Action, last.Confidence, len(vector),
	))
}

// updateSimilarDecisions embed le snapshot courant et interroge Qdrant pour trouver
// les décisions passées les plus proches. Résultat stocké dans lastSimilar pour le cycle suivant.
func (a *QdrantAgent) updateSimilarDecisions(ctx context.Context, actx AgentContext) {
	a.collMu.Lock()
	ready := a.collReady
	a.collMu.Unlock()
	if !ready {
		return // collection pas encore initialisée
	}

	snap := actx.Snapshot
	// Texte de requête décrivant l'état courant — même espace sémantique que les décisions stockées
	queryText := fmt.Sprintf("[observe] CPU:%.1f%% RAM:%.1f%% Disk:%.1f%%.",
		snap.CPUPercent, snap.MemPercent, snap.DiskPercent)

	queryVec, err := a.embedder.Embed(ctx, queryText)
	if err != nil {
		jxlog.Warn("QDRANT AGENT", fmt.Sprintf("search embed échoué : %v", err))
		return
	}

	similar, err := a.search(ctx, queryVec, similarDecisionsK)
	if err != nil {
		jxlog.Warn("QDRANT AGENT", fmt.Sprintf("search Qdrant échoué : %v", err))
		return
	}

	a.sdMu.Lock()
	a.lastSimilar = similar
	a.sdMu.Unlock()

	if len(similar) > 0 {
		jxlog.Debug("QDRANT AGENT", fmt.Sprintf("%d décisions similaires trouvées", len(similar)))
	}
}

// ensureCollectionOnce retente à chaque cycle jusqu'au premier succès Qdrant.
func (a *QdrantAgent) ensureCollectionOnce(ctx context.Context, dim int) error {
	a.collMu.Lock()
	defer a.collMu.Unlock()
	if a.collReady {
		return nil
	}
	if err := a.ensureCollection(ctx, dim); err != nil {
		return err
	}
	a.collReady = true
	return nil
}

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
		return nil
	}
	if resp.StatusCode != http.StatusNotFound {
		a.qdrantCircuit.RecordFailure()
		return fmt.Errorf("check collection: status inattendu %d", resp.StatusCode)
	}

	body, _ := json.Marshal(map[string]any{
		"vectors": map[string]any{"size": dim, "distance": "Cosine"},
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
		ID:     cyclePointID(a.instanceID, record.CycleNum),
		Vector: vector,
		Payload: map[string]any{
			"action":       record.Action,
			"analysis":     record.Analysis,
			"reason":       record.Reason,
			"confidence":   record.Confidence,
			"cycle_num":    record.CycleNum,
			"instance_id":  a.instanceID,
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

// search interroge Qdrant pour les k décisions les plus proches du vecteur query.
// Filtre les résultats sous minSimilarityScore. Protégé par le circuit breaker Qdrant.
func (a *QdrantAgent) search(ctx context.Context, queryVec []float32, k int) ([]string, error) {
	if err := a.qdrantCircuit.Allow(); err != nil {
		return nil, err
	}

	type searchReq struct {
		Vector      []float32 `json:"vector"`
		Limit       int       `json:"limit"`
		WithPayload bool      `json:"with_payload"`
	}
	type searchHit struct {
		Score   float32        `json:"score"`
		Payload map[string]any `json:"payload"`
	}
	type searchResp struct {
		Result []searchHit `json:"result"`
	}

	body, err := json.Marshal(searchReq{Vector: queryVec, Limit: k, WithPayload: true})
	if err != nil {
		return nil, fmt.Errorf("marshal search: %w", err)
	}

	url := fmt.Sprintf("%s/collections/%s/points/search", a.qdrantURL, qdrantCollection)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("create search request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		a.qdrantCircuit.RecordFailure()
		return nil, fmt.Errorf("qdrant search: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		a.qdrantCircuit.RecordFailure()
		return nil, fmt.Errorf("qdrant search: status %d", resp.StatusCode)
	}

	var out searchResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		a.qdrantCircuit.RecordFailure()
		return nil, fmt.Errorf("decode search response: %w", err)
	}

	a.qdrantCircuit.RecordSuccess()

	var results []string
	for _, hit := range out.Result {
		if hit.Score < minSimilarityScore {
			continue
		}
		action, _ := hit.Payload["action"].(string)
		analysis, _ := hit.Payload["analysis"].(string)
		reason, _ := hit.Payload["reason"].(string)
		confidence, _ := hit.Payload["confidence"].(float64)
		results = append(results, fmt.Sprintf(
			"[%s] %s. %s (score:%.2f conf:%.2f)",
			action, analysis, reason, hit.Score, confidence,
		))
	}

	return results, nil
}
