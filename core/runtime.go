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
	state  *memory.State
	agent  *agents.SystemAgent
}

func NewRuntime(cfg *config.Config) *Runtime {
	return &Runtime{
		cfg:    cfg,
		logger: memory.NewLogger(cfg.LogFile),
		state:  memory.NewState(cfg.StateFile),
		agent:  agents.NewSystemAgent(cfg.OllamaURL, cfg.Model),
	}
}

func (r *Runtime) Start() {
	fmt.Println("[ JARVINX ] Démarrage du runtime...")
	fmt.Printf("[ JARVINX ] Modèle     : %s\n", r.cfg.Model)
	fmt.Printf("[ JARVINX ] Intervalle : %v\n", r.cfg.Interval)
	fmt.Printf("[ JARVINX ] Historique : %d snapshots en mémoire\n\n",
		len(r.state.History))

	for {
		// 1. OBSERVE
		state, err := tools.Observe()
		if err != nil {
			fmt.Printf("[ ERREUR ] Observation : %v\n", err)
			time.Sleep(r.cfg.Interval)
			continue
		}
		state.Display()

		// 2. Construire le snapshot
		snap := memory.Snapshot{
			Timestamp:   state.Timestamp,
			CPUPercent:  state.CPUPercent,
			MemUsed:     state.MemUsed,
			MemTotal:    state.MemTotal,
			MemPercent:  state.MemPercent,
			DiskUsed:    state.DiskUsed,
			DiskTotal:   state.DiskTotal,
			DiskPercent: state.DiskPercent,
		}

		// 3. THINK — avec historique
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
			History:     r.state.Last(5),
		}

		decision, err := r.agent.Decide(ctx)
		if err != nil {
			fmt.Printf("[ ERREUR ] Agent : %v\n", err)
			// On sauvegarde quand même le snapshot
			r.state.Add(snap)
			r.state.Save()
			time.Sleep(r.cfg.Interval)
			continue
		}
		decision.Display()

		// 4. ACT
		if decision.Command != "" {
			fmt.Printf("[ EXEC ] Exécution : '%s'\n", decision.Command)
			result := tools.ExecuteCommand(decision.Command)
			result.Display()
		}

		// 5. MÉMORISER + LOG
		r.state.Add(snap)
		if err := r.state.Save(); err != nil {
			fmt.Printf("[ ERREUR ] State : %v\n", err)
		} else {
			fmt.Printf("[ STATE ] %d snapshots en mémoire\n", len(r.state.History))
		}

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
