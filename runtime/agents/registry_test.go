package agents

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Inkedi9/jarvinx/memory"
)

// mockAgent — un agent de test qu'on contrôle complètement
type mockAgent struct {
	BaseAgent
	runCount    atomic.Int32
	shouldPanic bool
}

func newMockAgent(name string, schedule time.Duration) *mockAgent {
	return &mockAgent{
		BaseAgent: NewBaseAgent(name, schedule),
	}
}

func (m *mockAgent) Run(ctx context.Context, actx AgentContext) error {
	if m.shouldPanic {
		panic("simulated panic in agent")
	}
	m.runCount.Add(1)
	m.recordSuccess()
	return nil
}

// ── Register ─────────────────────────────────────────────────────────────────

func TestRegistry_Register(t *testing.T) {
	r := NewRegistry()
	a := newMockAgent("test-agent", 1*time.Second)

	r.Register(a)

	_, found := r.Get("test-agent")
	if !found {
		t.Error("expected to find registered agent")
	}
}

func TestRegistry_GetUnknown(t *testing.T) {
	r := NewRegistry()

	_, found := r.Get("nonexistent")
	if found {
		t.Error("expected not to find unregistered agent")
	}
}

func TestRegistry_MultipleAgents(t *testing.T) {
	r := NewRegistry()
	r.Register(newMockAgent("agent-a", 1*time.Second))
	r.Register(newMockAgent("agent-b", 1*time.Second))
	r.Register(newMockAgent("agent-c", 1*time.Second))

	statuses := r.Statuses()
	if len(statuses) != 3 {
		t.Errorf("expected 3 agents, got %d", len(statuses))
	}
}

// ── Enable / Disable ─────────────────────────────────────────────────────────

func TestRegistry_EnableDisable(t *testing.T) {
	r := NewRegistry()
	a := newMockAgent("toggleable", 1*time.Second)
	r.Register(a)

	// Désactiver
	r.Disable("toggleable")
	if a.IsEnabled() {
		t.Error("expected agent to be disabled")
	}

	// Réactiver
	r.Enable("toggleable")
	if !a.IsEnabled() {
		t.Error("expected agent to be enabled")
	}
}

func TestRegistry_DisableUnknownAgent(t *testing.T) {
	r := NewRegistry()

	result := r.Disable("ghost")
	if result {
		t.Error("disabling unknown agent should return false")
	}
}

func TestRegistry_DisabledAgentSkipped(t *testing.T) {
	r := NewRegistry()
	a := newMockAgent("skippable", 50*time.Millisecond)
	r.Register(a)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// Désactivé avant le lancement
	r.Disable("skippable")

	makeCtx := func() AgentContext {
		return AgentContext{
			Snapshot: memory.Snapshot{},
			State:    memory.NewState(""),
			Logger:   memory.NewLogger(""),
		}
	}

	r.Start(ctx, makeCtx)
	<-ctx.Done()
	time.Sleep(20 * time.Millisecond)

	if a.runCount.Load() > 0 {
		t.Errorf("disabled agent should not run, got %d runs", a.runCount.Load())
	}
}

// ── Exécution ────────────────────────────────────────────────────────────────

func TestRegistry_AgentRuns(t *testing.T) {
	r := NewRegistry()
	a := newMockAgent("runner", 50*time.Millisecond)
	r.Register(a)

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	makeCtx := func() AgentContext {
		return AgentContext{
			Snapshot: memory.Snapshot{},
			State:    memory.NewState(""),
			Logger:   memory.NewLogger(""),
		}
	}

	r.Start(ctx, makeCtx)
	<-ctx.Done()
	time.Sleep(20 * time.Millisecond)

	if a.runCount.Load() == 0 {
		t.Error("expected agent to run at least once")
	}
}

func TestRegistry_PanicIsolation(t *testing.T) {
	r := NewRegistry()

	// Agent qui panic
	bad := newMockAgent("bad-agent", 50*time.Millisecond)
	bad.shouldPanic = true

	// Agent sain
	good := newMockAgent("good-agent", 50*time.Millisecond)

	r.Register(bad)
	r.Register(good)

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	makeCtx := func() AgentContext {
		return AgentContext{
			Snapshot: memory.Snapshot{},
			State:    memory.NewState(""),
			Logger:   memory.NewLogger(""),
		}
	}

	// Ne doit pas crasher malgré le panic dans bad-agent
	r.Start(ctx, makeCtx)
	<-ctx.Done()
	time.Sleep(20 * time.Millisecond)

	// Le bon agent doit avoir tourné malgré le panic du mauvais
	if good.runCount.Load() == 0 {
		t.Error("good agent should run despite panic in bad agent")
	}
}

// ── Status ───────────────────────────────────────────────────────────────────

func TestRegistry_StatusReflectsRunCount(t *testing.T) {
	r := NewRegistry()
	a := newMockAgent("tracked", 50*time.Millisecond)
	r.Register(a)

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	makeCtx := func() AgentContext {
		return AgentContext{
			Snapshot: memory.Snapshot{},
			State:    memory.NewState(""),
			Logger:   memory.NewLogger(""),
		}
	}

	r.Start(ctx, makeCtx)
	<-ctx.Done()
	time.Sleep(20 * time.Millisecond)

	status := a.Status()
	if status.RunCount == 0 {
		t.Error("expected RunCount > 0 after execution")
	}
	if status.LastRun.IsZero() {
		t.Error("expected LastRun to be set after execution")
	}
}
