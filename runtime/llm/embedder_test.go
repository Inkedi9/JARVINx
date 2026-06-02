package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockEmbedServer(embedding []float32, status int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/embeddings" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(status)
		if status == http.StatusOK {
			_ = json.NewEncoder(w).Encode(embedResponse{Embedding: embedding})
		}
	}))
}

func TestOllamaEmbedder_Success(t *testing.T) {
	want := []float32{0.1, 0.2, 0.3, 0.4}
	srv := mockEmbedServer(want, http.StatusOK)
	defer srv.Close()

	e := NewOllamaEmbedder(srv.URL, "nomic-embed-text")
	got, err := e.Embed(context.Background(), "test text")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("expected %d dims, got %d", len(want), len(got))
	}
	for i, v := range want {
		if got[i] != v {
			t.Errorf("dim[%d]: expected %f, got %f", i, v, got[i])
		}
	}
}

func TestOllamaEmbedder_ServerError(t *testing.T) {
	srv := mockEmbedServer(nil, http.StatusInternalServerError)
	defer srv.Close()

	e := NewOllamaEmbedder(srv.URL, "nomic-embed-text")
	_, err := e.Embed(context.Background(), "test text")
	if err == nil {
		t.Fatal("expected error on server 500")
	}
}

func TestOllamaEmbedder_EmptyEmbedding(t *testing.T) {
	srv := mockEmbedServer([]float32{}, http.StatusOK)
	defer srv.Close()

	e := NewOllamaEmbedder(srv.URL, "nomic-embed-text")
	_, err := e.Embed(context.Background(), "test text")
	if err == nil {
		t.Fatal("expected error on empty embedding")
	}
}

func TestOllamaEmbedder_CircuitOpenAfterFailures(t *testing.T) {
	srv := mockEmbedServer(nil, http.StatusInternalServerError)
	defer srv.Close()

	e := NewOllamaEmbedder(srv.URL, "nomic-embed-text")
	// 3 échecs pour ouvrir le circuit
	for range 3 {
		_, _ = e.Embed(context.Background(), "x")
	}
	_, err := e.Embed(context.Background(), "x")
	if err == nil {
		t.Fatal("expected circuit open error after 3 failures")
	}
	if err != ErrCircuitOpen {
		t.Errorf("expected ErrCircuitOpen, got: %v", err)
	}
}
