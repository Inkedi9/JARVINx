package memory

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestState_FilePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permissions Unix non applicables sur Windows")
	}

	path := t.TempDir() + "/test_state.json"
	s := NewState(path)
	if err := s.Add(Snapshot{Timestamp: time.Now(), CPUPercent: 10.0}); err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	if err := s.Save(); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() failed: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("expected permissions 0600, got %04o", perm)
	}
}

func TestState_CapAt20Snapshots(t *testing.T) {
	s := NewState("")
	for i := range 25 {
		_ = s.Add(Snapshot{Timestamp: time.Now(), CPUPercent: float64(i)})
	}
	got := s.Last(30)
	if len(got) != 20 {
		t.Fatalf("expected 20 snapshots (cap), got %d", len(got))
	}
	if got[len(got)-1].CPUPercent != 24.0 {
		t.Errorf("expected last snapshot CPU=24, got %.1f", got[len(got)-1].CPUPercent)
	}
}

func TestState_CapAt20Cycles(t *testing.T) {
	s := NewState("")
	for range 25 {
		_ = s.AddCycle(CycleRecord{Action: "log", Timestamp: time.Now()})
	}
	got := s.LastCycles(30)
	if len(got) != 20 {
		t.Fatalf("expected 20 cycles (cap), got %d", len(got))
	}
	if s.CurrentCycle() != 25 {
		t.Errorf("expected CurrentCycle=25, got %d", s.CurrentCycle())
	}
}

func TestState_SaveLoadRoundtrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	s := NewState(path)

	_ = s.Add(Snapshot{Timestamp: time.Now(), CPUPercent: 42.0, MemPercent: 55.0})
	_ = s.AddCycle(CycleRecord{Action: "log", Analysis: "stable", Reason: "ok", Timestamp: time.Now()})

	if err := s.Save(); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	s2 := NewState(path)

	snaps := s2.Last(1)
	if len(snaps) != 1 || snaps[0].CPUPercent != 42.0 {
		t.Errorf("expected CPU=42.0 after reload, got %v", snaps)
	}
	cycles := s2.LastCycles(1)
	if len(cycles) != 1 || cycles[0].Action != "log" {
		t.Errorf("expected action=log after reload, got %v", cycles)
	}
	if s2.CurrentCycle() != 1 {
		t.Errorf("expected CurrentCycle=1 after reload, got %d", s2.CurrentCycle())
	}
	if s2.Version != currentStateVersion {
		t.Errorf("expected Version=%d after reload, got %d", currentStateVersion, s2.Version)
	}
}

func TestState_CurrentCycle(t *testing.T) {
	s := NewState("")
	if s.CurrentCycle() != 0 {
		t.Errorf("expected CurrentCycle=0 on new state, got %d", s.CurrentCycle())
	}
	_ = s.AddCycle(CycleRecord{Action: "log", Timestamp: time.Now()})
	if s.CurrentCycle() != 1 {
		t.Errorf("expected CurrentCycle=1, got %d", s.CurrentCycle())
	}
	_ = s.AddCycle(CycleRecord{Action: "alert", Timestamp: time.Now()})
	if s.CurrentCycle() != 2 {
		t.Errorf("expected CurrentCycle=2, got %d", s.CurrentCycle())
	}
}

func TestLogger_FilePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permissions Unix non applicables sur Windows")
	}

	path := t.TempDir() + "/test_logs.jsonl"
	l := NewLogger(path)

	entry := LogEntry{Timestamp: time.Now(), CPUPercent: 10.0}
	if err := l.Write(entry); err != nil {
		t.Fatalf("Write() failed: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() failed: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("expected permissions 0600, got %04o", perm)
	}
}
