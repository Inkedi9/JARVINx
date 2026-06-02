package agents

import (
	"context"
	"sync"
	"time"

	"github.com/Inkedi9/jarvinx/memory"
)

// AgentContext est passé à chaque Run() — tout ce dont un agent peut avoir besoin
type AgentContext struct {
	Snapshot         memory.Snapshot
	State            memory.Store
	Logger           memory.EventLog
	SimilarDecisions []string // décisions passées similaires — rempli par QdrantAgent (v1.8), nil sinon
}

// AgentStatus représente l'état observable d'un agent
type AgentStatus struct {
	Name       string        `json:"name"`
	Enabled    bool          `json:"enabled"`
	LastRun    time.Time     `json:"last_run"`
	LastError  string        `json:"last_error,omitempty"`
	RunCount   int           `json:"run_count"`
	ErrorCount int           `json:"error_count"`
	AlertCount int           `json:"alert_count"`
	Schedule   time.Duration `json:"schedule_ms"`
}

// Agent est l'interface que tout agent doit implémenter
type Agent interface {
	Name() string
	Schedule() time.Duration
	Run(ctx context.Context, actx AgentContext) error
	Status() AgentStatus
	IsEnabled() bool
	Enable()
	Disable()
}

// BaseAgent fournit l'implémentation commune — à embedder dans chaque agent
type BaseAgent struct {
	name     string
	schedule time.Duration
	enabled  bool
	mu       sync.RWMutex

	lastRun    time.Time
	lastError  string
	runCount   int
	errorCount int
	alertCount int
}

func NewBaseAgent(name string, schedule time.Duration) BaseAgent {
	return BaseAgent{
		name:     name,
		schedule: schedule,
		enabled:  true,
	}
}

func (b *BaseAgent) Name() string            { return b.name }
func (b *BaseAgent) Schedule() time.Duration { return b.schedule }

func (b *BaseAgent) IsEnabled() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.enabled
}

func (b *BaseAgent) Enable() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.enabled = true
}

func (b *BaseAgent) Disable() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.enabled = false
}

func (b *BaseAgent) Status() AgentStatus {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return AgentStatus{
		Name:       b.name,
		Enabled:    b.enabled,
		LastRun:    b.lastRun,
		LastError:  b.lastError,
		RunCount:   b.runCount,
		ErrorCount: b.errorCount,
		AlertCount: b.alertCount,
		Schedule:   b.schedule,
	}
}

// recordRun met à jour les stats — appelé par le registry après chaque Run
func (b *BaseAgent) recordSuccess() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.lastRun = time.Now()
	b.runCount++
	b.lastError = ""
}

func (b *BaseAgent) recordError(err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.lastRun = time.Now()
	b.runCount++
	b.errorCount++
	b.lastError = err.Error()
}

func (b *BaseAgent) recordAlert() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.lastRun = time.Now()
	b.runCount++
	b.alertCount++
	b.lastError = ""
}
