# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

JARVINx is a local autonomous AI runtime: a Go backend that loops collect system metrics → call Ollama LLM → take decisions, plus a Next.js dashboard for real-time monitoring.

## Commands

### Go Runtime (`runtime/`)

```powershell
cd runtime
make run              # go run cmd/main.go (dev)
make build            # compile for current OS
make build-linux      # cross-compile linux/amd64
make build-windows    # cross-compile windows/amd64
go run cmd/main.go --dry-run   # simulation mode — no real alerts/notifications
go test ./...                    # all tests
go test ./llm/... -v             # single package, verbose
go test -race ./... -coverprofile=coverage.out  # with race detector (matches CI)
golangci-lint run                # lint (10 linters — see .golangci.yml)
```

### Next.js Dashboard (`dashboard/`)

```powershell
cd dashboard
npm install
npm run dev     # dev server on port 3000
npm run build   # production build
npm run lint    # ESLint
npm test        # Jest (--watchAll=false in CI)
npm start       # production server
```

### Environment

Create `runtime/.env` with the following vars (no `.env.example` exists). All `JARVINX_*` vars are read before the `.env` file is loaded, so environment variables already set take priority.

**Core**
- `JARVINX_OLLAMA_URL` — defaults to `http://localhost:11434`
- `JARVINX_MODEL` — e.g. `llama3.1:8b`
- `JARVINX_INTERVAL` — Go duration string, e.g. `30s` (defaults to `15s`, valid: 5s–1h)
- `JARVINX_DEBUG=true` — enables DEBUG-level logs
- `JARVINX_PORT` — web API port, defaults to `8080`
- `JARVINX_DRY_RUN=true` — simulates all notifications/commands (same as `--dry-run` CLI flag)
- `JARVINX_ALLOWED_ORIGINS` — comma-separated extra CORS origins (defaults include `localhost:3000`)

**Alerts**
- `JARVINX_CPU_THRESHOLD`, `JARVINX_RAM_THRESHOLD`, `JARVINX_DISK_THRESHOLD` — percentage floats (defaults: 85, 90, 85)
- `JARVINX_ALERT_COOLDOWN`, `JARVINX_ALERT_MIN_CYCLES` — alert dampening

**Notifications** (any combination; omit to disable)
- `DISCORD_WEBHOOK` — Discord webhook URL
- `SLACK_WEBHOOK` — Slack incoming webhook URL
- `NTFY_URL`, `NTFY_TOPIC` — ntfy.sh push notifications (defaults: `https://ntfy.sh`, `jarvinx`)
- `GOTIFY_URL`, `GOTIFY_TOKEN` — Gotify push notifications

**Log rotation**
- `JARVINX_LOG_MAX_MB` — max size of `logs.jsonl` in MB before rotation (default: 10)
- `JARVINX_LOG_MAX_BACKUPS` — number of rotated log files to keep (default: 3)

**Docker agent**
- `JARVINX_DOCKER_ENABLED=false` — disable Docker monitoring
- `JARVINX_DOCKER_WATCH` — comma-separated container names to watch (empty = all)

**File agent**
- `JARVINX_FILE_WATCH` — comma-separated directory paths to monitor
- `JARVINX_FILE_MAX_MB` — alert threshold per file in MB (default: 500)
- `JARVINX_FILE_ENABLED=false` — disable file monitoring

**Daily report**
- `JARVINX_DAILY_REPORT=true` — enable daily digest (disabled by default)
- `JARVINX_REPORT_HOUR`, `JARVINX_REPORT_MINUTE` — send time, 24h format (defaults: 8, 0)

Dashboard: use `.env.local` (dev), `.env.homelab`, or `.env.tailscale` for network deployments. The only variable is `NEXT_PUBLIC_RUNTIME_URL` (defaults to `http://localhost:8080`).

## CI/CD

GitHub Actions (`.github/workflows/ci.yml`) runs on push to `main`/`develop` and on PRs to `main`:
- **Go job:** `golangci-lint run` then `go test -race -coverprofile=coverage.out ./...`
- **Dashboard job:** Node 20, `npm run lint`, `npm test`, `npm run build`

## Architecture

### Runtime Package Responsibilities

| Package | Role |
|---------|------|
| `cmd/` | Entry point; wires config, runtime, signals; `--dry-run` CLI flag |
| `core/` | Runtime, Bus, Scheduler, Orchestrator, CLI |
| `agents/` | Agent interface, BaseAgent, Registry, SystemAgent, AlertAgent, DockerAgent, FileAgent, DailyReporter, NotifierDispatcher |
| `llm/` | OllamaClient, JSON parser, prompt builder, AdaptiveContext, retry logic |
| `memory/` | State (state.json), Logger with rotation (logs.jsonl / alerts.jsonl) |
| `tools/` | System metrics via gopsutil, shell executor with whitelist, Docker, filesystem scan |
| `web/` | HTTP server, CORS, embedded dashboard via embed.FS |
| `config/` | Config struct, .env loader, validation (interval 5s–1h) |
| `jxlog/` | Structured logging with custom slog handler |

Only external dependency: `github.com/shirou/gopsutil/v3` for cross-platform metrics.

### Core Loop (runtime/)

The observe → think → decide → act cycle runs every 15s (configurable):

```
Scheduler → tools.Observe() → Bus(EventObserved)
                                     ↓
                              Orchestrator
                                     ↓
                         Agent Registry (each agent)
                                     ↓
                     SystemAgent:  Ollama LLM → JSON decision
                     AlertAgent:   thresholds → NotifierDispatcher
                     DockerAgent:  container state changes → NotifierDispatcher (30s interval)
                     FileAgent:    large files / directory growth (5min interval)
                                     ↓
                         memory.State (state.json)   ← max 20 snapshots / 20 cycles
                         memory.Logger (logs.jsonl, alerts.jsonl, with rotation)
```

**Bus** is a buffered channel (size 10) with fan-out: `Subscribe()` returns a dedicated channel per consumer, `Unsubscribe()` closes it cleanly. Publishing is non-blocking; a warning is logged if full. Event types: `EventObserved`, `EventDecided`, `EventExecuted`, `EventLogged`, `EventError`.

**DailyReporter** runs as a standalone goroutine (not via the Registry): it ticks every minute and sends a 24h digest via the NotifierDispatcher at the configured hour:minute.

### Agent Pattern

All agents implement the `Agent` interface (`agents/agent.go`). Embed `BaseAgent` for common state (lastRun, errorCount, enabled flag, RWMutex). `Run()` receives an `AgentContext` with snapshot, state, and logger. The Registry runs each agent in its own goroutine and isolates panics — a crashing agent is disabled, not fatal.

To add a new agent: implement `Agent`, register in `core/runtime.go` via `registry.Register()`.

### Notification System (`agents/notifier.go`)

`NotifierDispatcher` fan-outs alerts to all registered `Notifier` implementations. Built-in channels: `DiscordNotifier`, `SlackNotifier`, `NtfyNotifier`, `GotifyNotifier`. Register via `dispatcher.Register()`. All channels respect `dryRun` mode — alerts are logged but not sent.

`AlertAgent` calls `dispatcher.Dispatch(alert)`. `DailyReporter` re-uses the same dispatcher with a special `Alert{Metric: "DAILY REPORT"}`.

### LLM Integration (`llm/`)

`OllamaClient` sends system + user prompts with retry (`DefaultRetryConfig`). The JSON parser (`parser.go`) strips markdown backticks, extracts embedded JSON from surrounding text, and falls back to a safe default decision on failure. Valid `action` values: `"log"`, `"alert"`, `"suggest"`, `"execute"`. The `ParseResult` struct exposes `Raw`, `Attempts`, and `Cleaned` for debugging.

`BuildAdaptiveContext()` (`context_builder.go`) analyzes recent cycles and snapshots to produce CPU/RAM/Disk trend strings (`stable`, `rising`, `high`, `falling`) and alert rate. `BuildAdaptiveSystemPrompt()` appends this context to the base system prompt, making the LLM aware of historical patterns.

A `CircuitBreaker` (`circuit_breaker.go`) wraps all Ollama calls: after `maxFailures` consecutive errors it opens (blocks calls, returns `ErrCircuitOpen`), then transitions to half-open after `resetTimeout` to probe recovery. The current state is exposed via `GET /api/status` as `circuit_state`.

### Web API (`web/`)

Go embeds the compiled dashboard into the binary via `embed.FS`. In dev, the Go server (`:8080`) serves only the API; the Next.js dev server (`:3000`) serves the UI. CORS origin check is an O(1) map lookup.

API endpoints:

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/status` | Runtime status, uptime, circuit state, last cycle |
| GET | `/api/history` | Last 10 cycles (most recent first) |
| GET | `/api/agents` | Agent list with enabled/run/error counts |
| POST | `/api/agents/toggle` | Toggle agent by name — body: `{"name": "..."}` |
| GET | `/api/docker` | Container list with running/exited counts |
| GET | `/api/logs/status` | Log file sizes and rotation status |
| GET | `/api/file` | FileAgent status and watched paths |
| GET | `/api/daily-report` | DailyReporter schedule and last/next send |
| POST | `/api/daily-report/send` | Trigger an immediate report dispatch |
| GET | `/api/llm-context` | Adaptive context fed to the LLM (trends, alert rate) |

### Dashboard (`dashboard/`)

Stack: **Next.js 16**, React 19, Tailwind v4, TypeScript, Jest. App Router with pages: Overview, Agents, History, Containers, LLM Context, Settings. Three domain hooks — `useStatus` (5s), `useAgents` (10s), `useHistory` (15s) — built on a generic `usePolling<T>`. TypeScript types in `lib/api.ts` mirror Go response structs exactly. Styling via Tailwind v4 CSS custom properties (`--color-bg-primary`, `--color-accent-blue`, etc.).

> **Important:** Next.js 16 has breaking changes. Before editing Next.js-specific code, read `dashboard/AGENTS.md` and check `node_modules/next/dist/docs/`.

### Shell Executor (`tools/shell.go`)

Commands run directly (no `sh -c`). Exact whitelist: `docker ps`, `docker stats`, `uptime`, `df -h`, `free -h`. Windows aliases are applied automatically when `runtime.GOOS == "windows"`. Default timeout: 10s. Arbitrary commands are rejected — extending the whitelist requires editing `CommandSpec` entries in `shell.go`.

### Cross-Platform

`tools/` detects OS for paths (Windows `C:\`, Unix `/`) and shell command aliases. Version is injected at build via ldflags: `-X main.Version=<tag>`.

## Known Constraints

- **Race conditions** exist in the current runtime (known, tracked). Be careful when touching shared state in `core/` without holding the appropriate mutex.
- Ollama must be running locally before the runtime starts; a health check runs at startup and exits if Ollama is unreachable.
- Alert cooldown and minimum consecutive-cycle logic live in `AlertAgent` — changes there affect alert frequency directly.
- State is capped at **20 snapshots** and **20 cycles** in memory; older entries are dropped silently.
- `DockerAgent` gracefully skips cycles when Docker is not available (`tools.DockerAvailable()` check). `FileAgent` is a no-op when `FileWatchPaths` is empty.
