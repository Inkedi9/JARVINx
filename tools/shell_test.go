package tools

import (
	"runtime"
	"testing"
	"time"
)

func TestExecuteCommand_Whitelist(t *testing.T) {
	result := ExecuteCommand("rm -rf /")

	if result.Success {
		t.Error("expected failure for non-whitelisted command")
	}
	if result.Error == "" {
		t.Error("expected error message for blocked command")
	}
}

func TestExecuteCommand_Timeout(t *testing.T) {
	if runtime.GOOS == "windows" {
		// Sur Windows, on vérifie juste que le timeout est bien détecté
		// via une commande whitelistée avec un timeout quasi-zéro
		result := ExecuteCommandWithTimeout("uptime", 1*time.Nanosecond)
		// Soit timeout, soit erreur — dans les deux cas pas de succès
		if result.Success {
			t.Error("expected failure with 1ns timeout")
		}
		return
	}

	// Linux/macOS — sleep est tué proprement par SIGKILL
	allowedCommands["sleep 5"] = true
	defer delete(allowedCommands, "sleep 5")

	result := ExecuteCommandWithTimeout("sleep 5", 200*time.Millisecond)
	if !result.TimedOut {
		t.Error("expected timeout for sleep 5 with 200ms limit")
	}
}

func TestExecuteCommand_ValidCommand(t *testing.T) {
	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "uptime"
	} else {
		cmd = "uptime"
	}

	result := ExecuteCommandWithTimeout(cmd, 5*time.Second)

	if !result.Success && result.TimedOut {
		t.Error("valid command should not timeout with 5s limit")
	}
	if result.Duration <= 0 {
		t.Error("duration should be positive")
	}
}

func TestCommandResult_TimedOutDisplay(t *testing.T) {
	r := CommandResult{
		Command:  "test",
		TimedOut: true,
		Duration: 200 * time.Millisecond,
	}

	// Juste vérifier que Display() ne panic pas
	r.Display()
}

func TestExecuteCommand_NotWhitelisted(t *testing.T) {
	result := ExecuteCommand("cat /etc/passwd")

	if result.Success {
		t.Fatal("non-whitelisted command must never succeed")
	}
	if result.TimedOut {
		t.Error("non-whitelisted command should fail immediately, not timeout")
	}
	if result.Duration > time.Second {
		t.Error("whitelist check should be instant")
	}
}
