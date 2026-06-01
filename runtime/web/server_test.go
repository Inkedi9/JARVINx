package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Inkedi9/jarvinx/agents"
	"github.com/Inkedi9/jarvinx/config"
	"github.com/Inkedi9/jarvinx/memory"
)

// mockToggleAgent — agent minimal pour tester le toggle
type mockToggleAgent struct {
	name    string
	enabled bool
}

func (m *mockToggleAgent) Name() string                                       { return m.name }
func (m *mockToggleAgent) Schedule() time.Duration                            { return 15 * time.Second }
func (m *mockToggleAgent) Run(_ context.Context, _ agents.AgentContext) error { return nil }
func (m *mockToggleAgent) IsEnabled() bool                                    { return m.enabled }
func (m *mockToggleAgent) Enable()                                            { m.enabled = true }
func (m *mockToggleAgent) Disable()                                           { m.enabled = false }
func (m *mockToggleAgent) Status() agents.AgentStatus {
	return agents.AgentStatus{Name: m.name, Enabled: m.enabled}
}

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
		mainLogger:     nil,
		alertLogger:    nil,
		dailyReporter:  nil, // ← ajoute ça
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

func TestToggle_AgentEnable(t *testing.T) {
	srv := makeTestServer([]string{"http://localhost:3000"})

	// Enregistre un agent de test
	a := &mockToggleAgent{name: "system", enabled: true}
	srv.registry.Register(a)

	// Désactive
	body := strings.NewReader(`{"name":"system"}`)
	req := httptest.NewRequest("POST", "/api/agents/toggle", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleAgentToggle(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}

	var resp ToggleResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("invalid response JSON: %v", err)
	}

	if resp.Enabled {
		t.Error("expected agent to be disabled after toggle")
	}
}

func TestToggle_AgentNotFound(t *testing.T) {
	srv := makeTestServer([]string{"http://localhost:3000"})

	body := strings.NewReader(`{"name":"nonexistent"}`)
	req := httptest.NewRequest("POST", "/api/agents/toggle", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleAgentToggle(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestToggle_MethodNotAllowed(t *testing.T) {
	srv := makeTestServer([]string{"http://localhost:3000"})

	req := httptest.NewRequest("GET", "/api/agents/toggle", nil)
	w := httptest.NewRecorder()

	srv.handleAgentToggle(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestToggle_MissingName(t *testing.T) {
	srv := makeTestServer([]string{"http://localhost:3000"})

	body := strings.NewReader(`{}`)
	req := httptest.NewRequest("POST", "/api/agents/toggle", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleAgentToggle(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestToggle_InvalidJSON(t *testing.T) {
	srv := makeTestServer([]string{"http://localhost:3000"})

	body := strings.NewReader(`not json`)
	req := httptest.NewRequest("POST", "/api/agents/toggle", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleAgentToggle(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestStatus_DryRunField(t *testing.T) {
	srv := makeTestServer([]string{"http://localhost:3000"})
	srv.cfg.DryRun = true

	req := httptest.NewRequest("GET", "/api/status", nil)
	w := httptest.NewRecorder()

	srv.handleStatus(w, req)

	var resp StatusResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("invalid response: %v", err)
	}

	if !resp.DryRun {
		t.Error("expected dry_run=true in status response")
	}
}

func TestStatus_DryRunFalseByDefault(t *testing.T) {
	srv := makeTestServer([]string{"http://localhost:3000"})

	req := httptest.NewRequest("GET", "/api/status", nil)
	w := httptest.NewRecorder()

	srv.handleStatus(w, req)

	var resp StatusResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("invalid response: %v", err)
	}

	if resp.DryRun {
		t.Error("expected dry_run=false by default")
	}
}

func TestLogsStatus_Endpoint(t *testing.T) {
	srv := makeTestServer([]string{"http://localhost:3000"})
	srv.mainLogger = memory.NewLogger("")
	srv.alertLogger = memory.NewLogger("")

	req := httptest.NewRequest("GET", "/api/logs/status", nil)
	w := httptest.NewRecorder()

	srv.handleLogsStatus(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp LogsStatusResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("invalid response JSON: %v", err)
	}
}

func TestLogsStatus_NilLoggers(t *testing.T) {
	srv := makeTestServer([]string{"http://localhost:3000"})
	// mainLogger et alertLogger nil par défaut dans makeTestServer

	req := httptest.NewRequest("GET", "/api/logs/status", nil)
	w := httptest.NewRecorder()

	srv.handleLogsStatus(w, req)

	// Ne doit pas crasher avec des loggers nil
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleFile_NoAgent(t *testing.T) {
	srv := makeTestServer([]string{"http://localhost:3000"})

	req := httptest.NewRequest("GET", "/api/file", nil)
	w := httptest.NewRecorder()

	srv.handleFile(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp FileAgentResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if resp.Enabled {
		t.Error("expected disabled when no file agent registered")
	}
}

func TestHandleDailyReport_Disabled(t *testing.T) {
	srv := makeTestServer([]string{"http://localhost:3000"})
	srv.cfg.DailyReportEnabled = false

	req := httptest.NewRequest("GET", "/api/daily-report", nil)
	w := httptest.NewRecorder()

	srv.handleDailyReport(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp DailyReportResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if resp.Enabled {
		t.Error("expected enabled=false")
	}
}

func TestHandleDailyReportSend_MethodNotAllowed(t *testing.T) {
	srv := makeTestServer([]string{"http://localhost:3000"})

	req := httptest.NewRequest("GET", "/api/daily-report/send", nil)
	w := httptest.NewRecorder()

	srv.handleDailyReportSend(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandleDailyReportSend_NoReporter(t *testing.T) {
	srv := makeTestServer([]string{"http://localhost:3000"})
	srv.dailyReporter = nil

	req := httptest.NewRequest("POST", "/api/daily-report/send", nil)
	w := httptest.NewRecorder()

	srv.handleDailyReportSend(w, req)

	var resp SendReportResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Sent {
		t.Error("expected sent=false when no reporter configured")
	}
}

func TestHandleLLMContext_Empty(t *testing.T) {
	srv := makeTestServer([]string{"http://localhost:3000"})

	req := httptest.NewRequest("GET", "/api/llm-context", nil)
	w := httptest.NewRecorder()

	srv.handleLLMContext(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp LLMContextResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if resp.CycleCount != 0 {
		t.Errorf("expected 0 cycles for empty state, got %d", resp.CycleCount)
	}
}
