package agents

import (
	"context"
	"fmt"
	"sync"
	"time"

	jxlog "github.com/Inkedi9/jarvinx/jxlog"
)

type Registry struct {
	agents []Agent
	mu     sync.RWMutex
}

func NewRegistry() *Registry {
	return &Registry{}
}

func (r *Registry) Register(a Agent) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.agents = append(r.agents, a)
	jxlog.Info("REGISTRY", fmt.Sprintf("Agent enregistré : %s (schedule: %v)",
		a.Name(), a.Schedule()))
}

func (r *Registry) Get(name string) (Agent, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, a := range r.agents {
		if a.Name() == name {
			return a, true
		}
	}
	return nil, false
}

func (r *Registry) Enable(name string) bool {
	if a, ok := r.Get(name); ok {
		a.Enable()
		jxlog.Info("REGISTRY", fmt.Sprintf("Agent activé : %s", name))
		return true
	}
	return false
}

func (r *Registry) Disable(name string) bool {
	if a, ok := r.Get(name); ok {
		a.Disable()
		jxlog.Info("REGISTRY", fmt.Sprintf("Agent désactivé : %s", name))
		return true
	}
	return false
}

func (r *Registry) Statuses() []AgentStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()
	statuses := make([]AgentStatus, 0, len(r.agents))
	for _, a := range r.agents {
		statuses = append(statuses, a.Status())
	}
	return statuses
}

func (r *Registry) Start(ctx context.Context, actxFn func() AgentContext) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, agent := range r.agents {
		go r.runAgent(ctx, agent, actxFn)
	}
}

func (r *Registry) runAgent(ctx context.Context, a Agent, actxFn func() AgentContext) {
	jxlog.Info("REGISTRY", fmt.Sprintf("Starting agent : %s", a.Name()))

	ticker := time.NewTicker(a.Schedule())
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			jxlog.Info("REGISTRY", fmt.Sprintf("Agent arrêté : %s", a.Name()))
			return

		case <-ticker.C:
			if !a.IsEnabled() {
				continue
			}

			func() {
				defer func() {
					if rec := recover(); rec != nil {
						jxlog.Error("REGISTRY", fmt.Sprintf("Panic dans %s : %v",
							a.Name(), rec))
					}
				}()

				actx := actxFn()
				if err := a.Run(ctx, actx); err != nil {
					jxlog.Error("REGISTRY", fmt.Sprintf("Erreur agent %s : %v",
						a.Name(), err))
				}
			}()
		}
	}
}
