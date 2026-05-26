package llm

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// httptest.NewServer crée un vrai serveur HTTP local pour les tests
// Pas besoin de mocker — on teste contre un vrai serveur de test

func makeOllamaServer(models []string, statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if statusCode != http.StatusOK {
			w.WriteHeader(statusCode)
			return
		}

		type modelEntry struct {
			Name string `json:"name"`
		}
		type response struct {
			Models []modelEntry `json:"models"`
		}

		entries := make([]modelEntry, 0, len(models))
		for _, m := range models {
			entries = append(entries, modelEntry{Name: m})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response{Models: entries})
	}))
}

func TestCheckOllama_Online(t *testing.T) {
	srv := makeOllamaServer([]string{"llama3.1:8b", "qwen2.5:7b"}, http.StatusOK)
	defer srv.Close()

	status := CheckOllama(srv.URL, "llama3.1:8b")

	if !status.Online {
		t.Fatalf("expected online, got error: %s", status.Error)
	}
	if status.Error != "" {
		t.Errorf("expected no error, got: %s", status.Error)
	}
	if len(status.Models) != 2 {
		t.Errorf("expected 2 models, got %d", len(status.Models))
	}
}

func TestCheckOllama_Offline(t *testing.T) {
	// URL qui ne répond pas
	status := CheckOllama("http://localhost:19999", "llama3.1:8b")

	if status.Online {
		t.Fatal("expected offline for unreachable URL")
	}
	if status.Error == "" {
		t.Error("expected error message when offline")
	}
}

func TestCheckOllama_ModelMissing(t *testing.T) {
	srv := makeOllamaServer([]string{"qwen2.5:7b", "mistral:7b"}, http.StatusOK)
	defer srv.Close()

	status := CheckOllama(srv.URL, "llama3.1:8b")

	if !status.Online {
		t.Fatal("server is online, expected Online=true")
	}
	if status.Error == "" {
		t.Error("expected error when requested model is missing")
	}
}

func TestCheckOllama_ServerError(t *testing.T) {
	srv := makeOllamaServer(nil, http.StatusInternalServerError)
	defer srv.Close()

	status := CheckOllama(srv.URL, "llama3.1:8b")

	if status.Online {
		t.Fatal("expected offline for 500 response")
	}
}

func TestCheckOllama_NoModels(t *testing.T) {
	srv := makeOllamaServer([]string{}, http.StatusOK)
	defer srv.Close()

	status := CheckOllama(srv.URL, "llama3.1:8b")

	if !status.Online {
		t.Fatal("server is online, expected Online=true")
	}
	// Aucun modèle installé = erreur sur le modèle demandé
	if status.Error == "" {
		t.Error("expected error when no models installed")
	}
}

func TestHealthStatus_Display_Online(t *testing.T) {
	h := HealthStatus{
		Online: true,
		Models: []string{"llama3.1:8b"},
	}
	// Juste vérifier que Display ne panic pas
	h.Display("llama3.1:8b")
}

func TestHealthStatus_Display_Offline(t *testing.T) {
	h := HealthStatus{
		Online: false,
		Error:  "connexion refusée",
	}
	h.Display("llama3.1:8b")
}
