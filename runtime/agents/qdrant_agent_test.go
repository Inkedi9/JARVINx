package agents

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Inkedi9/jarvinx/memory"
)

const (
	testQdrantURL = "http://localhost:6333"
	testOllamaURL = "http://localhost:11434"
	testEmbedModel = "nomic-embed-text"
)

func mockQdrantServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet: // ensureCollection check
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

func mockOllamaEmbedServer(t *testing.T, embedding []float32) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"embedding": embedding})
	}))
}

func TestQdrantAgent_SkipsWhenNoDecision(t *testing.T) {
	a := NewQdrantAgent(testQdrantURL, testOllamaURL, testEmbedModel)
	store := memory.NewState("state_qdrant_test.json")

	err := a.Run(context.Background(), AgentContext{
		Snapshot: memory.Snapshot{},
		State:    store,
		Logger:   memory.NewLogger(""),
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if a.Status().RunCount != 1 {
		t.Errorf("expected RunCount=1, got %d", a.Status().RunCount)
	}
}

func TestQdrantAgent_FailSilentOnEmbedError(t *testing.T) {
	// Ollama down — embedding échoue mais le cycle continue
	a := NewQdrantAgent(testQdrantURL, "http://localhost:19999", testEmbedModel)
	store := memory.NewState("")
	_ = store.AddCycle(memory.CycleRecord{
		Action: "log", Analysis: "CPU stable", Reason: "ok",
		CycleNum: 1, Timestamp: time.Now(),
	})

	err := a.Run(context.Background(), AgentContext{State: store, Logger: memory.NewLogger("")})
	if err != nil {
		t.Fatalf("expected nil (fail-silent), got: %v", err)
	}
	if a.Status().RunCount != 1 {
		t.Errorf("expected RunCount=1, got %d", a.Status().RunCount)
	}
}

func TestQdrantAgent_EmbedAndUpsert(t *testing.T) {
	embedding := make([]float32, 4)
	for i := range embedding {
		embedding[i] = float32(i) * 0.1
	}

	ollamaSrv := mockOllamaEmbedServer(t, embedding)
	defer ollamaSrv.Close()
	qdrantSrv := mockQdrantServer(t)
	defer qdrantSrv.Close()

	a := NewQdrantAgent(qdrantSrv.URL, ollamaSrv.URL, testEmbedModel)
	store := memory.NewState("")
	_ = store.AddCycle(memory.CycleRecord{
		Action: "log", Analysis: "CPU stable", Reason: "no anomaly detected",
		CycleNum: 5, Confidence: 0.9, Timestamp: time.Now(),
	})

	err := a.Run(context.Background(), AgentContext{State: store, Logger: memory.NewLogger("")})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if a.Status().RunCount != 1 {
		t.Errorf("expected RunCount=1, got %d", a.Status().RunCount)
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
