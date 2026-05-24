package agents

import (
	"context"
	"fmt"
	"time"

	"github.com/Inkedi9/jarvinx/llm"
)

// Decision est réexportée depuis llm pour garder l'API propre
type Decision = llm.Decision

type SystemAgent struct {
	client *llm.OllamaClient
	retry  llm.RetryConfig
}

func NewSystemAgent(baseURL, model string) *SystemAgent {
	return &SystemAgent{
		client: llm.NewOllamaClient(baseURL, model),
		retry:  llm.DefaultRetryConfig(),
	}
}

func (a *SystemAgent) Decide(ctx llm.SystemContext) (Decision, error) {
	// Timeout par cycle — 60s max pour une décision
	callCtx, cancel := context.WithTimeout(
		context.Background(),
		60*time.Second,
	)
	defer cancel()

	systemPrompt := llm.BuildSystemPrompt()
	userPrompt := llm.BuildUserPrompt(ctx)

	decision, attempts, err := a.client.ThinkWithDecision(
		callCtx,
		systemPrompt,
		userPrompt,
		a.retry,
	)

	if err != nil {
		fmt.Printf("[ AGENT ] Décision par fallback après %d tentatives\n", attempts)
	}

	// On retourne toujours quelque chose — jamais de nil
	return decision, err
}
