package agents

import (
	"context"
	"testing"
	"time"

	"github.com/Inkedi9/jarvinx/memory"
)

func TestQdrantAgent_SkipsWhenNoDecision(t *testing.T) {
	a := NewQdrantAgent("http://localhost:6333", "http://localhost:11434")
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

func TestQdrantAgent_LogsTextWhenDecisionExists(t *testing.T) {
	a := NewQdrantAgent("http://localhost:6333", "http://localhost:11434")
	store := memory.NewState("state_qdrant_test2.json")

	_ = store.AddCycle(memory.CycleRecord{
		Action:    "log",
		Analysis:  "CPU stable",
		Reason:    "no anomaly detected",
		CycleNum:  1,
		Timestamp: time.Now(),
	})

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

func TestNewQdrantAgent_DefaultSchedule(t *testing.T) {
	a := NewQdrantAgent("http://localhost:6333", "http://localhost:11434")
	if a.Schedule() != 15*time.Second {
		t.Errorf("expected 15s schedule, got %v", a.Schedule())
	}
	if a.Name() != "qdrant" {
		t.Errorf("expected name 'qdrant', got '%s'", a.Name())
	}
}
