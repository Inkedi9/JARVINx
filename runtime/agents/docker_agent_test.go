package agents

import (
	"context"
	"testing"
	"time"

	"github.com/Inkedi9/jarvinx/memory"
	"github.com/Inkedi9/jarvinx/tools"
)

func makeDockerAgent() *DockerAgent {
	return NewDockerAgent("", false)
}

func TestDockerAgent_Name(t *testing.T) {
	a := makeDockerAgent()
	if a.Name() != "docker" {
		t.Errorf("expected name 'docker', got '%s'", a.Name())
	}
}

func TestDockerAgent_Schedule(t *testing.T) {
	a := makeDockerAgent()
	if a.Schedule() != 30*time.Second {
		t.Errorf("expected 30s schedule, got %v", a.Schedule())
	}
}

func TestDockerAgent_DetectCrash(t *testing.T) {
	a := makeDockerAgent()

	// Simule un container running
	a.prevStates["nginx"] = "running"

	containers := []tools.ContainerState{
		{Name: "nginx", Image: "nginx:latest", Status: "exited", Exited: true},
	}

	alerts := a.detectChanges(containers)
	if len(alerts) == 0 {
		t.Error("expected crash alert for running → exited transition")
	}
}

func TestDockerAgent_DetectRestart(t *testing.T) {
	a := makeDockerAgent()

	// Simule un container exited qui redémarre
	a.prevStates["nginx"] = "exited"

	containers := []tools.ContainerState{
		{Name: "nginx", Image: "nginx:latest", Status: "running", Running: true},
	}

	alerts := a.detectChanges(containers)
	// Restart = info, pas alerte
	if len(alerts) != 0 {
		t.Errorf("restart should not trigger alert, got %d alerts", len(alerts))
	}
}

func TestDockerAgent_NoAlertIfStable(t *testing.T) {
	a := makeDockerAgent()
	a.prevStates["nginx"] = "running"

	containers := []tools.ContainerState{
		{Name: "nginx", Status: "running", Running: true},
	}

	alerts := a.detectChanges(containers)
	if len(alerts) != 0 {
		t.Errorf("expected no alerts for stable container, got %d", len(alerts))
	}
}

func TestDockerAgent_FilterWatchList(t *testing.T) {
	a := NewDockerAgent("", false, "nginx", "postgres")

	all := []tools.ContainerState{
		{Name: "nginx", Status: "running"},
		{Name: "redis", Status: "running"},
		{Name: "postgres", Status: "running"},
	}

	filtered := a.filterWatchList(all)
	if len(filtered) != 2 {
		t.Errorf("expected 2 containers after filter, got %d", len(filtered))
	}
}

func TestDockerAgent_WatchListCaseInsensitive(t *testing.T) {
	a := NewDockerAgent("", false, "NGINX")

	containers := []tools.ContainerState{
		{Name: "nginx", Status: "running"},
	}

	filtered := a.filterWatchList(containers)
	if len(filtered) != 1 {
		t.Errorf("watchlist should be case-insensitive, got %d matches", len(filtered))
	}
}

func TestDockerAgent_DryRunNoDiscord(t *testing.T) {
	a := NewDockerAgent("https://discord.com/fake-webhook", true) // dry-run = true

	ctx := context.Background()
	actx := AgentContext{
		Snapshot: memory.Snapshot{},
		State:    memory.NewState(""),
		Logger:   memory.NewLogger(""),
	}

	// Docker probablement pas dispo en CI — doit pas crasher
	err := a.Run(ctx, actx)
	if err != nil {
		t.Logf("Docker not available in test env (expected): %v", err)
	}
}

func TestCountByStatus(t *testing.T) {
	containers := []tools.ContainerState{
		{Status: "running", Running: true},
		{Status: "running", Running: true},
		{Status: "exited", Exited: true},
	}

	if countByStatus(containers, "running") != 2 {
		t.Error("expected 2 running")
	}
	if countByStatus(containers, "exited") != 1 {
		t.Error("expected 1 exited")
	}
}
