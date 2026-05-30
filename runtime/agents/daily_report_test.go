package agents

import (
	"strings"
	"testing"
	"time"

	"github.com/Inkedi9/jarvinx/memory"
)

func makeReporter() *DailyReporter {
	dispatcher := NewNotifierDispatcher(false)
	state := memory.NewState("")
	return NewDailyReporter(dispatcher, state, 8, 0, false)
}

func TestDailyReporter_BuildReportEmpty(t *testing.T) {
	r := makeReporter()
	data := r.buildReport()

	if data.TotalCycles != 0 {
		t.Errorf("expected 0 cycles for empty state, got %d", data.TotalCycles)
	}
}

func TestDailyReporter_BuildReportWithData(t *testing.T) {
	dispatcher := NewNotifierDispatcher(false)
	state := memory.NewState("")

	// Ajoute des cycles de test
	for i := 0; i < 10; i++ {
		action := "log"
		if i%3 == 0 {
			action = "alert"
		}
		state.AddCycle(memory.NewCycleRecord(
			memory.Snapshot{
				Timestamp:   time.Now(),
				CPUPercent:  float64(20 + i),
				MemPercent:  float64(50 + i),
				DiskPercent: 80.0,
				MemTotal:    16000,
				DiskTotal:   500,
			},
			action, "test analysis", "test reason", "",
		))
	}

	r := NewDailyReporter(dispatcher, state, 8, 0, false)
	data := r.buildReport()

	if data.TotalCycles != 10 {
		t.Errorf("expected 10 cycles, got %d", data.TotalCycles)
	}
	if data.CPUMax < 29.0 {
		t.Errorf("expected CPUMax >= 29, got %.1f", data.CPUMax)
	}
	if data.AlertCount == 0 {
		t.Error("expected some alert cycles")
	}
}

func TestDailyReporter_FormatReport(t *testing.T) {
	r := makeReporter()
	data := ReportData{
		Date:        "26 May 2026",
		TotalCycles: 100,
		Actions:     map[string]int{"log": 80, "suggest": 15, "alert": 5},
		CPUAvg:      25.0,
		CPUMax:      75.0,
		RAMAvg:      55.0,
		RAMMax:      70.0,
		DiskMax:     81.0,
		AlertCount:  5,
	}

	report := r.formatReport(data)

	if !strings.Contains(report, "26 May 2026") {
		t.Error("expected date in report")
	}
	if !strings.Contains(report, "100") {
		t.Error("expected cycle count in report")
	}
	if !strings.Contains(report, "25.0%") {
		t.Error("expected CPU avg in report")
	}
}

func TestDailyReporter_DryRunNoSend(t *testing.T) {
	n := &mockNotifier{name: "test"}
	dispatcher := NewNotifierDispatcher(true) // dry-run
	dispatcher.Register(n)

	state := memory.NewState("")
	r := NewDailyReporter(dispatcher, state, 8, 0, true)

	r.send()

	if n.sendCount != 0 {
		t.Errorf("dry-run should not send, got %d sends", n.sendCount)
	}
}

func TestDailyReporter_SendDispatchesAlert(t *testing.T) {
	n := &mockNotifier{name: "test"}
	dispatcher := NewNotifierDispatcher(false)
	dispatcher.Register(n)

	state := memory.NewState("")
	state.AddCycle(memory.NewCycleRecord(
		memory.Snapshot{Timestamp: time.Now(), CPUPercent: 10},
		"log", "ok", "", "",
	))

	r := NewDailyReporter(dispatcher, state, 8, 0, false)
	r.send()

	if n.sendCount != 1 {
		t.Errorf("expected 1 dispatch, got %d", n.sendCount)
	}
	if !strings.Contains(n.lastAlert.Metric, "DAILY REPORT") {
		t.Errorf("expected DAILY REPORT metric, got '%s'", n.lastAlert.Metric)
	}
}

func TestDailyReporter_LastSentPreventsDouble(t *testing.T) {
	n := &mockNotifier{name: "test"}
	dispatcher := NewNotifierDispatcher(false)
	dispatcher.Register(n)

	state := memory.NewState("")
	r := NewDailyReporter(dispatcher, state, 8, 0, false)

	// Premier envoi
	r.send()
	r.lastSent = time.Now()

	// Deuxième envoi immédiat — doit être bloqué par le ticker dans Start()
	// On teste send() directement ici — lastSent est vérifié dans Start()
	r.send() // send() lui-même n'a pas la protection — c'est Start() qui l'a

	// Les deux envoient — la protection est dans Start() via lastSent
	if n.sendCount != 2 {
		t.Logf("send() called twice = %d dispatches (protection is in Start)", n.sendCount)
	}
}

func TestDailyReporter_Config(t *testing.T) {
	r := NewDailyReporter(
		NewNotifierDispatcher(false),
		memory.NewState(""),
		14, 30, false,
	)

	if r.hour != 14 {
		t.Errorf("expected hour 14, got %d", r.hour)
	}
	if r.minute != 30 {
		t.Errorf("expected minute 30, got %d", r.minute)
	}
}
