package web

import (
	"sync"
	"time"
)

// tokenBucket implements a thread-safe token bucket rate limiter.
type tokenBucket struct {
	mu         sync.Mutex
	tokens     float64
	capacity   float64
	refillRate float64 // tokens per second
	lastRefill time.Time
}

func newTokenBucket(ratePerSec, capacity float64) *tokenBucket {
	return &tokenBucket{
		tokens:     capacity,
		capacity:   capacity,
		refillRate: ratePerSec,
		lastRefill: time.Now(),
	}
}

// Allow returns true and consumes one token if available, false otherwise.
func (tb *tokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	tb.tokens += now.Sub(tb.lastRefill).Seconds() * tb.refillRate
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}
	tb.lastRefill = now

	if tb.tokens >= 1 {
		tb.tokens--
		return true
	}
	return false
}
