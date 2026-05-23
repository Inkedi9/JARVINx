package tools

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// Commandes autorisées selon l'OS
var allowedCommands = map[string]bool{
	"docker ps":        true,
	"docker stats":     true,
	"uptime":           true,
	"df -h":            true,
	"free -h":          true,
	"systemctl status": true,
}

// Équivalents Windows pour le dev local
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
}

func ExecuteCommand(cmd string) CommandResult {
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

	// Adaptation Windows si nécessaire
	actualCmd := cmd
	if runtime.GOOS == "windows" {
		if alias, ok := windowsAliases[cmd]; ok {
			actualCmd = alias
		}
	}

	// Exécution
	result, err := runCommand(actualCmd)
	duration := time.Since(start)

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

func runCommand(cmd string) (string, error) {
	var command *exec.Cmd

	if runtime.GOOS == "windows" {
		command = exec.Command("cmd", "/C", cmd)
	} else {
		command = exec.Command("sh", "-c", cmd)
	}

	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	if err := command.Run(); err != nil {
		return "", fmt.Errorf("%v: %s", err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

func (r CommandResult) Display() {
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
