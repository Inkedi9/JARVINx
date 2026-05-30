package llm

import (
	"fmt"
	"strings"
	"time"

	"github.com/Inkedi9/jarvinx/memory"
)

// AdaptiveContext contient les insights tirés de l'historique
type AdaptiveContext struct {
	DominantAction string
	AlertRate      float64 // % de cycles en alerte
	CPUTrend       string  // "stable", "rising", "high"
	RAMTrend       string
	DiskTrend      string
	RecentAlerts   []string
	CycleCount     int
}

// BuildAdaptiveContext analyse les derniers cycles et retourne un contexte enrichi
func BuildAdaptiveContext(cycles []memory.CycleRecord, snapshots []memory.Snapshot) AdaptiveContext {
	ctx := AdaptiveContext{}

	// Analyse des cycles
	if len(cycles) > 0 {
		ctx.CycleCount = len(cycles)

		// Action dominante
		counts := make(map[string]int)
		for _, c := range cycles {
			counts[c.Action]++
		}
		ctx.DominantAction = dominantKey(counts)

		// Taux d'alerte
		ctx.AlertRate = float64(counts["alert"]) / float64(len(cycles)) * 100

		// Alertes récentes
		for _, c := range cycles {
			if c.Action == "alert" && len(ctx.RecentAlerts) < 3 {
				ctx.RecentAlerts = append(ctx.RecentAlerts, c.Analysis)
			}
		}
	}

	// Analyse des snapshots — indépendante des cycles
	if len(snapshots) >= 3 {
		ctx.CPUTrend = trend(snapshots, func(s memory.Snapshot) float64 { return s.CPUPercent })
		ctx.RAMTrend = trend(snapshots, func(s memory.Snapshot) float64 { return s.MemPercent })
		ctx.DiskTrend = trend(snapshots, func(s memory.Snapshot) float64 { return s.DiskPercent })
	}

	return ctx
}

// BuildAdaptiveSystemPrompt construit un system prompt enrichi du contexte
func BuildAdaptiveSystemPrompt(base string, ctx AdaptiveContext) string {
	if ctx.CycleCount == 0 {
		return base
	}

	var sb strings.Builder
	sb.WriteString(base)
	sb.WriteString("\n\n--- CONTEXTE ADAPTATIF ---\n")
	sb.WriteString(fmt.Sprintf("Analyse basée sur %d cycles récents.\n", ctx.CycleCount))

	// Tendance dominante
	if ctx.DominantAction != "" {
		sb.WriteString(fmt.Sprintf("Action dominante : %s\n", ctx.DominantAction))
	}

	// Taux d'alerte
	if ctx.AlertRate > 20 {
		sb.WriteString(fmt.Sprintf(
			"⚠ Taux d'alerte élevé : %.0f%% des cycles — sois particulièrement vigilant.\n",
			ctx.AlertRate,
		))
	} else if ctx.AlertRate == 0 {
		sb.WriteString("✓ Système stable — aucune alerte récente.\n")
	}

	// Tendances métriques
	sb.WriteString("\nTendances observées :\n")
	sb.WriteString(fmt.Sprintf("  CPU  : %s\n", ctx.CPUTrend))
	sb.WriteString(fmt.Sprintf("  RAM  : %s\n", ctx.RAMTrend))
	sb.WriteString(fmt.Sprintf("  Disk : %s\n", ctx.DiskTrend))

	// Alertes récentes
	if len(ctx.RecentAlerts) > 0 {
		sb.WriteString("\nDernières alertes :\n")
		for _, a := range ctx.RecentAlerts {
			sb.WriteString(fmt.Sprintf("  - %s\n", a))
		}
	}

	sb.WriteString("--- FIN CONTEXTE ---\n")
	sb.WriteString("Utilise ce contexte pour affiner ton analyse.")

	return sb.String()
}

// trend calcule la tendance d'une métrique sur une série de snapshots
func trend(snapshots []memory.Snapshot, getter func(memory.Snapshot) float64) string {
	if len(snapshots) < 3 {
		return "insufficient data"
	}

	first := average(snapshots[:len(snapshots)/2], getter)
	last := average(snapshots[len(snapshots)/2:], getter)
	diff := last - first

	current := getter(snapshots[len(snapshots)-1])

	switch {
	case current >= 85:
		return fmt.Sprintf("critique (%.1f%%) — action requise", current)
	case current >= 70:
		return fmt.Sprintf("élevé (%.1f%%)", current)
	case diff > 10:
		return fmt.Sprintf("en hausse (%.1f%% → %.1f%%)", first, last)
	case diff < -10:
		return fmt.Sprintf("en baisse (%.1f%% → %.1f%%)", first, last)
	default:
		return fmt.Sprintf("stable (%.1f%%)", current)
	}
}

func average(snapshots []memory.Snapshot, getter func(memory.Snapshot) float64) float64 {
	if len(snapshots) == 0 {
		return 0
	}
	sum := 0.0
	for _, s := range snapshots {
		sum += getter(s)
	}
	return sum / float64(len(snapshots))
}

func dominantKey(counts map[string]int) string {
	best := ""
	max := 0
	for k, v := range counts {
		if v > max {
			max = v
			best = k
		}
	}
	return best
}

// SystemContext — ajoute le timestamp pour le contexte adaptatif
type SystemContext struct {
	Timestamp   time.Time
	CPUPercent  float64
	MemUsed     uint64
	MemTotal    uint64
	MemPercent  float64
	DiskUsed    uint64
	DiskTotal   uint64
	DiskPercent float64
	History     []memory.Snapshot
	Cycles      []memory.CycleRecord // nouveau — historique des décisions
}
