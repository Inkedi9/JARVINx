package main

import (
	"fmt"
	"os"

	"github.com/Inkedi9/jarvinx/config"
	"github.com/Inkedi9/jarvinx/core"
	"github.com/Inkedi9/jarvinx/llm"
)

// la fonction principale. C'est ici que tout commence quand tu lances le binaire.
func main() {
	// Charge .env avant tout
	config.LoadEnv(".env")

	cfg := config.Default()
	cfg.DiscordWebhook = os.Getenv("DISCORD_WEBHOOK")

	// Validation — on sort immédiatement si la config est invalide
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "\n[ JARVINX ] Erreur de configuration :\n%v\n\n", err)
		os.Exit(1)
	}

	// 2. Health check Ollama
	fmt.Printf("\033[90m[ JARVINX ]\033[0m Vérification Ollama...\n")
	health := llm.CheckOllama(cfg.OllamaURL, cfg.Model)
	health.Display(cfg.Model)

	if !health.Online {
		fmt.Fprintf(os.Stderr,
			"\n\033[31m[ JARVINX ]\033[0m Ollama est requis pour démarrer.\n"+
				"  Lance : ollama serve\n"+
				"  Puis  : ollama pull %s\n\n", cfg.Model)
		os.Exit(1)
	}

	// Modèle manquant — warning mais on continue
	// (l'utilisateur a peut-être un alias ou une version différente)
	if health.Error != "" {
		fmt.Printf("\033[33m[ WARN ]\033[0m %s\n", health.Error)
		fmt.Printf("\033[33m[ WARN ]\033[0m Le runtime va démarrer mais les décisions LLM pourraient échouer.\n")
	}

	// 3. Discord
	if cfg.DiscordWebhook == "" {
		fmt.Println("\033[33m[ WARN ]\033[0m DISCORD_WEBHOOK non défini — alertes Discord désactivées")
	} else {
		fmt.Println("\033[32m[ OK   ]\033[0m Discord webhook chargé")
	}

	rt := core.NewRuntime(cfg)
	rt.Start()
}
