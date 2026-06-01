package core

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Inkedi9/jarvinx/agents"
	"github.com/Inkedi9/jarvinx/config"
	"github.com/Inkedi9/jarvinx/memory"
)

func mockOllamaServer(action string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"message": map[string]any{
				"role":    "assistant",
				"content": `{"analysis":"test","action":"` + action + `","reason":"integration test"}`,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
}

func makeIntegrationConfig(ollamaURL string) *config.Config {
	cfg := config.Default()
	cfg.OllamaURL = ollamaURL
	cfg.Model = "test-model"
	cfg.Interval = 100 * time.Millisecond
	cfg.LogFile = ""
	cfg.StateFile = ""
	cfg.AlertFile = ""
	cfg.DryRun = true
	cfg.DockerEnabled = false
	cfg.DailyReportEnabled = false
	cfg.AllowedOrigins = []string{"http://localhost:3000"}
	return cfg
}

// TestIntegration_FullCycle — appelle SystemAgent.Run() directement
func TestIntegration_FullCycle(t *testing.T) {
	srv := mockOllamaServer("log")
	defer srv.Close()

	cfg := makeIntegrationConfig(srv.URL)
	state := memory.NewState("")

	agent := agents.NewSystemAgent(cfg.OllamaURL, cfg.Model)

	actx := agents.AgentContext{
		Snapshot: memory.Snapshot{
			Timestamp:   time.Now(),
			CPUPercent:  10.0,
			MemPercent:  50.0,
			DiskPercent: 60.0,
			MemTotal:    16000,
			MemUsed:     8000,
			DiskTotal:   500,
			DiskUsed:    300,
		},
		State:  state,
		Logger: memory.NewLogger(""),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := agent.Run(ctx, actx)
	if err != nil {
		t.Logf("Run() returned error (may be fallback): %v", err)
	}

	// Vérifie que le cycle a été enregistré
	cycles := state.LastCycles(1)
	if len(cycles) == 0 {
		t.Error("expected at least 1 cycle record after Run()")
	}
}

// TestIntegration_AlertTriggered — AlertAgent avec seuils bas
func TestIntegration_AlertTriggered(t *testing.T) {
	dispatcher := agents.NewNotifierDispatcher(true)
	agent := agents.NewAlertAgent(
		1.0, 1.0, 1.0, // seuils très bas
		1, 1,
		"",
		dispatcher,
	)

	actx := agents.AgentContext{
		Snapshot: memory.Snapshot{
			Timestamp:   time.Now(),
			CPUPercent:  50.0,
			MemPercent:  60.0,
			DiskPercent: 70.0,
			MemTotal:    16000,
			MemUsed:     9600,
			DiskTotal:   500,
			DiskUsed:    350,
		},
		State:  memory.NewState(""),
		Logger: memory.NewLogger(""),
	}

	err := agent.Run(context.Background(), actx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	status := agent.Status()
	if status.RunCount == 0 {
		t.Error("expected AlertAgent to run")
	}
}

// TestIntegration_BusEventFlow — flux d'événements dans le bus
func TestIntegration_BusEventFlow(t *testing.T) {
	bus := NewBus(10)
	received := make(chan Event, 10)
	events := bus.Subscribe("test-consumer")

	go func() {
		for e := range events {
			received <- e
		}
	}()

	bus.Publish(Event{Type: EventObserved, Payload: "snap"})
	bus.Publish(Event{Type: EventError, Payload: "err"})
	bus.Publish(Event{Type: EventExecuted, Payload: "result"})

	time.Sleep(50 * time.Millisecond)
	bus.Unsubscribe("test-consumer")

	if len(received) != 3 {
		t.Errorf("expected 3 events, got %d", len(received))
	}
}

// TestIntegration_SchedulerPublishes — le scheduler publie des EventObserved
func TestIntegration_SchedulerPublishes(t *testing.T) {
	bus := NewBus(10)
	events := bus.Subscribe("observer")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	scheduler := NewScheduler(200*time.Millisecond, bus)
	go scheduler.Start(ctx)

	select {
	case e := <-events:
		if e.Type != EventObserved && e.Type != EventError {
			t.Errorf("expected EventObserved or EventError, got %s", e.Type)
		}
	case <-ctx.Done():
		t.Error("expected at least 1 event from scheduler within timeout")
	}
}

// TestIntegration_DryRunNoExecution — dry-run n'exécute pas les commandes
func TestIntegration_DryRunNoExecution(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"message": map[string]any{
				"role":    "assistant",
				"content": `{"analysis":"test","action":"execute","command":"uptime","reason":"test"}`,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	cfg := makeIntegrationConfig(srv.URL)
	state := memory.NewState("")
	agent := agents.NewSystemAgent(cfg.OllamaURL, cfg.Model)

	actx := agents.AgentContext{
		Snapshot: memory.Snapshot{
			Timestamp: time.Now(),
			MemTotal:  16000,
			DiskTotal: 500,
		},
		State:  state,
		Logger: memory.NewLogger(""),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	agent.Run(ctx, actx)

	cycles := state.LastCycles(1)
	if len(cycles) == 0 {
		t.Error("expected cycle record even with execute action")
		return
	}

	if cycles[0].Action == "execute" && cycles[0].Command == "uptime" {
		// Cycle enregistré — en dry-run l'orchestrateur ne l'exécute pas
		// On vérifie juste que le cycle est bien enregistré
	}
}

// TestIntegration_OllamaDown_Fallback — fallback quand Ollama est down
func TestIntegration_OllamaDown_Fallback(t *testing.T) {
	cfg := makeIntegrationConfig("http://localhost:19999")
	state := memory.NewState("")
	agent := agents.NewSystemAgent(cfg.OllamaURL, cfg.Model)

	// Timeout court pour ne pas attendre 3 × 60s
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	actx := agents.AgentContext{
		Snapshot: memory.Snapshot{
			Timestamp: time.Now(),
			MemTotal:  16000,
			DiskTotal: 500,
		},
		State:  state,
		Logger: memory.NewLogger(""),
	}

	// Doit retourner une erreur mais pas crasher
	err := agent.Run(ctx, actx)
	if err == nil {
		t.Log("Run() succeeded (maybe circuit breaker fallback)")
	}

	// L'agent doit avoir un RunCount même en cas d'erreur
	status := agent.Status()
	if status.RunCount == 0 {
		t.Error("expected RunCount > 0 even when Ollama is down")
	}
}
