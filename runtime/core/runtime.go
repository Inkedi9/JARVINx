package core

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Inkedi9/jarvinx/agents"
	"github.com/Inkedi9/jarvinx/config"
	"github.com/Inkedi9/jarvinx/jxlog"
	"github.com/Inkedi9/jarvinx/memory"
	"github.com/Inkedi9/jarvinx/web"
)

type Runtime struct {
	cfg          *config.Config
	version      string
	bus          *Bus
	scheduler    *Scheduler
	orchestrator *Orchestrator
	cli          *CLI
	webServer    *web.Server
	registry     *agents.Registry
}

func NewRuntime(cfg *config.Config, version string) *Runtime {
	bus := NewBus(10)
	state := memory.NewState(cfg.StateFile)
	logger := memory.NewLoggerWithRotation(
		cfg.LogFile,
		cfg.LogMaxSizeBytes,
		cfg.LogMaxBackups,
	)
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
		cfg.DryRun,
	))
	if cfg.DockerEnabled {
		registry.Register(agents.NewDockerAgent(
			cfg.DiscordWebhook,
			cfg.DryRun,
			cfg.DockerWatchList...,
		))
	}

	scheduler := NewScheduler(cfg.Interval, bus)
	orchestrator := NewOrchestrator(bus, registry, state, logger, cfg.DryRun)

	return &Runtime{
		cfg:          cfg,
		version:      version,
		bus:          bus,
		scheduler:    scheduler,
		orchestrator: orchestrator,
		cli:          NewCLI(state, scheduler),
		webServer:    web.NewServer(cfg, state, registry, cfg.WebPort, web.StaticFiles()),
		registry:     registry,
	}
}

func (r *Runtime) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		jxlog.Warn("JARVINX", fmt.Sprintf("Signal reçu : %v — arrêt propre...", sig))
		cancel()
	}()

	fmt.Println("\033[36m╔══════════════════════════════════════════════╗\033[0m")
	fmt.Printf("\033[36m║\033[0m        \033[97mJARVINx — RUNTIME v%s\033[0m", r.version)
	// Padding dynamique selon la longueur de la version
	padding := 20 - len(r.version)
	for i := 0; i < padding; i++ {
		fmt.Print(" ")
	}
	fmt.Println("\033[36m║\033[0m")
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
	go r.scheduler.Start(ctx)
	go r.cli.Start()
	go r.webServer.Start()

	<-ctx.Done()
	jxlog.Info("JARVINX", "Arrêt terminé. À bientôt.")
}
