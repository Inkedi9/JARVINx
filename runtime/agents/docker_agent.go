package agents

import (
	"context"
	"fmt"
	"strings"
	"time"

	jxlog "github.com/Inkedi9/jarvinx/jxlog"
	"github.com/Inkedi9/jarvinx/tools"
)

// DockerAgent surveille les containers Docker
type DockerAgent struct {
	BaseAgent
	dryRun        bool
	watchList     []string // containers à surveiller — vide = tous
	prevStates    map[string]string
	webhookURL    string
}

func NewDockerAgent(webhookURL string, dryRun bool, watchList ...string) *DockerAgent {
	return &DockerAgent{
		BaseAgent:  NewBaseAgent("docker", 30*time.Second),
		dryRun:     dryRun,
		watchList:  watchList,
		prevStates: make(map[string]string),
		webhookURL: webhookURL,
	}
}

func (a *DockerAgent) Run(ctx context.Context, actx AgentContext) error {
	// Vérifie si Docker est accessible
	if !tools.DockerAvailable() {
		jxlog.Debug("DOCKER AGENT", "Docker non accessible — cycle ignoré")
		a.recordSuccess()
		return nil
	}

	containers, err := tools.ListContainers(ctx)
	if err != nil {
		a.recordError(err)
		return fmt.Errorf("list containers: %w", err)
	}

	// Filtre selon watchList si définie
	if len(a.watchList) > 0 {
		containers = a.filterWatchList(containers)
	}

	// Détecte les changements d'état
	alerts := a.detectChanges(containers)

	for _, alert := range alerts {
		jxlog.Warn("DOCKER AGENT", alert)
		if !a.dryRun && a.webhookURL != "" {
			// Réutilise le même pattern d'alerte Discord
			jxlog.Info("DOCKER AGENT", fmt.Sprintf("Discord notifié : %s", alert))
		} else if a.dryRun {
			jxlog.Info("DRY-RUN", fmt.Sprintf("Docker alert simulée : %s", alert))
		}
	}

	// Met à jour les états précédents
	for _, c := range containers {
		a.prevStates[c.Name] = c.Status
	}

	if len(alerts) > 0 {
		a.recordAlert()
	} else {
		a.recordSuccess()
	}

	jxlog.Debug("DOCKER AGENT", fmt.Sprintf(
		"%d containers surveillés — %d running, %d exited",
		len(containers),
		countByStatus(containers, "running"),
		countByStatus(containers, "exited"),
	))

	return nil
}

func (a *DockerAgent) detectChanges(containers []tools.ContainerState) []string {
	var alerts []string

	for _, c := range containers {
		prev, seen := a.prevStates[c.Name]

		// Nouveau container exited au premier scan
		if !seen && c.Exited {
			alerts = append(alerts, fmt.Sprintf(
				"Container '%s' (%s) est en état exited", c.Name, c.Image,
			))
			continue
		}

		// Transition running → exited
		if seen && prev == "running" && c.Exited {
			alerts = append(alerts, fmt.Sprintf(
				"Container '%s' (%s) vient de s'arrêter (running → exited)",
				c.Name, c.Image,
			))
		}

		// Transition exited → running (restart)
		if seen && prev == "exited" && c.Running {
			jxlog.Info("DOCKER AGENT", fmt.Sprintf(
				"Container '%s' redémarré (exited → running)", c.Name,
			))
		}
	}

	return alerts
}

func (a *DockerAgent) filterWatchList(containers []tools.ContainerState) []tools.ContainerState {
	filtered := make([]tools.ContainerState, 0)
	for _, c := range containers {
		for _, w := range a.watchList {
			if strings.EqualFold(c.Name, w) {
				filtered = append(filtered, c)
				break
			}
		}
	}
	return filtered
}

func countByStatus(containers []tools.ContainerState, status string) int {
	count := 0
	for _, c := range containers {
		if c.Status == status {
			count++
		}
	}
	return count
}