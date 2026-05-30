package tools

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"os/exec"
	"runtime"
)

const defaultCommandTimeout = 10 * time.Second

// CommandSpec — binary + direct args, no shell in between
type CommandSpec struct {
	bin  string
	args []string
}

// whitelist — map commande → spec d'exécution directe
var commandSpecs = map[string]CommandSpec{
	"docker ps":    {bin: "docker", args: []string{"ps"}},
	"docker stats": {bin: "docker", args: []string{"stats", "--no-stream"}},
	"uptime":       {bin: "uptime", args: []string{}},
	"df -h":        {bin: "df", args: []string{"-h"}},
	"free -h":      {bin: "free", args: []string{"-h"}},
}

// windowsSpecs — équivalents Windows pour chaque commande Unix
var windowsSpecs = map[string]CommandSpec{
	"uptime":  {bin: "cmd", args: []string{"/C", "net statistics workstation"}},
	"df -h":   {bin: "wmic", args: []string{"logicaldisk", "get", "size,freespace,caption"}},
	"free -h": {bin: "wmic", args: []string{"OS", "get", "FreePhysicalMemory,TotalVisibleMemorySize"}},
}

type CommandResult struct {
	Command  string
	Output   string
	Error    string
	Duration time.Duration
	Success  bool
	TimedOut bool
}

// ExecuteCommandDryRun simule l'exécution sans rien faire
func ExecuteCommandDryRun(cmd string) CommandResult {
	cmd = strings.TrimSpace(cmd)

	// Vérifie quand même la whitelist — signale si la commande serait bloquée
	if _, ok := commandSpecs[cmd]; !ok {
		return CommandResult{
			Command: cmd,
			Error:   fmt.Sprintf("commande non autorisée : '%s'", cmd),
			Success: false,
		}
	}

	return CommandResult{
		Command:  cmd,
		Output:   "[dry-run] commande simulée",
		Duration: 0,
		Success:  true,
	}
}

func ExecuteCommand(cmd string) CommandResult {
	return ExecuteCommandWithTimeout(cmd, defaultCommandTimeout)
}

func ExecuteCommandWithTimeout(cmd string, timeout time.Duration) CommandResult {
	start := time.Now()
	cmd = strings.TrimSpace(cmd)

	// Lookup dans la whitelist
	spec, ok := commandSpecs[cmd]
	if !ok {
		return CommandResult{
			Command: cmd,
			Error:   fmt.Sprintf("commande non autorisée : '%s'", cmd),
			Success: false,
		}
	}

	// Adaptation Windows
	if runtime.GOOS == "windows" {
		if winSpec, hasWin := windowsSpecs[cmd]; hasWin {
			spec = winSpec
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	output, timedOut, err := runDirect(ctx, spec)
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
		Output:   output,
		Duration: duration,
		Success:  true,
	}
}

// runDirect — direct execution, no shell in between
func runDirect(ctx context.Context, spec CommandSpec) (string, bool, error) {
	command := exec.CommandContext(ctx, spec.bin, spec.args...)

	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	err := command.Run()

	if ctx.Err() == context.DeadlineExceeded {
		return "", true, nil
	}

	if err != nil {
		errMsg := stderr.String()
		if errMsg == "" {
			errMsg = err.Error()
		}
		return "", false, fmt.Errorf("%s", strings.TrimSpace(errMsg))
	}

	return strings.TrimSpace(stdout.String()), false, nil
}

func (r CommandResult) Display() {
	if r.TimedOut {
		fmt.Printf("[ EXEC ] ⏱ '%s' — timeout après %v\n",
			r.Command, r.Duration.Round(time.Millisecond))
		return
	}
	if r.Success {
		fmt.Printf("[ EXEC ] ✓ '%s' (%v)\n",
			r.Command, r.Duration.Round(time.Millisecond))
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
