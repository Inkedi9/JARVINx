package llm

import (
	"fmt"
	"strings"
)

func BuildSystemPrompt(cpuThreshold, ramThreshold, diskThreshold float64, goos string) string {
	var sb strings.Builder

	_, _ = fmt.Fprintf(&sb, `Tu es JARVINx, un agent de monitoring système autonome.
Tu reçois l'état actuel d'un système ainsi que son historique récent.
Tu retournes UNIQUEMENT un objet JSON valide.
Aucun texte avant ou après le JSON. Aucun markdown. Aucun backtick.

Format de réponse obligatoire :
{
  "analysis": "description courte de l'état et des tendances observées",
  "action": "log" | "alert" | "suggest" | "execute",
  "command": "commande à exécuter (seulement si action=execute)",
  "reason": "explication de ta décision basée sur les tendances",
  "confidence": ta certitude sur la décision, float entre 0.0 et 1.0
}

Rules:
- "log"     : system stable, no worrying trend
- "alert"   : critical threshold exceeded (CPU >%.0f%%, RAM >%.0f%%, Disk >%.0f%%)
- "suggest" : degraded trend over multiple cycles
- "execute" : diagnostic needed

Commands autorisées :
- "docker ps"
- "docker stats"
- "uptime"
- "df -h"
- "free -h"
`, cpuThreshold, ramThreshold, diskThreshold)

	if goos == "windows" {
		sb.WriteString(`
Note OS : Windows detected. Commands are auto-translated at runtime :
- "df -h"   → wmic logicaldisk get size,freespace,caption
- "free -h" → wmic OS get FreePhysicalMemory,TotalVisibleMemorySize
- "uptime"  → net statistics workstation
Utilise uniquement les noms listés ci-dessus, pas leurs équivalents Windows.
`)
	}

	sb.WriteString("\nAnalyse les TENDANCES, pas seulement l'instant présent.")
	return sb.String()
}

func BuildAdaptivePrompt(ctx SystemContext) string {
	base := BuildSystemPrompt(ctx.CPUThreshold, ctx.RAMThreshold, ctx.DiskThreshold, ctx.GOOS)

	adaptiveCtx := BuildAdaptiveContext(ctx.Cycles, ctx.History, ctx.CPUThreshold, ctx.RAMThreshold, ctx.DiskThreshold)
	adaptiveCtx.SimilarDecisions = ctx.SimilarDecisions
	return BuildAdaptiveSystemPrompt(base, adaptiveCtx)
}

func BuildUserPrompt(ctx SystemContext) string {
	var sb strings.Builder

	// Historique
	if len(ctx.History) > 0 {
		sb.WriteString("Historique des observations récentes :\n")
		for _, snap := range ctx.History {
			_, _ = fmt.Fprintf(&sb, "  %s → CPU: %.1f%% | RAM: %.1f%% | Disk: %.1f%%\n",
				snap.Timestamp.Format("15:04:05"),
				snap.CPUPercent,
				snap.MemPercent,
				snap.DiskPercent,
			)
		}
		sb.WriteString("\n")
	}

	_, _ = fmt.Fprintf(&sb,
		`Observation actuelle à %s :
- CPU    : %.1f%%
- RAM    : %d MB / %d MB (%.1f%%)
- DISQUE : %d GB / %d GB (%.1f%%)

Analyse les tendances et retourne ta décision JSON.`,
		ctx.Timestamp.Format("15:04:05"),
		ctx.CPUPercent,
		ctx.MemUsed, ctx.MemTotal, ctx.MemPercent,
		ctx.DiskUsed, ctx.DiskTotal, ctx.DiskPercent,
	)

	return sb.String()
}
