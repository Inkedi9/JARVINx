package memory

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func makeEntry() LogEntry {
	return LogEntry{
		Timestamp:   time.Now(),
		CPUPercent:  10.0,
		MemUsed:     4000,
		MemTotal:    16000,
		MemPercent:  25.0,
		DiskUsed:    100,
		DiskTotal:   500,
		DiskPercent: 20.0,
	}
}

func TestLogger_Write(t *testing.T) {
	path := t.TempDir() + "/test.jsonl"
	l := NewLogger(path)

	if err := l.Write(makeEntry()); err != nil {
		t.Fatalf("Write() failed: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("file not created: %v", err)
	}
	if info.Size() == 0 {
		t.Error("expected non-empty log file")
	}
}

func TestLogger_Rotation(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.jsonl"

	// Limite très basse pour forcer la rotation
	l := NewLoggerWithRotation(path, 100, 3)

	// Écrit suffisamment pour dépasser la limite
	entry := makeEntry()
	for i := 0; i < 20; i++ {
		if err := l.Write(entry); err != nil {
			t.Fatalf("Write() failed at iteration %d: %v", i, err)
		}
	}

	// Le backup .1 doit exister
	if _, err := os.Stat(path + ".1"); err != nil {
		t.Error("expected backup .1 to exist after rotation")
	}
}

func TestLogger_MaxBackups(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.jsonl"

	l := NewLoggerWithRotation(path, 50, 2) // max 2 backups

	entry := makeEntry()
	// Force plusieurs rotations
	for i := 0; i < 100; i++ {
		_ = l.Write(entry)
	}

	// backup .3 ne doit pas exister (max = 2)
	if _, err := os.Stat(path + ".3"); err == nil {
		t.Error("backup .3 should not exist with maxBackups=2")
	}

	// backup .1 et .2 doivent exister
	for _, n := range []int{1, 2} {
		bak := fmt.Sprintf("%s.%d", path, n)
		if _, err := os.Stat(bak); err != nil {
			t.Errorf("backup .%d should exist: %v", n, err)
		}
	}
}

func TestLogger_NoRotationWhenSmall(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.jsonl"

	l := NewLoggerWithRotation(path, 10*1024*1024, 3) // 10MB

	if err := l.Write(makeEntry()); err != nil {
		t.Fatalf("Write() failed: %v", err)
	}

	// Pas de backup — fichier trop petit
	if _, err := os.Stat(path + ".1"); err == nil {
		t.Error("no rotation expected for small file")
	}
}

func TestLogger_Size(t *testing.T) {
	path := t.TempDir() + "/test.jsonl"
	l := NewLogger(path)

	if l.Size() != 0 {
		t.Error("expected size 0 for non-existent file")
	}

	_ = l.Write(makeEntry())

	if l.Size() == 0 {
		t.Error("expected non-zero size after write")
	}
}

func TestLogger_ConcurrentWrites(t *testing.T) {
	path := t.TempDir() + "/concurrent.jsonl"
	l := NewLogger(path)

	done := make(chan struct{}, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_ = l.Write(makeEntry())
			done <- struct{}{}
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Vérifie que le fichier n'est pas corrompu
	if l.Size() == 0 {
		t.Error("expected non-empty file after concurrent writes")
	}
}

func TestLogger_Status_NoFile(t *testing.T) {
	path := t.TempDir() + "/nonexistent.jsonl"
	l := NewLoggerWithRotation(path, 10*1024*1024, 3)

	status := l.Status()

	if status.SizeBytes != 0 {
		t.Errorf("expected 0 size for nonexistent file, got %d", status.SizeBytes)
	}
	if status.Filepath != path {
		t.Errorf("expected filepath '%s', got '%s'", path, status.Filepath)
	}
}

func TestLogger_Status_WithFile(t *testing.T) {
	path := t.TempDir() + "/test.jsonl"
	l := NewLoggerWithRotation(path, 10*1024*1024, 3)

	_ = l.Write(makeEntry())

	status := l.Status()

	if status.SizeBytes == 0 {
		t.Error("expected non-zero size after write")
	}
	if status.SizeMB <= 0 {
		t.Error("expected positive SizeMB")
	}
	if status.UsedPercent <= 0 {
		t.Error("expected positive UsedPercent after write")
	}
}

func TestLogger_Status_BackupCount(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.jsonl"
	l := NewLoggerWithRotation(path, 50, 3)

	// Force plusieurs rotations
	entry := makeEntry()
	for i := 0; i < 50; i++ {
		_ = l.Write(entry)
	}

	status := l.Status()

	if status.BackupCount == 0 {
		t.Error("expected at least 1 backup after multiple rotations")
	}
	if len(status.Backups) != status.BackupCount {
		t.Errorf("Backups slice length %d != BackupCount %d",
			len(status.Backups), status.BackupCount)
	}
}
