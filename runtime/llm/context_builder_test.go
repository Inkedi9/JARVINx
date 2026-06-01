package llm

import (
	"strings"
	"testing"
	"time"

	"github.com/Inkedi9/jarvinx/memory"
)

func makeSnap(cpu, ram, disk float64) memory.Snapshot {
	return memory.Snapshot{
		Timestamp:   time.Now(),
		CPUPercent:  cpu,
		MemPercent:  ram,
		DiskPercent: disk,
		MemTotal:    16000,
		DiskTotal:   500,
	}
}

func makeCycle(action string) memory.CycleRecord {
	return memory.NewCycleRecord(
		makeSnap(20, 50, 60),
		action, "test", "test", "",
	)
}

func TestBuildAdaptiveContext_Empty(t *testing.T) {
	ctx := BuildAdaptiveContext(nil, nil, 85, 90, 85)
	if ctx.CycleCount != 0 {
		t.Errorf("expected 0 cycles, got %d", ctx.CycleCount)
	}
}

func TestBuildAdaptiveContext_DominantAction(t *testing.T) {
	cycles := []memory.CycleRecord{
		makeCycle("log"),
		makeCycle("log"),
		makeCycle("log"),
		makeCycle("alert"),
	}

	ctx := BuildAdaptiveContext(cycles, nil, 85, 90, 85)
	if ctx.DominantAction != "log" {
		t.Errorf("expected dominant 'log', got '%s'", ctx.DominantAction)
	}
}

func TestBuildAdaptiveContext_AlertRate(t *testing.T) {
	cycles := []memory.CycleRecord{
		makeCycle("alert"),
		makeCycle("alert"),
		makeCycle("log"),
		makeCycle("log"),
	}

	ctx := BuildAdaptiveContext(cycles, nil, 85, 90, 85)
	if ctx.AlertRate != 50.0 {
		t.Errorf("expected alert rate 50%%, got %.1f%%", ctx.AlertRate)
	}
}

func TestBuildAdaptiveContext_RecentAlerts(t *testing.T) {
	cycles := []memory.CycleRecord{
		memory.NewCycleRecord(makeSnap(90, 50, 60), "alert", "CPU critique", "", ""),
		memory.NewCycleRecord(makeSnap(20, 50, 60), "log", "stable", "", ""),
	}

	ctx := BuildAdaptiveContext(cycles, nil, 85, 90, 85)
	if len(ctx.RecentAlerts) == 0 {
		t.Error("expected recent alerts")
	}
	if ctx.RecentAlerts[0] != "CPU critique" {
		t.Errorf("expected 'CPU critique', got '%s'", ctx.RecentAlerts[0])
	}
}

func TestBuildAdaptiveContext_CPUTrend(t *testing.T) {
	snaps := []memory.Snapshot{
		makeSnap(20, 50, 60),
		makeSnap(22, 51, 60),
		makeSnap(85, 52, 60), // spike
	}

	ctx := BuildAdaptiveContext(nil, snaps, 85, 90, 85)
	if !strings.Contains(ctx.CPUTrend, "critique") {
		t.Errorf("expected 'critique' in CPU trend for 85%%, got: %s", ctx.CPUTrend)
	}
}

func TestBuildAdaptiveSystemPrompt_Empty(t *testing.T) {
	base := "base prompt"
	result := BuildAdaptiveSystemPrompt(base, AdaptiveContext{})

	// Sans données, retourne le prompt de base
	if result != base {
		t.Errorf("expected base prompt unchanged, got: %s", result)
	}
}

func TestBuildAdaptiveSystemPrompt_WithAlerts(t *testing.T) {
	base := "base prompt"
	ctx := AdaptiveContext{
		CycleCount:     10,
		AlertRate:      50.0,
		DominantAction: "alert",
		CPUTrend:       "stable (25.0%)",
		RAMTrend:       "stable (55.0%)",
		DiskTrend:      "stable (60.0%)",
	}

	result := BuildAdaptiveSystemPrompt(base, ctx)

	if !strings.Contains(result, "base prompt") {
		t.Error("expected base prompt in result")
	}
	if !strings.Contains(result, "CONTEXTE ADAPTATIF") {
		t.Error("expected adaptive context section")
	}
	if !strings.Contains(result, "50%") {
		t.Error("expected alert rate in prompt")
	}
}

func TestBuildAdaptiveSystemPrompt_Stable(t *testing.T) {
	base := "base"
	ctx := AdaptiveContext{
		CycleCount:     20,
		AlertRate:      0,
		DominantAction: "log",
		CPUTrend:       "stable (20.0%)",
		RAMTrend:       "stable (50.0%)",
		DiskTrend:      "stable (60.0%)",
	}

	result := BuildAdaptiveSystemPrompt(base, ctx)
	if !strings.Contains(result, "stable") {
		t.Error("expected stability mention for 0% alert rate")
	}
}

func TestTrend_Critical(t *testing.T) {
	snaps := []memory.Snapshot{
		makeSnap(85, 50, 60),
		makeSnap(87, 50, 60),
		makeSnap(90, 50, 60),
	}

	result := trendWithThreshold(snaps, func(s memory.Snapshot) float64 { return s.CPUPercent }, 85.0)
	if !strings.Contains(result, "critique") {
		t.Errorf("expected 'critique' for 90%% CPU, got: %s", result)
	}
}

func TestTrend_Rising(t *testing.T) {
	snaps := []memory.Snapshot{
		makeSnap(10, 50, 60),
		makeSnap(15, 50, 60),
		makeSnap(40, 50, 60), // diff > 10 entre les deux moitiés
	}

	result := trendWithThreshold(snaps, func(s memory.Snapshot) float64 { return s.CPUPercent }, 85.0)
	if !strings.Contains(result, "hausse") {
		t.Errorf("expected 'hausse' for rising CPU, got: %s", result)
	}
}

func TestTrend_Stable(t *testing.T) {
	snaps := []memory.Snapshot{
		makeSnap(25, 50, 60),
		makeSnap(26, 50, 60),
		makeSnap(25, 50, 60),
	}

	result := trendWithThreshold(snaps, func(s memory.Snapshot) float64 { return s.CPUPercent }, 85.0)
	if !strings.Contains(result, "stable") {
		t.Errorf("expected 'stable', got: %s", result)
	}
}

func TestDominantKey(t *testing.T) {
	counts := map[string]int{
		"log":     10,
		"alert":   3,
		"suggest": 5,
	}

	if dominantKey(counts) != "log" {
		t.Errorf("expected 'log' as dominant, got '%s'", dominantKey(counts))
	}
}
