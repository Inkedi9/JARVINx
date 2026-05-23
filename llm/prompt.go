package llm

import (
	"fmt"
	"time"
)

type SystemContext struct {
	Timestamp   time.Time
	CPUPercent  float64
	MemUsed     uint64
	MemTotal    uint64
	MemPercent  float64
	DiskUsed    uint64
	DiskTotal   uint64
	DiskPercent float64
}

func BuildSystemPrompt() string {
	return `Tu es JARVINx, un agent de monitoring système autonome.
Tu reçois l'état d'un système et tu retournes UNIQUEMENT un objet JSON valide.
Aucun texte avant ou après le JSON. Aucun markdown. Aucun backtick.

Format de réponse obligatoire :
{
  "analysis": "description courte de l'état système",
  "action": "log" | "alert" | "suggest",
  "reason": "explication de ta décision"
}

Règles :
- "log"     : tout va bien, on enregistre
- "alert"   : quelque chose dépasse un seuil critique (CPU >85%, RAM >90%, Disk >90%)
- "suggest" : situation dégradée mais pas critique, tu proposes une action

Commandes autorisées uniquement :
- "docker ps"
- "docker stats"  
- "uptime"
- "df -h"
- "free -h"

Utilise "execute" avec parcimonie, seulement si c'est vraiment utile.`
}

func BuildUserPrompt(ctx SystemContext) string {
	return fmt.Sprintf(`État système observé à %s :
- CPU    : %.1f%%
- RAM    : %d MB utilisés / %d MB total (%.1f%%)
- DISQUE : %d GB utilisés / %d GB total (%.1f%%)

Analyse cet état et retourne ta décision JSON.`,
		ctx.Timestamp.Format("15:04:05"),
		ctx.CPUPercent,
		ctx.MemUsed, ctx.MemTotal, ctx.MemPercent,
		ctx.DiskUsed, ctx.DiskTotal, ctx.DiskPercent,
	)
}
