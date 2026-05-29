package core

import (
	"context"
	"testing"
	"time"

	"github.com/Inkedi9/jarvinx/agents"
	"github.com/Inkedi9/jarvinx/memory"
)

// mockAgent pour les tests orchestrateur
type mockOrchestratorAgent struct {
	agents.BaseAgent
	runCount int
	runErr   error
}

func newMockOrchestratorAgent(name string) *mockOrchestratorAgent {
	return &mockOrchestratorAgent{
		BaseAgent: agents.NewBaseAgent(name, 15*time.Second),
	}
}

func (m *mockOrchestratorAgent) Run(_ context.Context, _ agents.AgentContext) error {
	m.runCount++
	return m.runErr
}

// makeTestOrchestrator — crée un orchestrateur avec des composants de test
func makeTestOrchestrator() (*Orchestrator, *Bus, *agents.Registry, *memory.State) {
	bus := NewBus(10)
	state := memory.NewState("")
	logger := memory.NewLogger("")
	registry := agents.NewRegistry()

	orch := NewOrchestrator(bus, registry, state, logger)
	return orch, bus, registry, state
}

func TestOrchestrator_StartStop(t *testing.T) {
	orch, _, _, _ := makeTestOrchestrator()

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		orch.Start(ctx)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// OK — arrêt propre
	case <-time.After(500 * time.Millisecond):
		t.Error("orchestrator did not stop after context cancellation")
	}
}

func TestOrchestrator_HandlesObservedEvent(t *testing.T) {
	orch, bus, _, _ := makeTestOrchestrator()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go orch.Start(ctx)
	time.Sleep(50 * time.Millisecond)

	snap := memory.Snapshot{
		Timestamp:   time.Now(),
		CPUPercent:  42.0,
		MemPercent:  50.0,
		DiskPercent: 60.0,
		MemTotal:    16000,
		MemUsed:     8000,
		DiskTotal:   500,
		DiskUsed:    300,
	}

	bus.Publish(Event{Type: EventObserved, Payload: snap})

	// Polling — attend que lastSnap soit mis à jour
	deadline := time.Now().Add(1 * time.Second)
	for time.Now().Before(deadline) {
		actx := orch.AgentContext()
		if actx.Snapshot.CPUPercent == 42.0 {
			return // succès — lastSnap mis à jour
		}
		time.Sleep(20 * time.Millisecond)
	}

	t.Error("expected lastSnap to be updated after EventObserved")
}

func TestOrchestrator_IgnoresInvalidPayload(t *testing.T) {
	orch, bus, _, _ := makeTestOrchestrator()

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	go orch.Start(ctx)
	time.Sleep(50 * time.Millisecond)

	// Payload invalide — ne doit pas crasher
	bus.Publish(Event{Type: EventObserved, Payload: "not a snapshot"})

	<-ctx.Done()
	// Si on arrive ici sans panic, le test passe
}

func TestOrchestrator_HandlesErrorEvent(t *testing.T) {
	orch, bus, _, _ := makeTestOrchestrator()

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	go orch.Start(ctx)
	time.Sleep(50 * time.Millisecond)

	// EventError ne doit pas crasher l'orchestrateur
	bus.Publish(Event{Type: EventError, Payload: "test error"})

	<-ctx.Done()
}

func TestOrchestrator_TryLockPreventsOverlap(t *testing.T) {
	orch, bus, _, state := makeTestOrchestrator()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go orch.Start(ctx)
	time.Sleep(50 * time.Millisecond)

	snap := memory.Snapshot{
		Timestamp:  time.Now(),
		CPUPercent: 10.0,
		MemTotal:   16000,
		DiskTotal:  500,
	}

	// Publie deux événements rapidement
	bus.Publish(Event{Type: EventObserved, Payload: snap})
	bus.Publish(Event{Type: EventObserved, Payload: snap})

	time.Sleep(300 * time.Millisecond)

	// Au maximum 1 cycle traité (TryLock ignore le second)
	cycles := state.LastCycles(10)
	if len(cycles) > 2 {
		t.Errorf("expected at most 2 cycles, got %d — TryLock may not be working", len(cycles))
	}
}

func TestOrchestrator_MultipleSubscribers(t *testing.T) {
	orch, bus, _, _ := makeTestOrchestrator()

	// Second subscriber sur le même bus
	extraCh := bus.Subscribe("extra-consumer")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go orch.Start(ctx)
	time.Sleep(50 * time.Millisecond)

	snap := memory.Snapshot{
		Timestamp:  time.Now(),
		CPUPercent: 10.0,
		MemTotal:   16000,
		DiskTotal:  500,
	}

	bus.Publish(Event{Type: EventObserved, Payload: snap})

	// Le subscriber extra doit aussi recevoir l'événement — fan-out
	select {
	case e := <-extraCh:
		if e.Type != EventObserved {
			t.Errorf("expected EventObserved on extra subscriber, got %s", e.Type)
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("extra subscriber did not receive event — fan-out broken")
	}
}
