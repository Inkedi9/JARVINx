package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Inkedi9/jarvinx/agents"
	"github.com/Inkedi9/jarvinx/config"
	"github.com/Inkedi9/jarvinx/memory"
)

func makeTestServer(origins []string) *Server {
	cfg := config.Default()
	cfg.AllowedOrigins = origins

	originMap := make(map[string]bool)
	for _, o := range origins {
		originMap[o] = true
	}

	return &Server{
		cfg:            cfg,
		state:          memory.NewState(""),
		registry:       agents.NewRegistry(),
		allowedOrigins: originMap,
	}
}

func TestCORS_AllowedOrigin(t *testing.T) {
	srv := makeTestServer([]string{"http://localhost:3000"})

	handler := srv.corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/status", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	got := w.Header().Get("Access-Control-Allow-Origin")
	if got != "http://localhost:3000" {
		t.Errorf("expected ACAO header 'http://localhost:3000', got '%s'", got)
	}
}

func TestCORS_BlockedOrigin(t *testing.T) {
	srv := makeTestServer([]string{"http://localhost:3000"})

	handler := srv.corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/status", nil)
	req.Header.Set("Origin", "http://evil.com")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	got := w.Header().Get("Access-Control-Allow-Origin")
	if got != "" {
		t.Errorf("expected no ACAO header for blocked origin, got '%s'", got)
	}
}

func TestCORS_PreflightAllowed(t *testing.T) {
	srv := makeTestServer([]string{"http://localhost:3000"})

	handler := srv.corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("OPTIONS", "/api/status", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204 for allowed preflight, got %d", w.Code)
	}
}

func TestCORS_PreflightBlocked(t *testing.T) {
	srv := makeTestServer([]string{"http://localhost:3000"})

	handler := srv.corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("OPTIONS", "/api/status", nil)
	req.Header.Set("Origin", "http://evil.com")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for blocked preflight, got %d", w.Code)
	}
}

func TestCORS_NoOriginHeader(t *testing.T) {
	srv := makeTestServer([]string{"http://localhost:3000"})

	handler := srv.corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Requête sans header Origin — curl direct, pas un browser
	req := httptest.NewRequest("GET", "/api/status", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Doit passer — pas de CORS sur requêtes sans Origin
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for no-origin request, got %d", w.Code)
	}
}

func TestConfig_AllowedOriginsValidation(t *testing.T) {
	cfg := config.Default()
	cfg.AllowedOrigins = []string{"not-a-url"}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for invalid origin format")
	}
}

func TestConfig_EmptyAllowedOrigins(t *testing.T) {
	cfg := config.Default()
	cfg.AllowedOrigins = []string{}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for empty AllowedOrigins")
	}
}
