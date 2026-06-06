package llm

import (
	"fmt"
	"strings"
	"time"

	"github.com/Inkedi9/jarvinx/memory"
)

// AdaptiveContext contient les insights tirés de l'historique
type AdaptiveContext struct {
	DominantAction   string
	AlertRate        float64 // % de cycles en alerte
	CPUTrend         string  // "stable", "rising", "high"
	RAMTrend         string
	DiskTrend        string
	RecentAlerts     []string
	CycleCount       int
	SimilarDecisions []string // décisions passées similaires issues de Qdrant (v1.8)

	// v1.10 — prompt adaptatif enrichi
	TimeOfDay    string // "nuit", "matin", "journée", "soirée"
	IsWeekend    bool
	Correlation  string // diagnostic pré-calculé CPU/RAM
	StableStreak int    // cycles consécutifs sans alerte
	CPUForecast  string // "~N cycles avant seuil" ou ""
	RAMForecast  string
	DiskForecast string
}

// BuildAdaptiveContext analyse les derniers cycles et retourne un contexte enrichi
func BuildAdaptiveContext(cycles []memory.CycleRecord, snapshots []memory.Snapshot, cpuT, ramT, diskT float64) AdaptiveContext {
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
		ctx.CPUTrend = trendWithThreshold(snapshots, func(s memory.Snapshot) float64 { return s.CPUPercent }, cpuT)
		ctx.RAMTrend = trendWithThreshold(snapshots, func(s memory.Snapshot) float64 { return s.MemPercent }, ramT)
		ctx.DiskTrend = trendWithThreshold(snapshots, func(s memory.Snapshot) float64 { return s.DiskPercent }, diskT)
		ctx.Correlation = computeCorrelation(snapshots)
		ctx.CPUForecast = forecastCycles(snapshots, func(s memory.Snapshot) float64 { return s.CPUPercent }, cpuT)
		ctx.RAMForecast = forecastCycles(snapshots, func(s memory.Snapshot) float64 { return s.MemPercent }, ramT)
		ctx.DiskForecast = forecastCycles(snapshots, func(s memory.Snapshot) float64 { return s.DiskPercent }, diskT)
	}

	// Heure + période (contexte temporel)
	now := time.Now()
	ctx.TimeOfDay = timePeriod(now)
	ctx.IsWeekend = now.Weekday() == time.Saturday || now.Weekday() == time.Sunday

	// Streak de stabilité
	ctx.StableStreak = stableStreak(cycles)

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
	_, _ = fmt.Fprintf(&sb, "Analyse basée sur %d cycles récents.\n", ctx.CycleCount)

	// Tendance dominante
	if ctx.DominantAction != "" {
		_, _ = fmt.Fprintf(&sb, "Action dominante : %s\n", ctx.DominantAction)
	}

	// Taux d'alerte
	if ctx.AlertRate > 20 {
		_, _ = fmt.Fprintf(&sb,
			"⚠ Taux d'alerte élevé : %.0f%% des cycles — sois particulièrement vigilant.\n",
			ctx.AlertRate,
		)
	} else if ctx.AlertRate == 0 {
		sb.WriteString("✓ Système stable — aucune alerte récente.\n")
	}

	// Contexte temporel
	weekendStr := ""
	if ctx.IsWeekend {
		weekendStr = ", weekend"
	}
	_, _ = fmt.Fprintf(&sb, "Période : %s%s\n", ctx.TimeOfDay, weekendStr)

	// Tendances métriques
	sb.WriteString("\nTendances observées :\n")
	_, _ = fmt.Fprintf(&sb, "  CPU  : %s\n", ctx.CPUTrend)
	_, _ = fmt.Fprintf(&sb, "  RAM  : %s\n", ctx.RAMTrend)
	_, _ = fmt.Fprintf(&sb, "  Disk : %s\n", ctx.DiskTrend)

	// Corrélation CPU/RAM
	if ctx.Correlation != "" {
		_, _ = fmt.Fprintf(&sb, "  Corrélation : %s\n", ctx.Correlation)
	}

	// Forecast vers seuil
	var forecasts []string
	if ctx.CPUForecast != "" {
		forecasts = append(forecasts, "CPU "+ctx.CPUForecast)
	}
	if ctx.RAMForecast != "" {
		forecasts = append(forecasts, "RAM "+ctx.RAMForecast)
	}
	if ctx.DiskForecast != "" {
		forecasts = append(forecasts, "Disk "+ctx.DiskForecast)
	}
	if len(forecasts) > 0 {
		_, _ = fmt.Fprintf(&sb, "\nProjection : %s\n", strings.Join(forecasts, " | "))
	}

	// Streak de stabilité
	if ctx.StableStreak >= 3 {
		_, _ = fmt.Fprintf(&sb, "Stabilité : %d cycles consécutifs sans alerte.\n", ctx.StableStreak)
	}

	// Alertes récentes
	if len(ctx.RecentAlerts) > 0 {
		sb.WriteString("\nDernières alertes :\n")
		for _, a := range ctx.RecentAlerts {
			_, _ = fmt.Fprintf(&sb, "  - %s\n", a)
		}
	}

	// Décisions similaires passées (injectées par QdrantAgent en v1.8)
	if len(ctx.SimilarDecisions) > 0 {
		sb.WriteString("\n[HISTORICAL DATA]\n")
		for _, d := range ctx.SimilarDecisions {
			_, _ = fmt.Fprintf(&sb, "  - %s\n", sanitizeSimilarDecision(d))
		}
		sb.WriteString("[/HISTORICAL DATA]\n")
	}

	sb.WriteString("--- FIN CONTEXTE ---\n")
	sb.WriteString("Utilise ce contexte pour affiner ton analyse.")

	return sb.String()
}

// trendWithThreshold calcule la tendance d'une métrique par rapport au seuil configuré
func trendWithThreshold(snapshots []memory.Snapshot, getter func(memory.Snapshot) float64, threshold float64) string {
	if len(snapshots) < 3 {
		return "insufficient data"
	}

	first := average(snapshots[:len(snapshots)/2], getter)
	last := average(snapshots[len(snapshots)/2:], getter)
	diff := last - first

	current := getter(snapshots[len(snapshots)-1])

	switch {
	case current >= threshold:
		return fmt.Sprintf("critique (%.1f%%) — action requise", current)
	case current >= threshold*0.85:
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

// sanitizeSimilarDecision strips newlines and truncates to 200 chars to prevent prompt injection.
func sanitizeSimilarDecision(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	if len(s) > 200 {
		s = s[:200]
	}
	return s
}

// timePeriod retourne la période de la journée pour contextualiser les décisions.
func timePeriod(t time.Time) string {
	switch h := t.Hour(); {
	case h < 6:
		return "nuit"
	case h < 12:
		return "matin"
	case h < 18:
		return "journée"
	default:
		return "soirée"
	}
}

// computeCorrelation produit un diagnostic pré-calculé CPU/RAM.
func computeCorrelation(snapshots []memory.Snapshot) string {
	if len(snapshots) < 3 {
		return ""
	}
	cpuSlope := slopeOf(snapshots, func(s memory.Snapshot) float64 { return s.CPUPercent })
	ramSlope := slopeOf(snapshots, func(s memory.Snapshot) float64 { return s.MemPercent })

	const threshold = 2.0 // % par snapshot pour être considéré "en hausse"
	cpuRising := cpuSlope > threshold
	ramRising := ramSlope > threshold

	switch {
	case ramRising && !cpuRising:
		return "RAM↑ CPU stable → memory leak potentiel"
	case cpuRising && !ramRising:
		return "CPU↑ RAM stable → surge de traitement"
	case cpuRising && ramRising:
		return "CPU↑ RAM↑ → surcharge générale"
	default:
		return ""
	}
}

func slopeOf(snapshots []memory.Snapshot, getter func(memory.Snapshot) float64) float64 {
	if len(snapshots) < 2 {
		return 0
	}
	return (getter(snapshots[len(snapshots)-1]) - getter(snapshots[0])) / float64(len(snapshots)-1)
}

// stableStreak compte les cycles consécutifs sans alerte depuis le plus récent.
func stableStreak(cycles []memory.CycleRecord) int {
	streak := 0
	for i := len(cycles) - 1; i >= 0; i-- {
		if cycles[i].Action == "alert" {
			break
		}
		streak++
	}
	return streak
}

// forecastCycles projette le nombre de cycles avant d'atteindre le seuil.
// Retourne "" si la tendance est stable/descendante ou trop lointaine (>100 cycles).
func forecastCycles(snapshots []memory.Snapshot, getter func(memory.Snapshot) float64, threshold float64) string {
	if len(snapshots) < 3 {
		return ""
	}
	current := getter(snapshots[len(snapshots)-1])
	if current >= threshold {
		return ""
	}
	delta := slopeOf(snapshots, getter)
	if delta <= 0 {
		return ""
	}
	remaining := (threshold - current) / delta
	if remaining > 100 {
		return ""
	}
	return fmt.Sprintf("~%.0f cycles avant seuil", remaining)
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
	Cycles      []memory.CycleRecord

	// Seuils configurés — synchronisent l'analyse LLM avec la config réelle
	CPUThreshold  float64
	RAMThreshold  float64
	DiskThreshold float64
	GOOS          string

	// Décisions similaires passées — rempli par QdrantAgent (v1.8), nil sinon
	SimilarDecisions []string
}
