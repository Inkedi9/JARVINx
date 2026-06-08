package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Inkedi9/jarvinx/agents"
	"github.com/Inkedi9/jarvinx/jxlog"
	"github.com/Inkedi9/jarvinx/memory"
	"github.com/Inkedi9/jarvinx/tools"
)

// executeGuard prevents the same command from running too frequently.
type executeGuard struct {
	mu         sync.Mutex
	LastCmd    string
	lastExecAt time.Time
	cooldown   time.Duration

	resultMu   sync.Mutex
	lastResult *tools.CommandResult
}

func (g *executeGuard) setLastResult(r tools.CommandResult) {
	g.resultMu.Lock()
	g.lastResult = &r
	g.resultMu.Unlock()
}

func (g *executeGuard) Allow(cmd string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	if cmd == g.LastCmd && time.Since(g.lastExecAt) < g.cooldown {
		return false
	}
	g.LastCmd = cmd
	g.lastExecAt = time.Now()
	return true
}

func (g *executeGuard) CooldownRemaining() time.Duration {
	elapsed := time.Since(g.lastExecAt)
	if elapsed >= g.cooldown {
		return 0
	}
	return g.cooldown - elapsed
}

// ExecGuardStatus retourne l'état courant de l'execute guard de façon thread-safe.
func (o *Orchestrator) ExecGuardStatus() (string, time.Duration) {
	o.execGuard.mu.Lock()
	defer o.execGuard.mu.Unlock()
	return o.execGuard.LastCmd, o.execGuard.CooldownRemaining()
}

// LastExecResultStatus retourne le résultat de la dernière commande exécutée.
func (o *Orchestrator) LastExecResultStatus() (tools.CommandResult, bool) {
	o.execGuard.resultMu.Lock()
	defer o.execGuard.resultMu.Unlock()
	if o.execGuard.lastResult == nil {
		return tools.CommandResult{}, false
	}
	return *o.execGuard.lastResult, true
}

type Orchestrator struct {
	bus       *Bus
	registry  *agents.Registry
	state     memory.Store
	logger    memory.EventLog
	mu        sync.Mutex
	lastSnap  memory.Snapshot
	snapMu    sync.RWMutex
	dryRun    bool
	execGuard *executeGuard

	// similarProvider est nil quand Qdrant n'est pas configuré
	similarProvider agents.SimilarDecisionsProvider
}

// SetSimilarDecisionsProvider câble le QdrantAgent comme source de contexte sémantique.
// Doit être appelé avant Start().
func (o *Orchestrator) SetSimilarDecisionsProvider(p agents.SimilarDecisionsProvider) {
	o.snapMu.Lock()
	defer o.snapMu.Unlock()
	o.similarProvider = p
}

func NewOrchestrator(
	bus *Bus,
	registry *agents.Registry,
	state memory.Store,
	logger memory.EventLog,
	dryRun bool,
	execCooldown time.Duration,
) *Orchestrator {
	return &Orchestrator{
		bus:       bus,
		registry:  registry,
		state:     state,
		logger:    logger,
		dryRun:    dryRun,
		execGuard: &executeGuard{cooldown: execCooldown},
	}
}

func (o *Orchestrator) AgentContext() agents.AgentContext {
	o.snapMu.RLock()
	defer o.snapMu.RUnlock()

	var similar []string
	if o.similarProvider != nil {
		similar = o.similarProvider.LastSimilarDecisions()
	}

	return agents.AgentContext{
		Snapshot:         o.lastSnap,
		State:            o.state,
		Logger:           o.logger,
		SimilarDecisions: similar,
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

// shouldExecute vérifie que la condition ayant motivé l'action est toujours active.
// Si tous les triggers sont à 0 (record ancien sans ces champs), retourne true pour la rétro-compatibilité.
func (o *Orchestrator) shouldExecute(cycle memory.CycleRecord, current memory.Snapshot) bool {
	const margin = 5.0

	if cycle.TriggerCPU == 0 && cycle.TriggerRAM == 0 && cycle.TriggerDisk == 0 {
		return true
	}

	if cycle.TriggerCPU > 0 && current.CPUPercent < cycle.TriggerCPU-margin {
		jxlog.Info("VERIFY", fmt.Sprintf("CPU normals (%.1f%% → %.1f%%) — execute annulé",
			cycle.TriggerCPU, current.CPUPercent))
		return false
	}
	if cycle.TriggerRAM > 0 && current.MemPercent < cycle.TriggerRAM-margin {
		jxlog.Info("VERIFY", fmt.Sprintf("RAM normals (%.1f%% → %.1f%%) — execute annulé",
			cycle.TriggerRAM, current.MemPercent))
		return false
	}
	if cycle.TriggerDisk > 0 && current.DiskPercent < cycle.TriggerDisk-margin {
		jxlog.Info("VERIFY", fmt.Sprintf("Disk normals (%.1f%% → %.1f%%) — execute annulé",
			cycle.TriggerDisk, current.DiskPercent))
		return false
	}

	return true
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
		cmd := cycles[0].Command
		if !o.shouldExecute(cycles[0], snap) {
			// annulation loguée dans shouldExecute
		} else if !o.execGuard.Allow(cmd) {
			jxlog.Info("EXEC-GUARD", fmt.Sprintf("cooldown actif — '%s' ignorée", cmd))
		} else if o.dryRun {
			jxlog.Info("DRY-RUN", fmt.Sprintf("Commande '%s' simulée — non exécutée", cmd))
			result := tools.ExecuteCommandDryRun(cmd)
			result.Display()
			o.execGuard.setLastResult(result)
		} else {
			result := tools.ExecuteCommand(cmd)
			result.Display()
			o.execGuard.setLastResult(result)
			o.bus.Publish(Event{Type: EventExecuted, Payload: result})
		}
	}

	fmt.Println()
}
