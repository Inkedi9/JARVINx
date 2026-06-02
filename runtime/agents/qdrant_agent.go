package agents

import (
	"context"
	"fmt"
	"time"

	jxlog "github.com/Inkedi9/jarvinx/jxlog"
)

// QdrantAgent vectorise chaque décision LLM et interroge la mémoire sémantique.
// Actif seulement si JARVINX_QDRANT_URL est défini.
// Pattern N-1 : lit la décision du cycle précédent, l'embeddise, et stocke dans Qdrant.
// Les décisions similaires trouvées sont injectées dans AgentContext au cycle suivant.
type QdrantAgent struct {
	BaseAgent
	qdrantURL string
	ollamaURL string
}

func NewQdrantAgent(qdrantURL, ollamaURL string) *QdrantAgent {
	return &QdrantAgent{
		BaseAgent: NewBaseAgent("qdrant", 15*time.Second),
		qdrantURL: qdrantURL,
		ollamaURL: ollamaURL,
	}
}

func (a *QdrantAgent) Run(ctx context.Context, actx AgentContext) error {
	cycles := actx.State.LastCycles(1)
	if len(cycles) == 0 {
		jxlog.Debug("QDRANT AGENT", "aucune décision disponible — cycle ignoré")
		a.recordSuccess()
		return nil
	}

	last := cycles[0]
	text := fmt.Sprintf("[%s] %s. %s", last.Action, last.Analysis, last.Reason)

	// TODO(v1.8): embed text via Ollama + upsert dans Qdrant
	// TODO(v1.8): rechercher décisions similaires + stocker pour injection au prochain cycle
	jxlog.Debug("QDRANT AGENT", fmt.Sprintf(
		"cycle %d — texte à embedder : %.80s…", last.CycleNum, text,
	))

	a.recordSuccess()
	return nil
}
