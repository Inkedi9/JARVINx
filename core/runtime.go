package core

import (
	"fmt"

	"github.com/Inkedi9/jarvinx/agents"
	"github.com/Inkedi9/jarvinx/config"
	"github.com/Inkedi9/jarvinx/memory"
)

type Runtime struct {
	cfg          *config.Config
	bus          *Bus
	scheduler    *Scheduler
	orchestrator *Orchestrator
}

func NewRuntime(cfg *config.Config) *Runtime {
	bus := NewBus(10)
	state := memory.NewState(cfg.StateFile)
	logger := memory.NewLogger(cfg.LogFile)
	agent := agents.NewSystemAgent(cfg.OllamaURL, cfg.Model)

	return &Runtime{
		cfg:          cfg,
		bus:          bus,
		scheduler:    NewScheduler(cfg.Interval, bus),
		orchestrator: NewOrchestrator(bus, agent, state, logger),
	}
}

func (r *Runtime) Start() {
	fmt.Println("╔══════════════════════════════════════════════╗")
	fmt.Println("║           JARVINX — RUNTIME v0.2            ║")
	fmt.Println("╚══════════════════════════════════════════════╝")
	fmt.Printf("  Modèle     : %s\n", r.cfg.Model)
	fmt.Printf("  Intervalle : %v\n", r.cfg.Interval)
	fmt.Println()

	// Orchestrateur dans sa propre goroutine — écoute le bus
	go r.orchestrator.Start()

	// Scheduler dans sa propre goroutine — émet les ticks
	go r.scheduler.Start()

	// Bloquer le main — attendre indéfiniment
	select {}
}
