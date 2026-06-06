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

func TestBuildAdaptiveSystemPrompt_SimilarDecisionsWrapped(t *testing.T) {
	ctx := AdaptiveContext{
		CycleCount:       1,
		SimilarDecisions: []string{"CPU was high, action: suggest"},
	}
	result := BuildAdaptiveSystemPrompt("base", ctx)
	if !strings.Contains(result, "[HISTORICAL DATA]") {
		t.Error("expected [HISTORICAL DATA] opening tag")
	}
	if !strings.Contains(result, "[/HISTORICAL DATA]") {
		t.Error("expected [/HISTORICAL DATA] closing tag")
	}
	if !strings.Contains(result, "CPU was high") {
		t.Error("expected decision content inside block")
	}
}

func TestSanitizeSimilarDecision_StripNewlines(t *testing.T) {
	input := "line1\nline2\r\nline3"
	result := sanitizeSimilarDecision(input)
	if strings.ContainsAny(result, "\n\r") {
		t.Errorf("expected no newlines, got: %q", result)
	}
}

func TestSanitizeSimilarDecision_Truncate(t *testing.T) {
	input := strings.Repeat("a", 300)
	result := sanitizeSimilarDecision(input)
	if len(result) != 200 {
		t.Errorf("expected length 200, got %d", len(result))
	}
}

func TestTimePeriod(t *testing.T) {
	cases := []struct {
		hour int
		want string
	}{
		{0, "nuit"}, {3, "nuit"}, {5, "nuit"},
		{6, "matin"}, {11, "matin"},
		{12, "journée"}, {17, "journée"},
		{18, "soirée"}, {23, "soirée"},
	}
	for _, tc := range cases {
		ts := time.Date(2025, 1, 6, tc.hour, 0, 0, 0, time.UTC) // Monday
		if got := timePeriod(ts); got != tc.want {
			t.Errorf("hour %d: want %q, got %q", tc.hour, tc.want, got)
		}
	}
}

func TestStableStreak_NoAlert(t *testing.T) {
	cycles := []memory.CycleRecord{
		makeCycle("log"),
		makeCycle("log"),
		makeCycle("suggest"),
	}
	if s := stableStreak(cycles); s != 3 {
		t.Errorf("expected streak 3, got %d", s)
	}
}

func TestStableStreak_AfterAlert(t *testing.T) {
	cycles := []memory.CycleRecord{
		makeCycle("log"),
		makeCycle("alert"),
		makeCycle("log"),
		makeCycle("log"),
	}
	if s := stableStreak(cycles); s != 2 {
		t.Errorf("expected streak 2, got %d", s)
	}
}

func TestStableStreak_Empty(t *testing.T) {
	if s := stableStreak(nil); s != 0 {
		t.Errorf("expected 0 for empty cycles, got %d", s)
	}
}

func TestComputeCorrelation_MemoryLeak(t *testing.T) {
	// RAM monte fortement, CPU stable
	snaps := []memory.Snapshot{
		makeSnap(20, 30, 60),
		makeSnap(21, 40, 60),
		makeSnap(20, 50, 60),
	}
	if c := computeCorrelation(snaps); !strings.Contains(c, "memory leak") {
		t.Errorf("expected memory leak correlation, got %q", c)
	}
}

func TestComputeCorrelation_CPUSurge(t *testing.T) {
	snaps := []memory.Snapshot{
		makeSnap(10, 50, 60),
		makeSnap(20, 51, 60),
		makeSnap(30, 52, 60),
	}
	if c := computeCorrelation(snaps); !strings.Contains(c, "surge") {
		t.Errorf("expected CPU surge correlation, got %q", c)
	}
}

func TestComputeCorrelation_Stable(t *testing.T) {
	snaps := []memory.Snapshot{
		makeSnap(20, 50, 60),
		makeSnap(21, 51, 60),
		makeSnap(20, 50, 60),
	}
	if c := computeCorrelation(snaps); c != "" {
		t.Errorf("expected empty correlation for stable metrics, got %q", c)
	}
}

func TestForecastCycles_Rising(t *testing.T) {
	// CPU monte de 10% par snapshot, seuil à 85%, départ à 55%
	snaps := []memory.Snapshot{
		makeSnap(55, 50, 60),
		makeSnap(65, 50, 60),
		makeSnap(75, 50, 60),
	}
	f := forecastCycles(snaps, func(s memory.Snapshot) float64 { return s.CPUPercent }, 85.0)
	if !strings.Contains(f, "cycles") {
		t.Errorf("expected forecast with 'cycles', got %q", f)
	}
}

func TestForecastCycles_Stable(t *testing.T) {
	snaps := []memory.Snapshot{
		makeSnap(30, 50, 60),
		makeSnap(30, 50, 60),
		makeSnap(30, 50, 60),
	}
	if f := forecastCycles(snaps, func(s memory.Snapshot) float64 { return s.CPUPercent }, 85.0); f != "" {
		t.Errorf("expected empty forecast for stable metric, got %q", f)
	}
}

func TestForecastCycles_AlreadyAtThreshold(t *testing.T) {
	snaps := []memory.Snapshot{
		makeSnap(85, 50, 60),
		makeSnap(87, 50, 60),
		makeSnap(90, 50, 60),
	}
	if f := forecastCycles(snaps, func(s memory.Snapshot) float64 { return s.CPUPercent }, 85.0); f != "" {
		t.Errorf("expected empty forecast when already at threshold, got %q", f)
	}
}

func TestBuildAdaptiveSystemPrompt_TimeAndStreak(t *testing.T) {
	ctx := AdaptiveContext{
		CycleCount:   10,
		TimeOfDay:    "journée",
		IsWeekend:    false,
		StableStreak: 5,
		CPUTrend:     "stable (30.0%)",
		RAMTrend:     "stable (50.0%)",
		DiskTrend:    "stable (60.0%)",
	}
	result := BuildAdaptiveSystemPrompt("base", ctx)
	if !strings.Contains(result, "journée") {
		t.Error("expected time of day in prompt")
	}
	if !strings.Contains(result, "5 cycles") {
		t.Error("expected stable streak in prompt")
	}
}

func TestBuildAdaptiveSystemPrompt_Forecast(t *testing.T) {
	ctx := AdaptiveContext{
		CycleCount:  5,
		CPUForecast: "~3 cycles avant seuil",
		CPUTrend:    "en hausse (60.0%)",
		RAMTrend:    "stable (50.0%)",
		DiskTrend:   "stable (60.0%)",
		TimeOfDay:   "matin",
	}
	result := BuildAdaptiveSystemPrompt("base", ctx)
	if !strings.Contains(result, "Projection") {
		t.Error("expected Projection section in prompt")
	}
	if !strings.Contains(result, "~3 cycles") {
		t.Error("expected forecast value in prompt")
	}
}

func TestBuildAdaptiveSystemPrompt_Correlation(t *testing.T) {
	ctx := AdaptiveContext{
		CycleCount:  5,
		Correlation: "RAM↑ CPU stable → memory leak potentiel",
		CPUTrend:    "stable (20.0%)",
		RAMTrend:    "en hausse (70.0%)",
		DiskTrend:   "stable (60.0%)",
		TimeOfDay:   "soirée",
	}
	result := BuildAdaptiveSystemPrompt("base", ctx)
	if !strings.Contains(result, "memory leak") {
		t.Error("expected correlation diagnostic in prompt")
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
