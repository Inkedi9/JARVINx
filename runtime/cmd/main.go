package main

import (
	"fmt"
	"os"

	"github.com/Inkedi9/jarvinx/config"
	"github.com/Inkedi9/jarvinx/core"
	jxlog "github.com/Inkedi9/jarvinx/jxlog"
	"github.com/Inkedi9/jarvinx/llm"
)

// Version est injectée au build via -ldflags "-X main.Version=x.y.z"
// Valeur par défaut pour le dev local
var Version = "dev"

// la fonction principale. C'est ici que tout commence quand tu lances le binaire.
func main() {
	// Logger en premier — avant tout le reste
	debug := os.Getenv("JARVINX_DEBUG") == "true"
	jxlog.Init(debug)
	// Charge .env juste apres le logger
	config.LoadEnv(".env")

	cfg := config.Default()
	cfg.FromEnv() // surcharge depuis les variables d'environnement

	// Validation — on sort immédiatement si la config est invalide
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr,
			"\n\033[31m[ JARVINX ]\033[0m Configuration invalide :\n%v\n\n", err)
		os.Exit(1)
	}

	// Affiche la config active
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
