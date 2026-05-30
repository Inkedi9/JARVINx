package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Inkedi9/jarvinx/config"
	"github.com/Inkedi9/jarvinx/core"
	jxlog "github.com/Inkedi9/jarvinx/jxlog"
	"github.com/Inkedi9/jarvinx/llm"
)

// Version est injectée au build via -ldflags "-X main.Version=x.y.z"
var Version = "dev"

func main() {
	// Flag --dry-run — doit être parsé avant tout
	dryRun := flag.Bool("dry-run", false, "Mode simulation — aucune action réelle exécutée")
	flag.Parse()

	// Logger en premier
	debug := os.Getenv("JARVINX_DEBUG") == "true"
	jxlog.Init(debug)

	config.LoadEnv(".env")

	cfg := config.Default()
	cfg.FromEnv()

	// --dry-run CLI a priorité sur la variable d'env
	if *dryRun {
		cfg.DryRun = true
	}

	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr,
			"\n\033[31m[ JARVINX ]\033[0m Configuration invalide :\n%v\n\n", err)
		os.Exit(1)
	}

	if cfg.DryRun {
		fmt.Println("\033[33m╔══════════════════════════════════════════════╗\033[0m")
		fmt.Println("\033[33m║         ⚠  MODE DRY-RUN ACTIVÉ  ⚠           ║\033[0m")
		fmt.Println("\033[33m║   Aucune commande ni alerte ne sera exécutée ║\033[0m")
		fmt.Println("\033[33m╚══════════════════════════════════════════════╝\033[0m")
		fmt.Println()
	}

	jxlog.Info("JARVINX", fmt.Sprintf("Modèle : %s | Intervalle : %v | CPU : %.0f%% RAM : %.0f%% Disk : %.0f%%",
		cfg.Model, cfg.Interval,
		cfg.CPUAlertThreshold, cfg.RAMAlertThreshold, cfg.DiskAlertThreshold,
	))

	fmt.Printf("\033[90m[ JARVINX ]\033[0m Vérification Ollama...\n")
	health := llm.CheckOllama(cfg.OllamaURL, cfg.Model)
	health.Display(cfg.Model)

	if !health.Online {
		fmt.Fprintf(os.Stderr,
			"\n\033[31m[ JARVINX ]\033[0m Ollama requis.\n"+
				"  Lance : ollama serve\n"+
				"  Puis  : ollama pull %s\n\n", cfg.Model)
		os.Exit(1)
	}

	if health.Error != "" {
		fmt.Printf("\033[33m[ WARN ]\033[0m %s\n", health.Error)
	}

	rt := core.NewRuntime(cfg, Version)
	rt.Start()
}
