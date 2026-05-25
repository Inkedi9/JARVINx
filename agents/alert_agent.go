package agents

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/Inkedi9/jarvinx/memory"
)

type AlertLevel string

const (
	AlertWarning  AlertLevel = "warning"
	AlertCritical AlertLevel = "critical"
)

type Alert struct {
	Timestamp   time.Time  `json:"timestamp"`
	Level       AlertLevel `json:"level"`
	Metric      string     `json:"metric"`
	Value       float64    `json:"value"`
	Threshold   float64    `json:"threshold"`
	Message     string     `json:"message"`
	CyclesAbove int        `json:"cycles_above"`
}

type AlertState struct {
	CPUCycles    int
	RAMCycles    int
	DiskCycles   int
	LastAlertCPU int
	LastAlertRAM int
	LastAlertDsk int
	CurrentCycle int
}

type AlertAgent struct {
	BaseAgent
	cpuThreshold  float64
	ramThreshold  float64
	diskThreshold float64
	minCycles     int
	cooldown      int
	alertFile     string
	webhookURL    string
	state         AlertState
	mu            sync.Mutex
	httpClient    *http.Client
}

func NewAlertAgent(
	cpuThreshold, ramThreshold, diskThreshold float64,
	minCycles, cooldown int,
	alertFile, webhookURL string,
) *AlertAgent {
	return &AlertAgent{
		BaseAgent:     NewBaseAgent("alert", 15*time.Second),
		cpuThreshold:  cpuThreshold,
		ramThreshold:  ramThreshold,
		diskThreshold: diskThreshold,
		minCycles:     minCycles,
		cooldown:      cooldown,
		alertFile:     alertFile,
		webhookURL:    webhookURL,
		httpClient:    &http.Client{Timeout: 10 * time.Second},
		state: AlertState{
			LastAlertCPU: -999,
			LastAlertRAM: -999,
			LastAlertDsk: -999,
		},
	}
}

func (a *AlertAgent) Run(ctx context.Context, actx AgentContext) error {
	alerts := a.Analyze(actx.Snapshot)
	a.Dispatch(alerts)

	if len(alerts) > 0 {
		a.recordError(fmt.Errorf("%d alertes déclenchées", len(alerts)))
	} else {
		a.recordSuccess()
	}

	return nil
}

func (a *AlertAgent) Analyze(snap memory.Snapshot) []Alert {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.state.CurrentCycle++
	var alerts []Alert

	// CPU
	if snap.CPUPercent >= a.cpuThreshold {
		a.state.CPUCycles++
		if a.state.CPUCycles >= a.minCycles &&
			a.state.CurrentCycle-a.state.LastAlertCPU >= a.cooldown {
			alert := Alert{
				Timestamp:   snap.Timestamp,
				Level:       AlertCritical,
				Metric:      "CPU",
				Value:       snap.CPUPercent,
				Threshold:   a.cpuThreshold,
				CyclesAbove: a.state.CPUCycles,
				Message: fmt.Sprintf("CPU à %.1f%% depuis %d cycles consécutifs",
					snap.CPUPercent, a.state.CPUCycles),
			}
			alerts = append(alerts, alert)
			a.state.LastAlertCPU = a.state.CurrentCycle
		}
	} else {
		a.state.CPUCycles = 0
	}

	// RAM
	if snap.MemPercent >= a.ramThreshold {
		a.state.RAMCycles++
		if a.state.RAMCycles >= a.minCycles &&
			a.state.CurrentCycle-a.state.LastAlertRAM >= a.cooldown {
			alert := Alert{
				Timestamp:   snap.Timestamp,
				Level:       AlertCritical,
				Metric:      "RAM",
				Value:       snap.MemPercent,
				Threshold:   a.ramThreshold,
				CyclesAbove: a.state.RAMCycles,
				Message: fmt.Sprintf("RAM à %.1f%% depuis %d cycles consécutifs",
					snap.MemPercent, a.state.RAMCycles),
			}
			alerts = append(alerts, alert)
			a.state.LastAlertRAM = a.state.CurrentCycle
		}
	} else {
		a.state.RAMCycles = 0
	}

	// Disk — seuil persistant, pas besoin de N cycles
	if snap.DiskPercent >= a.diskThreshold &&
		a.state.CurrentCycle-a.state.LastAlertDsk >= a.cooldown {
		alert := Alert{
			Timestamp: snap.Timestamp,
			Level:     AlertWarning,
			Metric:    "DISK",
			Value:     snap.DiskPercent,
			Threshold: a.diskThreshold,
			Message: fmt.Sprintf("Disque à %.1f%% — nettoyage recommandé",
				snap.DiskPercent),
		}
		alerts = append(alerts, alert)
		a.state.LastAlertDsk = a.state.CurrentCycle
	}

	return alerts
}

func (a *AlertAgent) Dispatch(alerts []Alert) {
	for _, alert := range alerts {
		a.logAlert(alert)

		if a.webhookURL != "" {
			if err := a.sendDiscord(alert); err != nil {
				fmt.Printf("[ ALERT ] Discord failed : %v\n", err)
			}
		}

		a.printAlert(alert)
	}
}

func (a *AlertAgent) printAlert(alert Alert) {
	level := "⚠️ WARNING"
	if alert.Level == AlertCritical {
		level = "🚨 CRITICAL"
	}
	fmt.Printf("[ ALERT ] %s — %s : %s\n", level, alert.Metric, alert.Message)
}

func (a *AlertAgent) logAlert(alert Alert) {
	file, err := os.OpenFile(a.alertFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("[ ALERT ] Log failed : %v\n", err)
		return
	}
	defer file.Close()
	json.NewEncoder(file).Encode(alert)
}

func (a *AlertAgent) sendDiscord(alert Alert) error {
	color := 16776960 // jaune warning
	if alert.Level == AlertCritical {
		color = 16711680 // rouge critical
	}

	emoji := "⚠️"
	if alert.Level == AlertCritical {
		emoji = "🚨"
	}

	payload := map[string]any{
		"username":   "JARVINx",
		"avatar_url": "",
		"embeds": []map[string]any{
			{
				"title":       emoji + " " + string(alert.Level) + " — " + alert.Metric,
				"description": alert.Message,
				"color":       color,
				"fields": []map[string]any{
					{"name": "Valeur", "value": fmt.Sprintf("%.1f%%", alert.Value), "inline": true},
					{"name": "Seuil", "value": fmt.Sprintf("%.1f%%", alert.Threshold), "inline": true},
					{"name": "Cycles", "value": fmt.Sprintf("%d", alert.CyclesAbove), "inline": true},
				},
				"footer": map[string]any{
					"text": "JARVINx · Autonomous Agent Runtime",
				},
				"timestamp": alert.Timestamp.Format(time.RFC3339),
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	resp, err := a.httpClient.Post(a.webhookURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("discord status: %d", resp.StatusCode)
	}

	fmt.Printf("[ ALERT ] Discord notifié — %s\n", alert.Metric)
	return nil
}
