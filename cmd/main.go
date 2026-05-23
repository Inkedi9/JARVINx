package main

import (
	"github.com/Inkedi9/jarvinx/config"
	"github.com/Inkedi9/jarvinx/core"
)

// la fonction principale. C'est ici que tout commence quand tu lances le binaire.
func main() {
	cfg := config.Default()
	rt := core.NewRuntime(cfg)
	rt.Start()
}
