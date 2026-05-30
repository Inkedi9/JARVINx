package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	jxlog "github.com/Inkedi9/jarvinx/jxlog"
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
	alertLogger   *memory.Logger
	dispatcher    *NotifierDispatcher
	dryRun        bool
	state         AlertState
	mu            sync.Mutex
}

func NewAlertAgent(
	cpuThreshold, ramThreshold, diskThreshold float64,
	minCycles, cooldown int,
	alertFile string,
	dispatcher *NotifierDispatcher,
) *AlertAgent {
	return &AlertAgent{
		BaseAgent:     NewBaseAgent("alert", 15*time.Second),
		cpuThreshold:  cpuThreshold,
		ramThreshold:  ramThreshold,
		diskThreshold: diskThreshold,
		minCycles:     minCycles,
		cooldown:      cooldown,
		alertLogger:   memory.NewLogger(alertFile),
		dispatcher:    dispatcher,
		dryRun:        dispatcher.dryRun,
		state: AlertState{
			LastAlertCPU: -999,
			LastAlertRAM: -999,
			LastAlertDsk: -999,
		},
	}
}

func (a *AlertAgent) Run(ctx context.Context, actx AgentContext) error {
	alerts := a.Analyze(actx.Snapshot)

	if len(alerts) == 0 {
		a.recordSuccess()
		return nil
	}

	a.Dispatch(alerts)
	a.recordAlert()
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
			alerts = append(alerts, Alert{
				Timestamp:   snap.Timestamp,
				Level:       AlertCritical,
				Metric:      "CPU",
				Value:       snap.CPUPercent,
				Threshold:   a.cpuThreshold,
				CyclesAbove: a.state.CPUCycles,
				Message: fmt.Sprintf("CPU a %.1f%% depuis %d cycles consecutifs",
					snap.CPUPercent, a.state.CPUCycles),
			})
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
			alerts = append(alerts, Alert{
				Timestamp:   snap.Timestamp,
				Level:       AlertCritical,
				Metric:      "RAM",
				Value:       snap.MemPercent,
				Threshold:   a.ramThreshold,
				CyclesAbove: a.state.RAMCycles,
				Message: fmt.Sprintf("RAM a %.1f%% depuis %d cycles consecutifs",
					snap.MemPercent, a.state.RAMCycles),
			})
			a.state.LastAlertRAM = a.state.CurrentCycle
		}
	} else {
		a.state.RAMCycles = 0
	}

	// Disk
	if snap.DiskPercent >= a.diskThreshold &&
		a.state.CurrentCycle-a.state.LastAlertDsk >= a.cooldown {
		alerts = append(alerts, Alert{
			Timestamp: snap.Timestamp,
			Level:     AlertWarning,
			Metric:    "DISK",
			Value:     snap.DiskPercent,
			Threshold: a.diskThreshold,
			Message: fmt.Sprintf("Disque a %.1f%% - nettoyage recommande",
				snap.DiskPercent),
		})
		a.state.LastAlertDsk = a.state.CurrentCycle
	}

	return alerts
}

func (a *AlertAgent) Dispatch(alerts []Alert) {
	for _, alert := range alerts {
		a.logAlert(alert)
		a.printAlert(alert)
		a.dispatcher.Dispatch(alert)
	}
}

func (a *AlertAgent) printAlert(alert Alert) {
	if alert.Level == AlertCritical {
		fmt.Printf("\033[31m[ ALERT ] CRITICAL - %s : %s\033[0m\n",
			alert.Metric, alert.Message)
	} else {
		fmt.Printf("\033[33m[ ALERT ] WARNING - %s : %s\033[0m\n",
			alert.Metric, alert.Message)
	}
}

func (a *AlertAgent) logAlert(alert Alert) {
	file, err := os.OpenFile(a.alertLogger.Filepath(),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		jxlog.Error("ALERT", fmt.Sprintf("Log open failed: %v", err))
		return
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(alert); err != nil {
		jxlog.Error("ALERT", fmt.Sprintf("Log encode failed: %v", err))
	}
}
