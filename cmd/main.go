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

	if cfg.DiscordWebhook == "" {
		fmt.Println("[ WARN ] DISCORD_WEBHOOK non défini — alertes Discord désactivées")
	} else {
		fmt.Println("[ OK   ] Discord webhook chargé")
	}

	rt := core.NewRuntime(cfg)
	rt.Start()
}
