package agents

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Inkedi9/jarvinx/llm"
	"github.com/Inkedi9/jarvinx/memory"
)

const (
	testQdrantURL  = "http://localhost:6333"
	testOllamaURL  = "http://localhost:11434"
	testEmbedModel = "nomic-embed-text"
)

// mockQdrantServer simule un Qdrant sain : GET 404, PUT collection 200, PUT/POST points 200.
func mockQdrantServer(t *testing.T, searchHits []map[string]any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet:
			w.WriteHeader(http.StatusNotFound)
		case r.Method == http.MethodPut && r.URL.Path == "/collections/jarvinx_decisions":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{"result": true})
		case r.Method == http.MethodPut && r.URL.Path == "/collections/jarvinx_decisions/points":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"status": "acknowledged"}})
		case r.Method == http.MethodPost && r.URL.Path == "/collections/jarvinx_decisions/points/search":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{"result": searchHits})
		default:
			http.NotFound(w, r)
		}
	}))
}

// mockOllamaEmbedServer simule Ollama /api/embeddings.
func mockOllamaEmbedServer(t *testing.T, embedding []float32) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"embedding": embedding})
	}))
}

func addCycle(t *testing.T, store memory.Store, action string, cycleNum int) {
	t.Helper()
	_ = store.AddCycle(memory.CycleRecord{
		Action: action, Analysis: "CPU stable", Reason: "no anomaly",
		CycleNum: cycleNum, Confidence: 0.9, Timestamp: time.Now(),
	})
}

// --- Tests de base ---

func TestQdrantAgent_SkipsWhenNoDecision(t *testing.T) {
	a := NewQdrantAgent(testQdrantURL, testOllamaURL, testEmbedModel)
	store := memory.NewState("state_qdrant_test.json")

	err := a.Run(context.Background(), AgentContext{State: store, Logger: memory.NewLogger("")})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if a.Status().RunCount != 1 {
		t.Errorf("expected RunCount=1, got %d", a.Status().RunCount)
	}
}

func TestQdrantAgent_FailSilentOnEmbedError(t *testing.T) {
	a := NewQdrantAgent(testQdrantURL, "http://localhost:19999", testEmbedModel)
	store := memory.NewState("")
	addCycle(t, store, "log", 1)

	err := a.Run(context.Background(), AgentContext{State: store, Logger: memory.NewLogger("")})
	if err != nil {
		t.Fatalf("expected nil (fail-silent), got: %v", err)
	}
}

func TestQdrantAgent_EmbedAndUpsert(t *testing.T) {
	embedding := []float32{0.1, 0.2, 0.3, 0.4}
	ollamaSrv := mockOllamaEmbedServer(t, embedding)
	defer ollamaSrv.Close()
	qdrantSrv := mockQdrantServer(t, nil)
	defer qdrantSrv.Close()

	a := NewQdrantAgent(qdrantSrv.URL, ollamaSrv.URL, testEmbedModel)
	store := memory.NewState("")
	addCycle(t, store, "log", 5)

	err := a.Run(context.Background(), AgentContext{State: store, Logger: memory.NewLogger("")})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	a.collMu.Lock()
	ready := a.collReady
	a.collMu.Unlock()
	if !ready {
		t.Error("expected collReady=true after successful upsert")
	}
}

func TestNewQdrantAgent_DefaultSchedule(t *testing.T) {
	a := NewQdrantAgent(testQdrantURL, testOllamaURL, testEmbedModel)
	if a.Schedule() != 15*time.Second {
		t.Errorf("expected 15s schedule, got %v", a.Schedule())
	}
	if a.Name() != "qdrant" {
		t.Errorf("expected name 'qdrant', got '%s'", a.Name())
	}
}

// --- Tests recherche sémantique ---

func TestQdrantAgent_SearchPopulatesLastSimilar(t *testing.T) {
	embedding := []float32{0.1, 0.2, 0.3, 0.4}

	hits := []map[string]any{
		{
			"score": float64(0.92),
			"payload": map[string]any{
				"action": "log", "analysis": "CPU stable", "reason": "no anomaly",
				"confidence": float64(0.9),
			},
		},
		{
			"score": float64(0.50), // sous minSimilarityScore — doit être filtré
			"payload": map[string]any{
				"action": "alert", "analysis": "CPU high", "reason": "spike detected",
				"confidence": float64(0.8),
			},
		},
	}

	ollamaSrv := mockOllamaEmbedServer(t, embedding)
	defer ollamaSrv.Close()
	qdrantSrv := mockQdrantServer(t, hits)
	defer qdrantSrv.Close()

	a := NewQdrantAgent(qdrantSrv.URL, ollamaSrv.URL, testEmbedModel)
	store := memory.NewState("")
	addCycle(t, store, "log", 5)

	// Premier Run : embed + upsert + init collection → collReady = true
	_ = a.Run(context.Background(), AgentContext{
		Snapshot: memory.Snapshot{CPUPercent: 40.0, MemPercent: 55.0, DiskPercent: 30.0},
		State:    store,
		Logger:   memory.NewLogger(""),
	})

	similar := a.LastSimilarDecisions()
	// Seul le hit score >= 0.6 doit apparaître
	if len(similar) != 1 {
		t.Fatalf("expected 1 similar decision (filtered score<0.6), got %d: %v", len(similar), similar)
	}
	if similar[0] == "" {
		t.Error("expected non-empty similar decision string")
	}
}

func TestQdrantAgent_SearchSkippedWhenCollectionNotReady(t *testing.T) {
	// Qdrant down : ensureCollection échoue → collReady false → search jamais appelé
	embedding := []float32{0.1, 0.2}
	ollamaSrv := mockOllamaEmbedServer(t, embedding)
	defer ollamaSrv.Close()

	qdrantSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer qdrantSrv.Close()

	a := NewQdrantAgent(qdrantSrv.URL, ollamaSrv.URL, testEmbedModel)
	store := memory.NewState("")
	addCycle(t, store, "log", 1)

	_ = a.Run(context.Background(), AgentContext{State: store, Logger: memory.NewLogger("")})

	if got := a.LastSimilarDecisions(); len(got) != 0 {
		t.Errorf("expected no similar decisions when collection not ready, got %v", got)
	}
}

// --- Tests circuit breaker ---

func TestQdrantAgent_CollectionInitRetry(t *testing.T) {
	embedding := []float32{0.1, 0.2, 0.3}
	ollamaSrv := mockOllamaEmbedServer(t, embedding)
	defer ollamaSrv.Close()

	var calls atomic.Int32
	qdrantSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		switch {
		case r.Method == http.MethodGet:
			w.WriteHeader(http.StatusNotFound)
		case r.Method == http.MethodPut && r.URL.Path == "/collections/jarvinx_decisions":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{"result": true})
		case r.Method == http.MethodPut && r.URL.Path == "/collections/jarvinx_decisions/points":
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost:
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{"result": []any{}})
		}
	}))
	defer qdrantSrv.Close()

	a := NewQdrantAgent(qdrantSrv.URL, ollamaSrv.URL, testEmbedModel)
	store := memory.NewState("")
	addCycle(t, store, "log", 1)
	ctx := context.Background()
	actx := AgentContext{State: store, Logger: memory.NewLogger("")}

	_ = a.Run(ctx, actx)
	a.collMu.Lock()
	if a.collReady {
		t.Error("collReady doit rester false après échec init")
	}
	a.collMu.Unlock()

	_ = a.Run(ctx, actx)
	a.collMu.Lock()
	if !a.collReady {
		t.Error("collReady doit être true après init réussie")
	}
	a.collMu.Unlock()
}

func TestCyclePointID_Uniqueness(t *testing.T) {
	// Same instance, different cycles → different IDs
	id1 := cyclePointID("instance-A", 1)
	id2 := cyclePointID("instance-A", 2)
	if id1 == id2 {
		t.Error("same instance different cycles must produce different point IDs")
	}

	// Different instances, same cycle → different IDs
	id3 := cyclePointID("instance-B", 1)
	if id1 == id3 {
		t.Error("different instances same cycle must produce different point IDs")
	}

	// Deterministic: same inputs → same output
	if cyclePointID("instance-A", 1) != id1 {
		t.Error("cyclePointID must be deterministic")
	}
}

func TestQdrantAgent_QdrantCircuitBreaker(t *testing.T) {
	embedding := []float32{0.1, 0.2, 0.3}
	ollamaSrv := mockOllamaEmbedServer(t, embedding)
	defer ollamaSrv.Close()

	qdrantSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer qdrantSrv.Close()

	a := NewQdrantAgent(qdrantSrv.URL, ollamaSrv.URL, testEmbedModel)
	store := memory.NewState("")
	addCycle(t, store, "log", 1)
	ctx := context.Background()
	actx := AgentContext{State: store, Logger: memory.NewLogger("")}

	for range 3 {
		_ = a.Run(ctx, actx)
	}

	if a.qdrantCircuit.State() != llm.StateOpen {
		t.Errorf("expected Qdrant circuit open after 3 failures, got: %s", a.qdrantCircuit.State())
	}

	start := time.Now()
	_ = a.Run(ctx, actx)
	if elapsed := time.Since(start); elapsed > 500*time.Millisecond {
		t.Errorf("expected near-instant response with open circuit, got %v", elapsed)
	}
}
