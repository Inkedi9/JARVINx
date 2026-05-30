package llm

import (
	"fmt"
	"strings"
)

func BuildSystemPrompt() string {
	return `Tu es JARVINx, un agent de monitoring système autonome.
Tu reçois l'état actuel d'un système ainsi que son historique récent.
Tu retournes UNIQUEMENT un objet JSON valide.
Aucun texte avant ou après le JSON. Aucun markdown. Aucun backtick.

Format de réponse obligatoire :
{
  "analysis": "description courte de l'état et des tendances observées",
  "action": "log" | "alert" | "suggest" | "execute",
  "command": "commande à exécuter (seulement si action=execute)",
  "reason": "explication de ta décision basée sur les tendances"
}

Rules:
- "log"     : system stable, no worrying trend
- "alert"   : critical threshold exceeded (CPU >85%, RAM >90%, Disk >90%)
- "suggest" : degraded trend over multiple cycles
- "execute" : diagnostic needed

Commands autorisées :
- "docker ps"
- "docker stats"
- "uptime"
- "df -h"
- "free -h"

Analyse les TENDANCES, pas seulement l'instant présent.`
}

func BuildAdaptivePrompt(ctx SystemContext) string {
	base := BuildSystemPrompt()

	// Construit le contexte adaptatif depuis l'historique
	adaptiveCtx := BuildAdaptiveContext(ctx.Cycles, ctx.History)
	return BuildAdaptiveSystemPrompt(base, adaptiveCtx)
}

func BuildUserPrompt(ctx SystemContext) string {
	var sb strings.Builder

	// Historique
	if len(ctx.History) > 0 {
		sb.WriteString("Historique des observations récentes :\n")
		for _, snap := range ctx.History {
			sb.WriteString(fmt.Sprintf("  %s → CPU: %.1f%% | RAM: %.1f%% | Disk: %.1f%%\n",
				snap.Timestamp.Format("15:04:05"),
				snap.CPUPercent,
				snap.MemPercent,
				snap.DiskPercent,
			))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf(
		`Observation actuelle à %s :
- CPU    : %.1f%%
- RAM    : %d MB / %d MB (%.1f%%)
- DISQUE : %d GB / %d GB (%.1f%%)

Analyse les tendances et retourne ta décision JSON.`,
		ctx.Timestamp.Format("15:04:05"),
		ctx.CPUPercent,
		ctx.MemUsed, ctx.MemTotal, ctx.MemPercent,
		ctx.DiskUsed, ctx.DiskTotal, ctx.DiskPercent,
	))

	return sb.String()
}
