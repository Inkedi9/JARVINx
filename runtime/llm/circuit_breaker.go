package llm

import (
	"fmt"
	"sync"
	"time"
)

type CircuitState int

const (
	StateClosed   CircuitState = iota // normal — appels autorisés
	StateOpen                         // panne — appels bloqués
	StateHalfOpen                     // test — un appel autorisé
)

func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// ErrCircuitOpen est retourné quand le circuit est ouvert
var ErrCircuitOpen = fmt.Errorf("circuit breaker open — Ollama unreachable")

type CircuitBreaker struct {
	mu sync.Mutex

	// Config
	maxFailures  int           // échecs avant ouverture
	resetTimeout time.Duration // durée avant passage en half-open

	// State
	state       CircuitState
	failures    int
	lastFailure time.Time
	successes   int // succès consécutifs en half-open
}

func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        StateClosed,
	}
}

// DefaultCircuitBreaker — 3 échecs consécutifs, reset après 30s
func DefaultCircuitBreaker() *CircuitBreaker {
	return NewCircuitBreaker(3, 30*time.Second)
}

// Allow vérifie si un appel est autorisé
func (cb *CircuitBreaker) Allow() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return nil

	case StateOpen:
		// Vérifie si le timeout de reset est écoulé
		if time.Since(cb.lastFailure) >= cb.resetTimeout {
			cb.state = StateHalfOpen
			cb.successes = 0
			return nil // autorise un appel test
		}
		return ErrCircuitOpen

	case StateHalfOpen:
		return nil // autorise l'appel test
	}

	return nil
}

// RecordSuccess enregistre un succès
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures = 0

	if cb.state == StateHalfOpen {
		cb.successes++
		if cb.successes >= 1 { // un succès suffit pour refermer
			cb.state = StateClosed
		}
	}
}

// RecordFailure enregistre un échec
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailure = time.Now()

	switch cb.state {
	case StateClosed:
		if cb.failures >= cb.maxFailures {
			cb.state = StateOpen
		}
	case StateHalfOpen:
		// Échec en half-open → retour open immédiat
		cb.state = StateOpen
		cb.successes = 0
	}
}

// State retourne l'état courant
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// Stats retourne les stats pour monitoring
type CircuitStats struct {
	State       string `json:"state"`
	Failures    int    `json:"failures"`
	MaxFailures int    `json:"max_failures"`
}

func (cb *CircuitBreaker) Stats() CircuitStats {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return CircuitStats{
		State:       cb.state.String(),
		Failures:    cb.failures,
		MaxFailures: cb.maxFailures,
	}
}
