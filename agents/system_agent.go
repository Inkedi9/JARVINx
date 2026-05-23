package agents

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Inkedi9/jarvinx/llm"
)

type Decision struct {
	Analysis string `json:"analysis"`
	Action   string `json:"action"`
	Command  string `json:"command,omitempty"`
	Reason   string `json:"reason"`
}

type SystemAgent struct {
	llm *llm.OllamaClient
}

func NewSystemAgent(baseURL, model string) *SystemAgent {
	return &SystemAgent{
		llm: llm.NewOllamaClient(baseURL, model),
	}
}

func (a *SystemAgent) Decide(ctx llm.SystemContext) (Decision, error) {
	systemPrompt := llm.BuildSystemPrompt()
	userPrompt := llm.BuildUserPrompt(ctx)

	raw, err := a.llm.Think(systemPrompt, userPrompt)
	if err != nil {
		return Decision{}, fmt.Errorf("llm think: %w", err)
	}

	// Nettoyage défensif — certains LLM ajoutent des backticks malgré les instructions
	cleaned := strings.TrimSpace(raw)
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)

	var decision Decision
	if err := json.Unmarshal([]byte(cleaned), &decision); err != nil {
		return Decision{}, fmt.Errorf("parse decision: %w\nraw response: %s", err, raw)
	}

	return decision, nil
}

func (d Decision) Display() {
	fmt.Printf("[ AGENT ] Action   : %s\n", d.Action)
	fmt.Printf("[ AGENT ] Analyse  : %s\n", d.Analysis)

	reason := d.Reason
	if reason == "" {
		reason = "—"
	}
	fmt.Printf("[ AGENT ] Raison   : %s\n", reason)

	if d.Command != "" {
		fmt.Printf("[ AGENT ] Commande : %s\n", d.Command)
	}
}
