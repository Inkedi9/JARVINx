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
	bus    *Bus
	agent  *agents.SystemAgent
	state  *memory.State
	logger *memory.Logger
	mu     sync.Mutex
}

func NewOrchestrator(
	bus *Bus,
	agent *agents.SystemAgent,
	state *memory.State,
	logger *memory.Logger,
) *Orchestrator {
	return &Orchestrator{
		bus:    bus,
		agent:  agent,
		state:  state,
		logger: logger,
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
				fmt.Println("[ ORCHESTRATOR ] Payload invalide pour EventObserved")
				continue
			}
			// Lance le traitement dans une goroutine séparée
			// mais le mutex empêche deux cycles simultanés
			go o.handleObserved(snap)

		case EventError:
			msg, ok := event.Payload.(string)
			if !ok {
				continue
			}
			fmt.Printf("[ ERREUR ] %s\n", msg)
		}
	}
}

func (o *Orchestrator) handleObserved(snap memory.Snapshot) {
	// TryLock — si un cycle tourne déjà, on abandonne ce tick
	if !o.mu.TryLock() {
		fmt.Println("[ ORCHESTRATOR ] Cycle précédent encore en cours — tick ignoré")
		return
	}
	defer o.mu.Unlock()

	// Mémoriser + logger
	o.state.Add(snap)
	if err := o.state.Save(); err != nil {
		fmt.Printf("[ ERREUR ] State save : %v\n", err)
	}

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

	// Penser
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
		return
	}
	decision.Display()

	o.bus.Publish(Event{
		Type:    EventDecided,
		Payload: decision,
	})

	// Agir
	if decision.Command != "" {
		fmt.Printf("[ EXEC ] Exécution : '%s'\n", decision.Command)
		result := tools.ExecuteCommand(decision.Command)
		result.Display()

		o.bus.Publish(Event{
			Type:    EventExecuted,
			Payload: result,
		})
	}

	fmt.Println()
}
