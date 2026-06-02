package llm

import (
	"testing"
	"time"
)

func TestCircuitBreaker_InitialState(t *testing.T) {
	cb := NewCircuitBreaker(3, 30*time.Second)
	if cb.State() != StateClosed {
		t.Errorf("expected initial state Closed, got %s", cb.State())
	}
}

func TestCircuitBreaker_OpenAfterMaxFailures(t *testing.T) {
	cb := NewCircuitBreaker(3, 30*time.Second)

	cb.RecordFailure()
	cb.RecordFailure()

	if cb.State() != StateClosed {
		t.Error("should still be closed after 2 failures (max=3)")
	}

	cb.RecordFailure()

	if cb.State() != StateOpen {
		t.Errorf("expected Open after 3 failures, got %s", cb.State())
	}
}

func TestCircuitBreaker_AllowBlocked_WhenOpen(t *testing.T) {
	cb := NewCircuitBreaker(1, 30*time.Second)
	cb.RecordFailure()

	if err := cb.Allow(); err == nil {
		t.Error("expected ErrCircuitOpen when state is Open")
	}
}

func TestCircuitBreaker_AllowPasses_WhenClosed(t *testing.T) {
	cb := NewCircuitBreaker(3, 30*time.Second)

	if err := cb.Allow(); err != nil {
		t.Errorf("expected nil error when Closed, got: %v", err)
	}
}

func TestCircuitBreaker_HalfOpenAfterTimeout(t *testing.T) {
	cb := NewCircuitBreaker(1, 50*time.Millisecond)
	cb.RecordFailure() // → Open

	time.Sleep(60 * time.Millisecond)

	// Allow() doit passer en HalfOpen
	if err := cb.Allow(); err != nil {
		t.Errorf("expected nil after reset timeout, got: %v", err)
	}
	if cb.State() != StateHalfOpen {
		t.Errorf("expected HalfOpen after timeout, got %s", cb.State())
	}
}

func TestCircuitBreaker_ClosedAfterSuccessInHalfOpen(t *testing.T) {
	cb := NewCircuitBreaker(1, 50*time.Millisecond)
	cb.RecordFailure()
	time.Sleep(60 * time.Millisecond)
	_ = cb.Allow() // → HalfOpen

	cb.RecordSuccess()

	if cb.State() != StateClosed {
		t.Errorf("expected Closed after success in HalfOpen, got %s", cb.State())
	}
}

func TestCircuitBreaker_ReOpenAfterFailureInHalfOpen(t *testing.T) {
	cb := NewCircuitBreaker(1, 50*time.Millisecond)
	cb.RecordFailure()
	time.Sleep(60 * time.Millisecond)
	_ = cb.Allow() // → HalfOpen

	cb.RecordFailure() // échec en half-open → retour Open

	if cb.State() != StateOpen {
		t.Errorf("expected Open after failure in HalfOpen, got %s", cb.State())
	}
}

func TestCircuitBreaker_SuccessResetsFailures(t *testing.T) {
	cb := NewCircuitBreaker(3, 30*time.Second)

	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordSuccess() // reset

	cb.RecordFailure()
	cb.RecordFailure()

	if cb.State() != StateClosed {
		t.Errorf("expected Closed after success reset, got %s", cb.State())
	}
}

func TestCircuitBreaker_Stats(t *testing.T) {
	cb := NewCircuitBreaker(3, 30*time.Second)
	cb.RecordFailure()

	stats := cb.Stats()
	if stats.State != "closed" {
		t.Errorf("expected state 'closed', got '%s'", stats.State)
	}
	if stats.Failures != 1 {
		t.Errorf("expected 1 failure, got %d", stats.Failures)
	}
	if stats.MaxFailures != 3 {
		t.Errorf("expected max 3, got %d", stats.MaxFailures)
	}
}

func TestCircuitBreaker_StateString(t *testing.T) {
	if StateClosed.String() != "closed" {
		t.Errorf("expected 'closed', got '%s'", StateClosed.String())
	}
	if StateOpen.String() != "open" {
		t.Errorf("expected 'open', got '%s'", StateOpen.String())
	}
	if StateHalfOpen.String() != "half-open" {
		t.Errorf("expected 'half-open', got '%s'", StateHalfOpen.String())
	}
}
