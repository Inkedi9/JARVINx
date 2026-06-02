package core

import (
	"context"
	"fmt"
	"io"
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
	cfg           *config.Config
	version       string
	bus           *Bus
	scheduler     *Scheduler
	orchestrator  *Orchestrator
	cli           *CLI
	webServer     *web.Server
	registry      *agents.Registry
	dailyReporter *agents.DailyReporter
	alertLogger   *memory.Logger
	sqliteCloser  io.Closer // non-nil when SQLiteStore is active
}

func NewRuntime(cfg *config.Config, version string) *Runtime {
	bus := NewBus(10)
	state := memory.NewState(cfg.StateFile)

	// Build the active Store: DoubleWriteStore when SQLite is configured, JSON-only otherwise.
	var store memory.Store = state
	var sqliteCloser io.Closer
	var historyReader memory.HistoryReader
	if cfg.SQLitePath != "" {
		sq, err := memory.OpenSQLiteStore(cfg.SQLitePath)
		if err != nil {
			jxlog.Warn("SQLITE", fmt.Sprintf("ouverture échouée (%v) — JSON seul", err))
			store = memory.NewDoubleWriteStore(state, memory.NoopStore{})
		} else {
			store = memory.NewDoubleWriteStore(state, sq)
			sqliteCloser = sq
			historyReader = sq
			jxlog.Info("SQLITE", fmt.Sprintf("store actif : %s", cfg.SQLitePath))
		}
	}

	logger := memory.NewLoggerWithRotation(
		cfg.LogFile,
		cfg.LogMaxSizeBytes,
		cfg.LogMaxBackups,
	)
	alertLogger := memory.NewLoggerWithRotation(
		cfg.AlertFile,
		cfg.AlertMaxSizeBytes,
		cfg.AlertMaxBackups,
	)
	registry := agents.NewRegistry()

	// Construit le dispatcher de notifications
	dispatcher := agents.NewNotifierDispatcher(cfg.DryRun)
	if cfg.DiscordWebhook != "" {
		dispatcher.Register(agents.NewDiscordNotifier(cfg.DiscordWebhook))
	}
	if cfg.SlackWebhook != "" {
		dispatcher.Register(agents.NewSlackNotifier(cfg.SlackWebhook))
	}
	if cfg.GotifyURL != "" && cfg.GotifyToken != "" {
		dispatcher.Register(agents.NewGotifyNotifier(cfg.GotifyURL, cfg.GotifyToken))
	}
	if cfg.NtfyTopic != "" {
		dispatcher.Register(agents.NewNtfyNotifier(cfg.NtfyURL, cfg.NtfyTopic))
	}

	registry.Register(agents.NewSystemAgent(cfg.OllamaURL, cfg.Model, cfg.CPUAlertThreshold, cfg.RAMAlertThreshold, cfg.DiskAlertThreshold))
	registry.Register(agents.NewAlertAgent(
		cfg.CPUAlertThreshold,
		cfg.RAMAlertThreshold,
		cfg.DiskAlertThreshold,
		cfg.AlertMinCycles,
		cfg.AlertCooldown,
		cfg.AlertFile,
		dispatcher,
	))

	if cfg.DockerEnabled {
		registry.Register(agents.NewDockerAgent(
			cfg.DryRun,
			cfg.DockerWatchList...,
		))
	}

	if cfg.FileEnabled && len(cfg.FileWatchPaths) > 0 {
		registry.Register(agents.NewFileAgent(
			cfg.FileWatchPaths,
			cfg.FileMaxSizeMB,
			cfg.DryRun,
		))
	}

	// QdrantAgent — v1.8 — enregistré seulement si JARVINX_QDRANT_URL est défini
	if cfg.QdrantURL != "" {
		registry.Register(agents.NewQdrantAgent(cfg.QdrantURL, cfg.OllamaURL))
		jxlog.Info("QDRANT", fmt.Sprintf("mémoire sémantique activée : %s", cfg.QdrantURL))
	}

	scheduler := NewScheduler(cfg.Interval, bus)
	orchestrator := NewOrchestrator(bus, registry, store, logger, cfg.DryRun, cfg.ExecCooldown)

	var dailyReporter *agents.DailyReporter
	if cfg.DailyReportEnabled {
		dailyReporter = agents.NewDailyReporter(
			dispatcher,
			store,
			cfg.DailyReportHour,
			cfg.DailyReportMinute,
			cfg.DryRun,
		)
	}

	return &Runtime{
		cfg:          cfg,
		version:      version,
		bus:          bus,
		scheduler:    scheduler,
		orchestrator: orchestrator,
		cli:          NewCLI(state, scheduler),
		webServer: web.NewServer(
			cfg, state, registry,
			logger, alertLogger,
			dailyReporter,
			orchestrator,
			historyReader,
			cfg.WebPort,
			web.StaticFiles(),
		),
		alertLogger:   alertLogger,
		registry:      registry,
		dailyReporter: dailyReporter,
		sqliteCloser:  sqliteCloser,
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

	if r.dailyReporter != nil {
		go r.dailyReporter.Start(ctx)
		jxlog.Info("JARVINX", fmt.Sprintf(
			"Rapport quotidien activé — %02d:%02d",
			r.cfg.DailyReportHour,
			r.cfg.DailyReportMinute,
		))
	}

	go r.orchestrator.Start(ctx)
	go r.scheduler.Start(ctx)
	go r.cli.Start()
	go r.webServer.Start()

	<-ctx.Done()
	if r.sqliteCloser != nil {
		if err := r.sqliteCloser.Close(); err != nil {
			jxlog.Warn("SQLITE", fmt.Sprintf("fermeture : %v", err))
		}
	}
	jxlog.Info("JARVINX", "Arrêt terminé. À bientôt.")
}
