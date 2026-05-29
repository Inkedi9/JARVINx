package memory

import (
	"os"
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
	s.Add(Snapshot{Timestamp: time.Now(), CPUPercent: 10.0})

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
