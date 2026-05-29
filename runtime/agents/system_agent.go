package agents

import (
	"context"
	"fmt"
	"time"

	jxlog "github.com/Inkedi9/jarvinx/jxlog"
	"github.com/Inkedi9/jarvinx/llm"
	"github.com/Inkedi9/jarvinx/memory"
)

type SystemAgent struct {
	BaseAgent
	client *llm.OllamaClient
	retry  llm.RetryConfig
}

func NewSystemAgent(baseURL, model string) *SystemAgent {
	return &SystemAgent{
		BaseAgent: NewBaseAgent("system", 15*time.Second),
		client:    llm.NewOllamaClient(baseURL, model),
		retry:     llm.DefaultRetryConfig(),
	}
}

func (a *SystemAgent) Run(ctx context.Context, actx AgentContext) error {
	snap := actx.Snapshot

	callCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	history := actx.State.Last(5)
	llmCtx := llm.SystemContext{
		Timestamp:   snap.Timestamp,
		CPUPercent:  snap.CPUPercent,
		MemUsed:     snap.MemUsed,
		MemTotal:    snap.MemTotal,
		MemPercent:  snap.MemPercent,
		DiskUsed:    snap.DiskUsed,
		DiskTotal:   snap.DiskTotal,
		DiskPercent: snap.DiskPercent,
		History:     history,
	}

	systemPrompt := llm.BuildSystemPrompt()
	userPrompt := llm.BuildUserPrompt(llmCtx)

	jxlog.Info("SYSTEM AGENT", "Analyse en cours...")

	decision, attempts, err := a.client.ThinkWithDecision(
		callCtx,
		systemPrompt,
		userPrompt,
		a.retry,
	)

	if err != nil {
		a.recordError(err)
		jxlog.Warn("SYSTEM AGENT", fmt.Sprintf("Fallback après %d tentatives", attempts))
	} else {
		a.recordSuccess()
	}

	decision.Display()

	record := memory.NewCycleRecord(
		snap,
		decision.Action,
		decision.Analysis,
		decision.Reason,
		decision.Command,
	)
	actx.State.AddCycle(record)
	actx.State.Add(snap)

	if err := actx.State.Save(); err != nil {
		jxlog.Error("SYSTEM AGENT", fmt.Sprintf("State save : %v", err))
	}

	jxlog.Info("STATE", fmt.Sprintf("Cycle #%d enregistré", actx.State.CycleNum))

	return err
}
