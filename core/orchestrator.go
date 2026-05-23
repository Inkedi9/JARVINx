package core

import (
	"fmt"
	"sync"

	"github.com/Inkedi9/jarvinx/agents"
	"github.com/Inkedi9/jarvinx/llm"
	"github.com/Inkedi9/jarvinx/memory"
	"github.com/Inkedi9/jarvinx/tools"
)

type Orchestrator struct {
	bus        *Bus
	agent      *agents.SystemAgent
	alertAgent *agents.AlertAgent
	state      *memory.State
	logger     *memory.Logger
	mu         sync.Mutex
}

func NewOrchestrator(
	bus *Bus,
	agent *agents.SystemAgent,
	alertAgent *agents.AlertAgent,
	state *memory.State,
	logger *memory.Logger,
) *Orchestrator {
	return &Orchestrator{
		bus:        bus,
		agent:      agent,
		alertAgent: alertAgent,
		state:      state,
		logger:     logger,
	}
}

func (o *Orchestrator) Start() {
	fmt.Println("[ ORCHESTRATOR ] En écoute sur le bus...")
	events := o.bus.Subscribe()
	for event := range events {
		switch event.Type {
		case EventObserved:
			snap, ok := event.Payload.(memory.Snapshot)
			if !ok {
				continue
			}
			go o.handleObserved(snap)
		case EventError:
			msg, _ := event.Payload.(string)
			fmt.Printf("[ ERREUR ] %s\n", msg)
		}
	}
}

func (o *Orchestrator) handleObserved(snap memory.Snapshot) {
	if !o.mu.TryLock() {
		fmt.Println("[ ORCHESTRATOR ] Cycle précédent encore en cours — tick ignoré")
		return
	}
	defer o.mu.Unlock()

	// 1. Alertes — avant le LLM, instantané
	alerts := o.alertAgent.Analyze(snap)
	o.alertAgent.Dispatch(alerts)

	// 2. Log
	entry := memory.LogEntry{
		Timestamp:   snap.Timestamp,
		CPUPercent:  snap.CPUPercent,
		MemUsed:     snap.MemUsed,
		MemTotal:    snap.MemTotal,
		MemPercent:  snap.MemPercent,
		DiskUsed:    snap.DiskUsed,
		DiskTotal:   snap.DiskTotal,
		DiskPercent: snap.DiskPercent,
	}
	if err := o.logger.Write(entry); err != nil {
		fmt.Printf("[ ERREUR ] Log : %v\n", err)
	}

	// 3. Think
	fmt.Println("[ AGENT ] Analyse en cours...")
	ctx := llm.SystemContext{
		Timestamp:   snap.Timestamp,
		CPUPercent:  snap.CPUPercent,
		MemUsed:     snap.MemUsed,
		MemTotal:    snap.MemTotal,
		MemPercent:  snap.MemPercent,
		DiskUsed:    snap.DiskUsed,
		DiskTotal:   snap.DiskTotal,
		DiskPercent: snap.DiskPercent,
		History:     o.state.Last(5),
	}

	decision, err := o.agent.Decide(ctx)
	if err != nil {
		fmt.Printf("[ ERREUR ] Agent : %v\n", err)
		o.state.Add(snap)
		o.state.Save()
		return
	}
	decision.Display()

	// 4. Enregistrer le cycle
	record := memory.CycleRecord{
		Snapshot:  snap,
		Action:    decision.Action,
		Analysis:  decision.Analysis,
		Reason:    decision.Reason,
		Command:   decision.Command,
		Timestamp: snap.Timestamp,
	}
	o.state.AddCycle(record)
	o.state.Add(snap)
	if err := o.state.Save(); err != nil {
		fmt.Printf("[ ERREUR ] State : %v\n", err)
	}
	fmt.Printf("[ STATE ] Cycle #%d enregistré\n", o.state.CycleNum)

	// 5. Act
	if decision.Command != "" {
		fmt.Printf("[ EXEC ] Exécution : '%s'\n", decision.Command)
		result := tools.ExecuteCommand(decision.Command)
		result.Display()
		o.bus.Publish(Event{Type: EventExecuted, Payload: result})
	}

	o.bus.Publish(Event{Type: EventDecided, Payload: decision})
	fmt.Println()
}
