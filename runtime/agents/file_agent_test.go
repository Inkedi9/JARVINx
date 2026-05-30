package agents

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Inkedi9/jarvinx/memory"
)

func TestFileAgent_Name(t *testing.T) {
	a := NewFileAgent([]string{}, 500, false)
	if a.Name() != "file" {
		t.Errorf("expected 'file', got '%s'", a.Name())
	}
}

func TestFileAgent_Schedule(t *testing.T) {
	a := NewFileAgent([]string{}, 500, false)
	if a.Schedule() != 5*time.Minute {
		t.Errorf("expected 5m, got %v", a.Schedule())
	}
}

func TestFileAgent_NoPathsSkipsCycle(t *testing.T) {
	a := NewFileAgent([]string{}, 500, false)

	err := a.Run(context.Background(), AgentContext{
		Snapshot: memory.Snapshot{},
		State:    memory.NewState(""),
		Logger:   memory.NewLogger(""),
	})

	if err != nil {
		t.Fatalf("expected no error with empty watchPaths, got: %v", err)
	}

	if a.Status().RunCount != 1 {
		t.Error("expected RunCount=1 even with no paths")
	}
}

func TestFileAgent_ScansExistingDir(t *testing.T) {
	// Crée un dossier temporaire avec des fichiers
	dir := t.TempDir()

	// Crée un fichier de 1KB
	f, _ := os.CreateTemp(dir, "test-*.txt")
	f.Write(make([]byte, 1024))
	f.Close()

	a := NewFileAgent([]string{dir}, 500, false)

	err := a.Run(context.Background(), AgentContext{
		Snapshot: memory.Snapshot{},
		State:    memory.NewState(""),
		Logger:   memory.NewLogger(""),
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// prevSizes doit être mis à jour
	if _, ok := a.prevSizes[dir]; !ok {
		t.Error("expected prevSizes to be updated after scan")
	}
}

func TestFileAgent_DetectsLargeFile(t *testing.T) {
	dir := t.TempDir()

	// Crée un fichier de 2MB
	f, _ := os.CreateTemp(dir, "large-*.bin")
	f.Write(make([]byte, 2*1024*1024))
	f.Close()

	// Seuil à 1MB pour déclencher l'alerte
	a := NewFileAgent([]string{dir}, 1, false)

	err := a.Run(context.Background(), AgentContext{
		Snapshot: memory.Snapshot{},
		State:    memory.NewState(""),
		Logger:   memory.NewLogger(""),
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Doit avoir enregistré une alerte
	if a.Status().AlertCount != 1 {
		t.Errorf("expected AlertCount=1 for large file, got %d", a.Status().AlertCount)
	}
}

func TestFileAgent_InvalidPathHandled(t *testing.T) {
	a := NewFileAgent([]string{"/nonexistent/path/xyz"}, 500, false)

	err := a.Run(context.Background(), AgentContext{
		Snapshot: memory.Snapshot{},
		State:    memory.NewState(""),
		Logger:   memory.NewLogger(""),
	})

	// Ne doit pas crasher — juste logger un warning
	if err != nil {
		t.Fatalf("invalid path should not return error, got: %v", err)
	}
}

func TestFileAgent_DryRunMode(t *testing.T) {
	dir := t.TempDir()

	f, _ := os.CreateTemp(dir, "large-*.bin")
	f.Write(make([]byte, 2*1024*1024))
	f.Close()

	// dry-run = true
	a := NewFileAgent([]string{dir}, 1, true)

	err := a.Run(context.Background(), AgentContext{
		Snapshot: memory.Snapshot{},
		State:    memory.NewState(""),
		Logger:   memory.NewLogger(""),
	})

	if err != nil {
		t.Fatalf("unexpected error in dry-run: %v", err)
	}
}
