package main

import (
	"fmt"
	"os"

	"github.com/Inkedi9/jarvinx/config"
	"github.com/Inkedi9/jarvinx/core"
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

	if cfg.DiscordWebhook == "" {
		fmt.Println("\033[33m[ WARN ]\033[0m DISCORD_WEBHOOK non défini — alertes Discord désactivées")
	} else {
		fmt.Println("\033[32m[ OK   ]\033[0m Discord webhook chargé")
	}

	rt := core.NewRuntime(cfg)
	rt.Start()
}
