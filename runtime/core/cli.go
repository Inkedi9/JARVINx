package core

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	jxlog "github.com/Inkedi9/jarvinx/jxlog"
	"github.com/Inkedi9/jarvinx/memory"
)

type CLI struct {
	state     *memory.State
	scheduler *Scheduler
}

func NewCLI(state *memory.State, scheduler *Scheduler) *CLI {
	return &CLI{
		state:     state,
		scheduler: scheduler,
	}
}

func (c *CLI) Start() {
	jxlog.Info("CLI", "Prêt — tape 'help' pour les commands")

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		cmd := parts[0]
		args := parts[1:]

		switch cmd {
		case "help":
			c.cmdHelp()
		case "status":
			c.cmdStatus()
		case "history":
			c.cmdHistory(args)
		case "interval":
			c.cmdInterval(args)
		case "clear":
			fmt.Print("\033[H\033[2J")
		default:
			jxlog.Warn("CLI", fmt.Sprintf("Commande inconnue : '%s' — tape 'help'", cmd))
		}
	}
}

func (c *CLI) cmdHelp() {
	fmt.Println()
	fmt.Println("  Commandes disponibles :")
	fmt.Println("  ─────────────────────────────────────────")
	fmt.Println("  status              État du dernier cycle")
	fmt.Println("  history [n]         Derniers N snapshots (défaut: 5)")
	fmt.Println("  interval <secondes> Change l'intervalle de tick")
	fmt.Println("  clear               Efface l'écran")
	fmt.Println("  help                Cette aide")
	fmt.Println("  ─────────────────────────────────────────")
	fmt.Println()
}

func (c *CLI) cmdStatus() {
	snapshots := c.state.Last(1)
	if len(snapshots) == 0 {
		jxlog.Warn("CLI", "Aucune observation disponible")
		return
	}

	s := snapshots[0]
	fmt.Println()
	fmt.Println("  ┌─[ DERNIER SNAPSHOT ]─────────────────────┐")
	fmt.Printf("  │ Timestamp : %s\n", s.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("  │ CPU       : %.1f%%\n", s.CPUPercent)
	fmt.Printf("  │ RAM       : %d MB / %d MB (%.1f%%)\n", s.MemUsed, s.MemTotal, s.MemPercent)
	fmt.Printf("  │ DISK      : %d GB / %d GB (%.1f%%)\n", s.DiskUsed, s.DiskTotal, s.DiskPercent)
	fmt.Println("  └──────────────────────────────────────────┘")
	fmt.Println()
}

func (c *CLI) cmdHistory(args []string) {
	n := 5
	if len(args) > 0 {
		parsed, err := strconv.Atoi(args[0])
		if err != nil || parsed < 1 {
			jxlog.Warn("CLI", "Usage : history [nombre]")
			return
		}
		n = parsed
	}

	snapshots := c.state.Last(n)
	if len(snapshots) == 0 {
		jxlog.Warn("CLI", "Aucun historique disponible")
		return
	}

	fmt.Printf("\n  Historique — %d derniers snapshots :\n", len(snapshots))
	fmt.Println("  ──────────────────────────────────────────────────────")
	fmt.Printf("  %-10s  %-8s  %-18s  %-12s\n", "Heure", "CPU", "RAM", "Disk")
	fmt.Println("  ──────────────────────────────────────────────────────")

	for _, s := range snapshots {
		ram := fmt.Sprintf("%d/%d MB", s.MemUsed, s.MemTotal)
		disk := fmt.Sprintf("%d/%d GB", s.DiskUsed, s.DiskTotal)
		fmt.Printf("  %-10s  %-7.1f%%  %-18s  %-12s\n",
			s.Timestamp.Format("15:04:05"),
			s.CPUPercent,
			ram,
			disk,
		)
	}
	fmt.Println()
}

func (c *CLI) cmdInterval(args []string) {
	if len(args) == 0 {
		fmt.Printf("[ CLI ] Intervalle actuel : %v\n", c.scheduler.interval)
		jxlog.Info("CLI", "Usage : interval <secondes>")
		return
	}

	secs, err := strconv.Atoi(args[0])
	if err != nil || secs < 5 {
		jxlog.Warn("CLI", "Minimum 5 secondes")
		return
	}

	c.scheduler.SetInterval(time.Duration(secs) * time.Second)
	jxlog.Info("CLI", fmt.Sprintf("Intervalle mis à jour : %ds", secs))
}
