package core

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

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
	registry     *agents.Registry
}

func NewRuntime(cfg *config.Config) *Runtime {
	bus := NewBus(10)
	state := memory.NewState(cfg.StateFile)
	logger := memory.NewLogger(cfg.LogFile)
	registry := agents.NewRegistry()

	// Enregistrement des agents
	registry.Register(agents.NewSystemAgent(cfg.OllamaURL, cfg.Model))
	registry.Register(agents.NewAlertAgent(
		cfg.CPUAlertThreshold,
		cfg.RAMAlertThreshold,
		cfg.DiskAlertThreshold,
		cfg.AlertMinCycles,
		cfg.AlertCooldown,
		cfg.AlertFile,
		cfg.DiscordWebhook,
	))

	scheduler := NewScheduler(cfg.Interval, bus)
	orchestrator := NewOrchestrator(bus, registry, state, logger)

	return &Runtime{
		cfg:          cfg,
		bus:          bus,
		scheduler:    scheduler,
		orchestrator: orchestrator,
		cli:          NewCLI(state, scheduler),
		webServer:    web.NewServer(cfg, state, cfg.WebPort, web.StaticFiles()),
		registry:     registry,
	}
}

func (r *Runtime) Start() {
	// Context annulable — Ctrl+C déclenche un shutdown propre
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Écoute SIGINT et SIGTERM
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		fmt.Printf("\n\033[33m[ JARVINX ]\033[0m Signal reçu : %v — arrêt propre...\n", sig)
		cancel()
	}()

	fmt.Println("\033[36m╔══════════════════════════════════════════════╗\033[0m")
	fmt.Println("\033[36m║\033[0m           \033[97mJARVINx — RUNTIME v0.6\033[0m            \033[36m║\033[0m")
	fmt.Println("\033[36m╚══════════════════════════════════════════════╝\033[0m")
	fmt.Printf("  Modèle     : \033[97m%s\033[0m\n", r.cfg.Model)
	fmt.Printf("  Intervalle : \033[97m%v\033[0m\n", r.cfg.Interval)
	fmt.Printf("  Seuils     : CPU \033[33m%.0f%%\033[0m · RAM \033[33m%.0f%%\033[0m · Disk \033[33m%.0f%%\033[0m\n",
		r.cfg.CPUAlertThreshold,
		r.cfg.RAMAlertThreshold,
		r.cfg.DiskAlertThreshold,
	)
	fmt.Println()

	go r.orchestrator.Start(ctx)
	go r.scheduler.Start()
	go r.cli.Start()
	go r.webServer.Start()

	// Attend l'annulation du context
	<-ctx.Done()
	fmt.Println("\033[90m[ JARVINX ]\033[0m Arrêt terminé. À bientôt.")
}
