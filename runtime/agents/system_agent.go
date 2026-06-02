package agents

import (
	"context"
	"fmt"
	"runtime"
	"time"

	jxlog "github.com/Inkedi9/jarvinx/jxlog"
	"github.com/Inkedi9/jarvinx/llm"
	"github.com/Inkedi9/jarvinx/memory"
)

type SystemAgent struct {
	BaseAgent
	client        *llm.OllamaClient
	retry         llm.RetryConfig
	cpuThreshold  float64
	ramThreshold  float64
	diskThreshold float64
}

func NewSystemAgent(baseURL, model string, cpuThreshold, ramThreshold, diskThreshold float64) *SystemAgent {
	return &SystemAgent{
		BaseAgent:     NewBaseAgent("system", 15*time.Second),
		client:        llm.NewOllamaClient(baseURL, model),
		retry:         llm.DefaultRetryConfig(),
		cpuThreshold:  cpuThreshold,
		ramThreshold:  ramThreshold,
		diskThreshold: diskThreshold,
	}
}

func (a *SystemAgent) Run(ctx context.Context, actx AgentContext) error {
	snap := actx.Snapshot

	callCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// Récupère l'historique des snapshots ET des décisions
	history := actx.State.Last(10)
	cycles := actx.State.LastCycles(20)

	llmCtx := llm.SystemContext{
		Timestamp:        snap.Timestamp,
		CPUPercent:       snap.CPUPercent,
		MemUsed:          snap.MemUsed,
		MemTotal:         snap.MemTotal,
		MemPercent:       snap.MemPercent,
		DiskUsed:         snap.DiskUsed,
		DiskTotal:        snap.DiskTotal,
		DiskPercent:      snap.DiskPercent,
		History:          history,
		Cycles:           cycles,
		CPUThreshold:     a.cpuThreshold,
		RAMThreshold:     a.ramThreshold,
		DiskThreshold:    a.diskThreshold,
		GOOS:             runtime.GOOS,
		SimilarDecisions: actx.SimilarDecisions,
	}

	// Prompt adaptatif — enrichi du contexte historique
	systemPrompt := llm.BuildAdaptivePrompt(llmCtx)
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
	record.Confidence = decision.Confidence
	if decision.Action == "execute" {
		record.TriggerCPU = snap.CPUPercent
		record.TriggerRAM = snap.MemPercent
		record.TriggerDisk = snap.DiskPercent
	}
	if storeErr := actx.State.AddCycle(record); storeErr != nil {
		jxlog.Error("SYSTEM AGENT", fmt.Sprintf("AddCycle : %v", storeErr))
	}
	if storeErr := actx.State.Add(snap); storeErr != nil {
		jxlog.Error("SYSTEM AGENT", fmt.Sprintf("Add : %v", storeErr))
	}

	if saveErr := actx.State.Save(); saveErr != nil {
		jxlog.Error("SYSTEM AGENT", fmt.Sprintf("State save : %v", saveErr))
	}

	jxlog.Info("STATE", fmt.Sprintf("Cycle #%d enregistré", actx.State.CurrentCycle()))

	return err
}
