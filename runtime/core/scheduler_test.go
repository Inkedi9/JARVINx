package core

import (
	"context"
	"testing"
	"time"
)

func TestScheduler_StartsAndStops(t *testing.T) {
	bus := NewBus(10)
	s := NewScheduler(2*time.Second, bus)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		s.Start(ctx)
		close(done)
	}()

	// Laisse le scheduler démarrer
	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// OK
	case <-time.After(2 * time.Second):
		t.Error("scheduler goroutine did not stop after context cancellation")
	}
}

func TestScheduler_PublishesEvents(t *testing.T) {
	bus := NewBus(10)
	s := NewScheduler(500*time.Millisecond, bus)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go s.Start(ctx)

	// Attend un événement sur le bus
	events := bus.Subscribe("test-scheduler")
	select {
	case e := <-events:
		if e.Type != EventObserved && e.Type != EventError {
			t.Errorf("unexpected event type: %s", e.Type)
		}
	case <-ctx.Done():
		t.Error("no event received within timeout")
	}
}

func TestScheduler_SetInterval(t *testing.T) {
	bus := NewBus(10)
	s := NewScheduler(1*time.Second, bus)

	s.SetInterval(30 * time.Second)

	if s.getInterval() != 30*time.Second {
		t.Errorf("expected 30s interval, got %v", s.getInterval())
	}
}

func TestScheduler_ContextCancellation(t *testing.T) {
	bus := NewBus(10)
	s := NewScheduler(1*time.Minute, bus) // intervalle long — ne tick pas

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		s.Start(ctx)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// OK
	case <-time.After(500 * time.Millisecond):
		t.Error("scheduler should stop on context cancellation")
	}
}
