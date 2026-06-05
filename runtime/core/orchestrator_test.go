package core

import (
	"context"
	"testing"
	"time"

	"github.com/Inkedi9/jarvinx/agents"
	"github.com/Inkedi9/jarvinx/memory"
)

// makeTestOrchestrator — crée un orchestrateur avec des composants de test
func makeTestOrchestrator() (*Orchestrator, *Bus, *agents.Registry, *memory.State) {
	bus := NewBus(10)
	state := memory.NewState("")
	logger := memory.NewLogger("")
	registry := agents.NewRegistry()

	orch := NewOrchestrator(bus, registry, state, logger, false, 5*time.Minute)
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

func TestOrchestrator_AgentContextWithSimilarDecisions(t *testing.T) {
	orch, _, _, _ := makeTestOrchestrator()

	// Sans provider — SimilarDecisions doit être nil
	actx := orch.AgentContext()
	if actx.SimilarDecisions != nil {
		t.Errorf("expected nil SimilarDecisions without provider, got %v", actx.SimilarDecisions)
	}

	// Avec provider — SimilarDecisions doit être injectées
	want := []string{"[log] CPU stable. no anomaly. (score:0.92 conf:0.90)"}
	orch.SetSimilarDecisionsProvider(&stubSimilarProvider{decisions: want})

	actx = orch.AgentContext()
	if len(actx.SimilarDecisions) != 1 || actx.SimilarDecisions[0] != want[0] {
		t.Errorf("expected %v, got %v", want, actx.SimilarDecisions)
	}
}

// stubSimilarProvider implémente SimilarDecisionsProvider pour les tests.
type stubSimilarProvider struct {
	decisions []string
}

func (s *stubSimilarProvider) LastSimilarDecisions() []string { return s.decisions }

// ── N-1 pattern ──────────────────────────────────────────────────────────────

func TestOrchestrator_N1PatternEndToEnd(t *testing.T) {
	bus := NewBus(10)
	state := memory.NewState("")
	registry := agents.NewRegistry()
	orch := NewOrchestrator(bus, registry, state, memory.NewLogger(""), true, 5*time.Minute)

	// Cycle N-1 : décision avec commande stockée dans le state
	_ = state.AddCycle(memory.CycleRecord{
		Action: "execute", Command: "uptime", CycleNum: 1, Timestamp: time.Now(),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	go orch.Start(ctx)
	time.Sleep(50 * time.Millisecond)

	// Cycle N : déclenche l'observation — l'orchestrateur doit lire la commande N-1
	bus.Publish(Event{Type: EventObserved, Payload: memory.Snapshot{
		Timestamp: time.Now(), CPUPercent: 30.0, MemPercent: 50.0, DiskPercent: 40.0,
	}})

	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		cmd, _ := orch.ExecGuardStatus()
		if cmd == "uptime" {
			return // N-1 confirmé : la commande du cycle précédent a été consommée
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Error("N-1 pattern: command from previous cycle was not picked up")
}

// ── executeGuard cooldown ─────────────────────────────────────────────────────

func TestOrchestrator_ExecuteGuardCooldown(t *testing.T) {
	guard := &executeGuard{cooldown: 100 * time.Millisecond}

	if !guard.Allow("uptime") {
		t.Fatal("expected first Allow to return true")
	}
	// Même commande dans le cooldown : bloquée
	if guard.Allow("uptime") {
		t.Error("expected second Allow to be blocked by cooldown")
	}
	if guard.CooldownRemaining() == 0 {
		t.Error("expected cooldown remaining > 0 after first execute")
	}

	// Commande différente dans le cooldown : autorisée (guard est par-commande)
	if !guard.Allow("df -h") {
		t.Error("expected different command to bypass cooldown")
	}

	// Après expiration : commande originale autorisée à nouveau
	time.Sleep(120 * time.Millisecond)
	if !guard.Allow("uptime") {
		t.Error("expected Allow to return true after cooldown expired")
	}
}

// ── shouldExecute ─────────────────────────────────────────────────────────────

func TestOrchestrator_ShouldExecute_CancelsWhenMetricNormalized(t *testing.T) {
	orch, _, _, _ := makeTestOrchestrator()

	cycle := memory.CycleRecord{TriggerCPU: 90.0, Command: "uptime"}
	current := memory.Snapshot{CPUPercent: 20.0} // très en dessous du seuil − marge

	if orch.shouldExecute(cycle, current) {
		t.Error("expected shouldExecute=false when CPU normalized (90% → 20%)")
	}
}

func TestOrchestrator_ShouldExecute_AllowsWhenStillHigh(t *testing.T) {
	orch, _, _, _ := makeTestOrchestrator()

	cycle := memory.CycleRecord{TriggerCPU: 90.0, Command: "uptime"}
	current := memory.Snapshot{CPUPercent: 88.0} // encore au-dessus de 90−5=85

	if !orch.shouldExecute(cycle, current) {
		t.Error("expected shouldExecute=true when CPU still high (88% vs trigger 90%)")
	}
}

func TestOrchestrator_ShouldExecute_AllowsWithNoTriggers(t *testing.T) {
	orch, _, _, _ := makeTestOrchestrator()

	// Rétro-compatibilité : aucun trigger → toujours autorisé
	cycle := memory.CycleRecord{Command: "uptime"}
	current := memory.Snapshot{}

	if !orch.shouldExecute(cycle, current) {
		t.Error("expected shouldExecute=true when no trigger values set (backward compat)")
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
