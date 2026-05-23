package core

import (
	"fmt"

	"github.com/Inkedi9/jarvinx/agents"
	"github.com/Inkedi9/jarvinx/config"
	"github.com/Inkedi9/jarvinx/memory"
	"github.com/Inkedi9/jarvinx/web"
)

type Runtime struct {
	cfg          *config.Config
	bus          *Bus
	scheduler    *Scheduler
	orchestrator *Orchestrator
	cli          *CLI
	webServer    *web.Server
}

func NewRuntime(cfg *config.Config) *Runtime {
	bus := NewBus(10)
	state := memory.NewState(cfg.StateFile)
	logger := memory.NewLogger(cfg.LogFile)
	agent := agents.NewSystemAgent(cfg.OllamaURL, cfg.Model)
	scheduler := NewScheduler(cfg.Interval, bus)

	return &Runtime{
		cfg:          cfg,
		bus:          bus,
		scheduler:    scheduler,
		orchestrator: NewOrchestrator(bus, agent, state, logger),
		cli:          NewCLI(state, scheduler),
		webServer:    web.NewServer(cfg, state, cfg.WebPort),
	}
}

func (r *Runtime) Start() {
	fmt.Println("╔══════════════════════════════════════════════╗")
	fmt.Println("║           JARVINX — RUNTIME v0.4            ║")
	fmt.Println("╚══════════════════════════════════════════════╝")
	fmt.Printf("  Modèle     : %s\n", r.cfg.Model)
	fmt.Printf("  Intervalle : %v\n", r.cfg.Interval)
	fmt.Println()

	go r.orchestrator.Start()
	go r.scheduler.Start()
	go r.cli.Start()
	go r.webServer.Start()

	select {}
}
