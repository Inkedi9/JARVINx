package core

import (
	"fmt"
	"time"

	"github.com/Inkedi9/jarvinx/agents"
	"github.com/Inkedi9/jarvinx/config"
	"github.com/Inkedi9/jarvinx/llm"
	"github.com/Inkedi9/jarvinx/memory"
	"github.com/Inkedi9/jarvinx/tools"
)

type Runtime struct {
	cfg    *config.Config
	logger *memory.Logger
	agent  *agents.SystemAgent
}

// Go gère le temps avec des types explicites. 5 * time.Second est plus lisible et moins dangereux que juste 5
func NewRuntime(cfg *config.Config) *Runtime {
	return &Runtime{
		cfg:    cfg,
		logger: memory.NewLogger(cfg.LogFile),
		agent:  agents.NewSystemAgent(cfg.OllamaURL, cfg.Model),
	}
}

// Au lieu de copier la struct en mémoire à chaque appel, tu passes son adresse.
// en Go, * = pointeur, & = adresse.
func (r *Runtime) Start() {
	fmt.Println("[ JARVINX ] Démarrage du runtime...")
	fmt.Printf("[ JARVINX ] Modèle     : %s\n", r.cfg.Model)
	fmt.Printf("[ JARVINX ] Intervalle : %v\n\n", r.cfg.Interval)

	for {
		// 1. OBSERVE
		state, err := tools.Observe()
		if err != nil {
			fmt.Printf("[ ERREUR ] Observation : %v\n", err)
			time.Sleep(r.cfg.Interval)
			continue
		}
		state.Display()

		// 2. THINK
		fmt.Println("[ AGENT ] Analyse en cours...")
		ctx := llm.SystemContext{
			Timestamp:   state.Timestamp,
			CPUPercent:  state.CPUPercent,
			MemUsed:     state.MemUsed,
			MemTotal:    state.MemTotal,
			MemPercent:  state.MemPercent,
			DiskUsed:    state.DiskUsed,
			DiskTotal:   state.DiskTotal,
			DiskPercent: state.DiskPercent,
		}

		decision, err := r.agent.Decide(ctx)
		if err != nil {
			fmt.Printf("[ ERREUR ] Agent : %v\n", err)
			time.Sleep(r.cfg.Interval)
			continue
		}
		decision.Display()

		// 3. ACT
		if decision.Action == "execute" && decision.Command != "" {
			fmt.Printf("[ EXEC ] Exécution de : '%s'\n", decision.Command)
			result := tools.ExecuteCommand(decision.Command)
			result.Display()
		}

		// 4. LOG
		entry := memory.LogEntry{
			Timestamp:   state.Timestamp,
			CPUPercent:  state.CPUPercent,
			MemUsed:     state.MemUsed,
			MemTotal:    state.MemTotal,
			MemPercent:  state.MemPercent,
			DiskUsed:    state.DiskUsed,
			DiskTotal:   state.DiskTotal,
			DiskPercent: state.DiskPercent,
		}
		if err := r.logger.Write(entry); err != nil {
			fmt.Printf("[ ERREUR ] Log : %v\n", err)
		}

		fmt.Println()
		time.Sleep(r.cfg.Interval)
	}
}
