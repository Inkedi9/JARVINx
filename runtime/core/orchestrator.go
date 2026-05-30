package core

import (
	"context"
	"fmt"
	"sync"

	"github.com/Inkedi9/jarvinx/agents"
	"github.com/Inkedi9/jarvinx/jxlog"
	"github.com/Inkedi9/jarvinx/memory"
	"github.com/Inkedi9/jarvinx/tools"
)

type Orchestrator struct {
	bus      *Bus
	registry *agents.Registry
	state    *memory.State
	logger   *memory.Logger
	mu       sync.Mutex
	lastSnap memory.Snapshot
	snapMu   sync.RWMutex
}

func NewOrchestrator(
	bus *Bus,
	registry *agents.Registry,
	state *memory.State,
	logger *memory.Logger,
) *Orchestrator {
	return &Orchestrator{
		bus:      bus,
		registry: registry,
		state:    state,
		logger:   logger,
	}
}

func (o *Orchestrator) AgentContext() agents.AgentContext {
	o.snapMu.RLock()
	defer o.snapMu.RUnlock()
	return agents.AgentContext{
		Snapshot: o.lastSnap,
		State:    o.state,
		Logger:   o.logger,
	}
}

func (o *Orchestrator) Start(ctx context.Context) {
	jxlog.Info("ORCHESTRATOR", "En écoute sur le bus...")

	// Nom unique pour ce subscriber
	events := o.bus.Subscribe("orchestrator")

	go o.registry.Start(ctx, o.AgentContext)

	for {
		select {
		case <-ctx.Done():
			o.bus.Unsubscribe("orchestrator")
			jxlog.Info("ORCHESTRATOR", "Arrêt propre")
			return

		case event, ok := <-events:
			if !ok {
				return
			}
			switch event.Type {
			case EventObserved:
				snap, ok := event.Payload.(memory.Snapshot)
				if !ok {
					continue
				}
				go o.handleObserved(snap)
			case EventError:
				msg, _ := event.Payload.(string)
				jxlog.Error("ORCHESTRATOR", msg)
			}
		}
	}
}

func (o *Orchestrator) handleObserved(snap memory.Snapshot) {
	if !o.mu.TryLock() {
		jxlog.Debug("ORCHESTRATOR", "Cycle précédent en cours — tick ignoré")
		return
	}
	defer o.mu.Unlock()

	// Met à jour le dernier snapshot pour le registry
	o.snapMu.Lock()
	o.lastSnap = snap
	o.snapMu.Unlock()

	// Log
	entry := memory.LogEntry(snap)
	if err := o.logger.Write(entry); err != nil {
		jxlog.Error("ORCHESTRATOR", fmt.Sprintf("Log : %v", err))
	}

	// Run command if action is execute
	cycles := o.state.LastCycles(1)
	if len(cycles) > 0 && cycles[0].Command != "" {
		result := tools.ExecuteCommand(cycles[0].Command)
		result.Display()
		o.bus.Publish(Event{Type: EventExecuted, Payload: result})
	}

	fmt.Println()
}
