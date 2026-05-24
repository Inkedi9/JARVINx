package agents

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Registry gère le cycle de vie de tous les agents
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
	fmt.Printf("[ REGISTRY ] Agent enregistré : %s (schedule: %v)\n",
		a.Name(), a.Schedule())
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
		fmt.Printf("[ REGISTRY ] Agent activé : %s\n", name)
		return true
	}
	return false
}

func (r *Registry) Disable(name string) bool {
	if a, ok := r.Get(name); ok {
		a.Disable()
		fmt.Printf("[ REGISTRY ] Agent désactivé : %s\n", name)
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

// Start lance chaque agent dans sa propre goroutine avec son propre ticker
func (r *Registry) Start(ctx context.Context, actxFn func() AgentContext) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, agent := range r.agents {
		go r.runAgent(ctx, agent, actxFn)
	}
}

func (r *Registry) runAgent(ctx context.Context, a Agent, actxFn func() AgentContext) {
	fmt.Printf("[ REGISTRY ] Démarrage agent : %s\n", a.Name())

	ticker := time.NewTicker(a.Schedule())
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("[ REGISTRY ] Agent arrêté : %s\n", a.Name())
			return

		case <-ticker.C:
			if !a.IsEnabled() {
				continue
			}

			// Chaque agent dans sa propre goroutine — un panic ne tue pas les autres
			func() {
				defer func() {
					if rec := recover(); rec != nil {
						fmt.Printf("[ REGISTRY ] Panic récupéré dans %s : %v\n",
							a.Name(), rec)
					}
				}()

				actx := actxFn()
				if err := a.Run(ctx, actx); err != nil {
					fmt.Printf("[ REGISTRY ] Erreur agent %s : %v\n",
						a.Name(), err)
				}
			}()
		}
	}
}
