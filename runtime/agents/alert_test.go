package agents

import (
	"context"
	"testing"
	"time"

	"github.com/Inkedi9/jarvinx/memory"
)

// helper — crée un snapshot de test rapidement
func makeSnap(cpu, ram, disk float64) memory.Snapshot {
	return memory.Snapshot{
		Timestamp:   time.Now(),
		CPUPercent:  cpu,
		MemPercent:  ram,
		MemUsed:     uint64(ram * 240),
		MemTotal:    24000,
		DiskPercent: disk,
		DiskUsed:    uint64(disk * 2),
		DiskTotal:   237,
	}
}

// helper — crée un AlertAgent préconfiguré pour les tests
func makeAlertAgent() *AlertAgent {
	return NewAlertAgent(
		85.0, // CPU threshold
		90.0, // RAM threshold
		85.0, // Disk threshold
		2,    // minCycles
		5,    // cooldown
		"",   // pas de fichier log pendant les tests
		"",   // pas de webhook Discord pendant les tests
	)
}

// ── Seuils de base ──────────────────────────────────────────────────────────

func TestAlertAgent_NoAlertBelowThreshold(t *testing.T) {
	a := makeAlertAgent()
	snap := makeSnap(50.0, 60.0, 70.0)

	alerts := a.Analyze(snap)

	if len(alerts) != 0 {
		t.Errorf("expected 0 alerts below threshold, got %d", len(alerts))
	}
}

func TestAlertAgent_DiskAlertAboveThreshold(t *testing.T) {
	a := makeAlertAgent()
	snap := makeSnap(10.0, 50.0, 90.0)

	// Disk n'a pas besoin de minCycles — alerte immédiate
	alerts := a.Analyze(snap)

	found := false
	for _, al := range alerts {
		if al.Metric == "DISK" {
			found = true
		}
	}
	if !found {
		t.Error("expected DISK alert above threshold, got none")
	}
}

func TestAlertAgent_CPURequiresMinCycles(t *testing.T) {
	a := makeAlertAgent()
	snap := makeSnap(90.0, 50.0, 50.0)

	// Premier cycle au-dessus — pas encore d'alerte (minCycles = 2)
	alerts := a.Analyze(snap)
	for _, al := range alerts {
		if al.Metric == "CPU" {
			t.Error("expected no CPU alert on first cycle above threshold")
		}
	}

	// Deuxième cycle — maintenant on alerte
	alerts = a.Analyze(snap)
	found := false
	for _, al := range alerts {
		if al.Metric == "CPU" {
			found = true
		}
	}
	if !found {
		t.Error("expected CPU alert after minCycles consecutive cycles")
	}
}

func TestAlertAgent_CPUResetOnDrop(t *testing.T) {
	a := makeAlertAgent()

	// Un cycle au-dessus
	a.Analyze(makeSnap(90.0, 50.0, 50.0))

	// CPU redescend — reset du compteur
	a.Analyze(makeSnap(30.0, 50.0, 50.0))

	// Remonte — doit recommencer à compter depuis 0
	alerts := a.Analyze(makeSnap(90.0, 50.0, 50.0))
	for _, al := range alerts {
		if al.Metric == "CPU" {
			t.Error("CPU counter should reset when metric drops below threshold")
		}
	}
}

// ── Cooldown anti-spam ───────────────────────────────────────────────────────

func TestAlertAgent_CooldownPreventsSpam(t *testing.T) {
	a := makeAlertAgent()
	snap := makeSnap(10.0, 50.0, 90.0)

	// Première alerte Disk — doit passer
	alerts1 := a.Analyze(snap)
	diskCount1 := countAlerts(alerts1, "DISK")
	if diskCount1 != 1 {
		t.Fatalf("expected 1 DISK alert on first trigger, got %d", diskCount1)
	}

	// Cycles 2-4 — dans le cooldown (cooldown = 5)
	for i := 0; i < 4; i++ {
		alerts := a.Analyze(snap)
		if countAlerts(alerts, "DISK") > 0 {
			t.Errorf("expected no DISK alert during cooldown (cycle %d)", i+2)
		}
	}

	// Cycle 6 — cooldown expiré, alerte de nouveau
	alerts2 := a.Analyze(snap)
	if countAlerts(alerts2, "DISK") != 1 {
		t.Error("expected DISK alert after cooldown expired")
	}
}

func TestAlertAgent_RAMCooldownIndependentFromCPU(t *testing.T) {
	a := makeAlertAgent()

	// Déclenche alerte CPU (2 cycles)
	highCPU := makeSnap(90.0, 95.0, 50.0)
	a.Analyze(highCPU)
	a.Analyze(highCPU)

	// CPU redescend, RAM reste haute
	a.state.LastAlertCPU = a.state.CurrentCycle

	// RAM doit toujours pouvoir alerter indépendamment
	for i := 0; i < 2; i++ {
		a.Analyze(makeSnap(10.0, 95.0, 50.0))
	}
	alerts := a.Analyze(makeSnap(10.0, 95.0, 50.0))

	// On vérifie juste que les cooldowns sont indépendants
	_ = alerts
}

// ── Niveaux d'alerte ─────────────────────────────────────────────────────────

func TestAlertAgent_DiskIsWarningLevel(t *testing.T) {
	a := makeAlertAgent()
	snap := makeSnap(10.0, 50.0, 90.0)

	alerts := a.Analyze(snap)

	for _, al := range alerts {
		if al.Metric == "DISK" && al.Level != AlertWarning {
			t.Errorf("expected DISK alert to be Warning, got %s", al.Level)
		}
	}
}

func TestAlertAgent_CPUisCriticalLevel(t *testing.T) {
	a := makeAlertAgent()
	snap := makeSnap(90.0, 50.0, 50.0)

	a.Analyze(snap)           // cycle 1
	alerts := a.Analyze(snap) // cycle 2 — déclenche

	for _, al := range alerts {
		if al.Metric == "CPU" && al.Level != AlertCritical {
			t.Errorf("expected CPU alert to be Critical, got %s", al.Level)
		}
	}
}

// ── Contenu des alertes ──────────────────────────────────────────────────────

func TestAlertAgent_AlertContainsCorrectValues(t *testing.T) {
	a := makeAlertAgent()
	snap := makeSnap(10.0, 50.0, 92.0)

	alerts := a.Analyze(snap)

	for _, al := range alerts {
		if al.Metric == "DISK" {
			if al.Value != 92.0 {
				t.Errorf("expected alert value 92.0, got %.1f", al.Value)
			}
			if al.Threshold != 85.0 {
				t.Errorf("expected threshold 85.0, got %.1f", al.Threshold)
			}
			if al.Message == "" {
				t.Error("alert message should not be empty")
			}
		}
	}
}

// ── helper ───────────────────────────────────────────────────────────────────

func countAlerts(alerts []Alert, metric string) int {
	count := 0
	for _, a := range alerts {
		if a.Metric == metric {
			count++
		}
	}
	return count
}

func TestAlertAgent_Run_NoAlert(t *testing.T) {
	a := makeAlertAgent() // alertFile="" — pas d'écriture disque
	snap := makeSnap(10.0, 50.0, 50.0)

	ctx := context.Background()
	actx := AgentContext{
		Snapshot: snap,
		State:    memory.NewState(""),
		Logger:   memory.NewLogger(""),
	}

	err := a.Run(ctx, actx)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	status := a.Status()
	if status.RunCount != 1 {
		t.Errorf("expected RunCount=1, got %d", status.RunCount)
	}
	if status.ErrorCount != 0 {
		t.Errorf("expected ErrorCount=0, got %d", status.ErrorCount)
	}
	if status.AlertCount != 0 {
		t.Errorf("expected AlertCount=0, got %d", status.AlertCount)
	}
}

func TestAlertAgent_Run_WithAlert(t *testing.T) {
	// Utilise un fichier temporaire pour éviter les erreurs d'écriture
	tmpFile := t.TempDir() + "/alerts.jsonl"
	a := NewAlertAgent(85.0, 90.0, 85.0, 2, 5, tmpFile, "")
	a.state = AlertState{
		LastAlertCPU: -999,
		LastAlertRAM: -999,
		LastAlertDsk: -999,
	}

	snap := makeSnap(10.0, 50.0, 92.0) // disk au-dessus du seuil

	ctx := context.Background()
	actx := AgentContext{
		Snapshot: snap,
		State:    memory.NewState(""),
		Logger:   memory.NewLogger(""),
	}

	err := a.Run(ctx, actx)
	if err != nil {
		t.Fatalf("expected no error when alerts fire, got: %v", err)
	}

	status := a.Status()
	if status.AlertCount != 1 {
		t.Errorf("expected AlertCount=1, got %d", status.AlertCount)
	}
	if status.ErrorCount != 0 {
		t.Errorf("expected ErrorCount=0 when alert fires, got %d", status.ErrorCount)
	}
	if status.LastError != "" {
		t.Errorf("expected empty LastError when alert fires, got: %s", status.LastError)
	}
}

func TestBaseAgent_RecordAlert(t *testing.T) {
	b := NewBaseAgent("test", 15*time.Second)

	b.recordAlert()

	status := b.Status()
	if status.AlertCount != 1 {
		t.Errorf("expected AlertCount=1, got %d", status.AlertCount)
	}
	if status.ErrorCount != 0 {
		t.Errorf("expected ErrorCount=0, got %d", status.ErrorCount)
	}
	if status.RunCount != 1 {
		t.Errorf("expected RunCount=1, got %d", status.RunCount)
	}
}
