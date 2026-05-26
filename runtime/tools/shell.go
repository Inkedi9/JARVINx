package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const defaultCommandTimeout = 10 * time.Second

var allowedCommands = map[string]bool{
	"docker ps":        true,
	"docker stats":     true,
	"uptime":           true,
	"df -h":            true,
	"free -h":          true,
	"systemctl status": true,
}

var windowsAliases = map[string]string{
	"uptime":  "net statistics workstation",
	"df -h":   "wmic logicaldisk get size,freespace,caption",
	"free -h": "wmic OS get FreePhysicalMemory,TotalVisibleMemorySize",
}

type CommandResult struct {
	Command  string
	Output   string
	Error    string
	Duration time.Duration
	Success  bool
	TimedOut bool
}

func ExecuteCommand(cmd string) CommandResult {
	return ExecuteCommandWithTimeout(cmd, defaultCommandTimeout)
}

func ExecuteCommandWithTimeout(cmd string, timeout time.Duration) CommandResult {
	start := time.Now()
	cmd = strings.TrimSpace(cmd)

	// Vérification whitelist
	if !allowedCommands[cmd] {
		return CommandResult{
			Command: cmd,
			Error:   fmt.Sprintf("commande non autorisée : '%s'", cmd),
			Success: false,
		}
	}

	// Adaptation Windows
	actualCmd := cmd
	if runtime.GOOS == "windows" {
		if alias, ok := windowsAliases[cmd]; ok {
			actualCmd = alias
		}
	}

	// Context avec timeout — tué proprement après N secondes
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	result, timedOut, err := runCommandWithContext(ctx, actualCmd)
	duration := time.Since(start)

	if timedOut {
		return CommandResult{
			Command:  cmd,
			Error:    fmt.Sprintf("timeout après %v", timeout),
			Duration: duration,
			Success:  false,
			TimedOut: true,
		}
	}

	if err != nil {
		return CommandResult{
			Command:  cmd,
			Error:    err.Error(),
			Duration: duration,
			Success:  false,
		}
	}

	return CommandResult{
		Command:  cmd,
		Output:   result,
		Duration: duration,
		Success:  true,
	}
}

func runCommandWithContext(ctx context.Context, cmd string) (string, bool, error) {
	var command *exec.Cmd

	if runtime.GOOS == "windows" {
		command = exec.CommandContext(ctx, "cmd", "/C", cmd)
	} else {
		command = exec.CommandContext(ctx, "sh", "-c", cmd)
	}

	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	err := command.Run()

	// Vérifie si c'est un timeout
	if ctx.Err() == context.DeadlineExceeded {
		return "", true, nil
	}

	if err != nil {
		return "", false, fmt.Errorf("%v: %s", err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), false, nil
}

func (r CommandResult) Display() {
	if r.TimedOut {
		fmt.Printf("[ EXEC ] ⏱ '%s' — timeout après %v\n", r.Command, r.Duration.Round(time.Millisecond))
		return
	}
	if r.Success {
		fmt.Printf("[ EXEC ] ✓ '%s' (%v)\n", r.Command, r.Duration.Round(time.Millisecond))
		fmt.Printf("[ EXEC ] Output : %s\n", truncate(r.Output, 200))
	} else {
		fmt.Printf("[ EXEC ] ✗ '%s' — %s\n", r.Command, r.Error)
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
