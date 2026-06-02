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

// mockQdrantServer simule un Qdrant sain : GET 404, PUT collection 200, PUT points 200.
func mockQdrantServer(t *testing.T) *httptest.Server {
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
	if a.Status().RunCount != 1 {
		t.Errorf("expected RunCount=1, got %d", a.Status().RunCount)
	}
}

func TestQdrantAgent_EmbedAndUpsert(t *testing.T) {
	embedding := []float32{0.1, 0.2, 0.3, 0.4}
	ollamaSrv := mockOllamaEmbedServer(t, embedding)
	defer ollamaSrv.Close()
	qdrantSrv := mockQdrantServer(t)
	defer qdrantSrv.Close()

	a := NewQdrantAgent(qdrantSrv.URL, ollamaSrv.URL, testEmbedModel)
	store := memory.NewState("")
	addCycle(t, store, "log", 5)

	err := a.Run(context.Background(), AgentContext{State: store, Logger: memory.NewLogger("")})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if a.Status().RunCount != 1 {
		t.Errorf("expected RunCount=1, got %d", a.Status().RunCount)
	}
	// collReady doit être true après un upsert réussi
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

// --- Tests circuit breaker ---

// TestQdrantAgent_CollectionInitRetry vérifie que l'init Qdrant est retentée si elle a échoué.
// Serveur en erreur au premier Run, sain au second → collReady devient true.
func TestQdrantAgent_CollectionInitRetry(t *testing.T) {
	embedding := []float32{0.1, 0.2, 0.3}
	ollamaSrv := mockOllamaEmbedServer(t, embedding)
	defer ollamaSrv.Close()

	var calls atomic.Int32
	qdrantSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n == 1 {
			// Premier appel (GET check) — Qdrant down
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		// Appels suivants — Qdrant sain
		switch {
		case r.Method == http.MethodGet:
			w.WriteHeader(http.StatusNotFound)
		case r.Method == http.MethodPut && r.URL.Path == "/collections/jarvinx_decisions":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{"result": true})
		case r.Method == http.MethodPut && r.URL.Path == "/collections/jarvinx_decisions/points":
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer qdrantSrv.Close()

	a := NewQdrantAgent(qdrantSrv.URL, ollamaSrv.URL, testEmbedModel)
	store := memory.NewState("")
	addCycle(t, store, "log", 1)
	ctx := context.Background()
	actx := AgentContext{State: store, Logger: memory.NewLogger("")}

	// Cycle 1 : init échoue (503) — fail-silent, collReady reste false
	_ = a.Run(ctx, actx)
	a.collMu.Lock()
	if a.collReady {
		t.Error("collReady doit rester false après un échec d'init")
	}
	a.collMu.Unlock()

	// Cycle 2 : Qdrant est sain — init réussit
	_ = a.Run(ctx, actx)
	a.collMu.Lock()
	if !a.collReady {
		t.Error("collReady doit être true après un init réussi")
	}
	a.collMu.Unlock()
}

// TestQdrantAgent_QdrantCircuitBreaker vérifie que le CB Qdrant s'ouvre après 3 échecs
// et bloque les appels sans attendre le timeout HTTP.
func TestQdrantAgent_QdrantCircuitBreaker(t *testing.T) {
	embedding := []float32{0.1, 0.2, 0.3}
	ollamaSrv := mockOllamaEmbedServer(t, embedding)
	defer ollamaSrv.Close()

	// Qdrant toujours en erreur
	qdrantSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer qdrantSrv.Close()

	a := NewQdrantAgent(qdrantSrv.URL, ollamaSrv.URL, testEmbedModel)
	store := memory.NewState("")
	addCycle(t, store, "log", 1)
	ctx := context.Background()
	actx := AgentContext{State: store, Logger: memory.NewLogger("")}

	// 3 cycles : chaque échec incrémente le compteur du CB
	for range 3 {
		_ = a.Run(ctx, actx)
	}

	// Le CB Qdrant doit être ouvert
	if a.qdrantCircuit.State() != llm.StateOpen {
		t.Errorf("expected Qdrant circuit open after 3 failures, got: %s", a.qdrantCircuit.State())
	}

	// Le 4e appel doit retourner ErrCircuitOpen sans appel HTTP (immédiat)
	start := time.Now()
	_ = a.Run(ctx, actx)
	elapsed := time.Since(start)
	if elapsed > 500*time.Millisecond {
		t.Errorf("expected near-instant response with open circuit, got %v", elapsed)
	}
	if a.Status().RunCount != 4 {
		t.Errorf("expected RunCount=4 (fail-silent), got %d", a.Status().RunCount)
	}
}
