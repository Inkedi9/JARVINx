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

Create `runtime/.env` with the following vars (no `.env.example` exists):
- `OLLAMA_URL` — defaults to `http://localhost:11434`
- `OLLAMA_MODEL` — e.g. `llama3`
- `DISCORD_WEBHOOK` — optional; omit to disable Discord alerts
- `JARVINX_DEBUG=true` — enables DEBUG-level logs
- `WEB_PORT` — defaults to `8080`
- `CPU_ALERT_THRESHOLD`, `RAM_ALERT_THRESHOLD`, `DISK_ALERT_THRESHOLD` — percentage floats
- `ALERT_COOLDOWN`, `ALERT_MIN_CYCLES` — alert dampening
- `ALLOWED_ORIGINS` — CORS origins (defaults include `localhost:3000`)

Dashboard: use `.env.local` (dev), `.env.homelab`, or `.env.tailscale` for network deployments. The only variable is `NEXT_PUBLIC_RUNTIME_URL` (defaults to `http://localhost:8080`).

## CI/CD

GitHub Actions (`.github/workflows/ci.yml`) runs on push to `main`/`develop` and on PRs to `main`:
- **Go job:** `golangci-lint run` then `go test -race -coverprofile=coverage.out ./...`
- **Dashboard job:** Node 20, `npm run lint`, `npm test`, `npm run build`

## Architecture

### Runtime Package Responsibilities

| Package | Role |
|---------|------|
| `cmd/` | Entry point; wires config, runtime, signals |
| `core/` | Runtime, Bus, Scheduler, Orchestrator, CLI |
| `agents/` | Agent interface, BaseAgent, Registry, SystemAgent, AlertAgent |
| `llm/` | OllamaClient, JSON parser, prompt builder, retry logic |
| `memory/` | State (state.json), Logger (logs.jsonl / alerts.jsonl) |
| `tools/` | System metrics via gopsutil, shell executor with whitelist |
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
                     SystemAgent: Ollama LLM → JSON decision
                     AlertAgent:  thresholds → Discord webhook
                                     ↓
                         memory.State (state.json)   ← max 20 snapshots / 20 cycles
                         memory.Logger (logs.jsonl)
```

**Bus** is a buffered channel (size 10) with fan-out: `Subscribe()` returns a dedicated channel per consumer, `Unsubscribe()` closes it cleanly. Publishing is non-blocking; a warning is logged if full. Event types: `EventObserved`, `EventDecided`, `EventExecuted`, `EventLogged`, `EventError`.

### Agent Pattern

All agents implement the `Agent` interface (`agents/agent.go`). Embed `BaseAgent` for common state (lastRun, errorCount, enabled flag, RWMutex). `Run()` receives an `AgentContext` with snapshot, state, and logger. The Registry runs each agent in its own goroutine and isolates panics — a crashing agent is disabled, not fatal.

To add a new agent: implement `Agent`, register in `core/runtime.go` via `registry.Register()`.

### LLM Integration (`llm/`)

`OllamaClient` sends system + user prompts with retry (`DefaultRetryConfig`). The JSON parser (`parser.go`) strips markdown backticks, extracts embedded JSON from surrounding text, and falls back to a safe default decision on failure. Valid `action` values: `"log"`, `"alert"`, `"suggest"`, `"execute"`. The `ParseResult` struct exposes `Raw`, `Attempts`, and `Cleaned` for debugging.

### Web API (`web/`)

Go embeds the compiled dashboard into the binary via `embed.FS`. In dev, the Go server (`:8080`) serves only the API; the Next.js dev server (`:3000`) serves the UI. CORS origin check is an O(1) map lookup.

API endpoints: `GET /api/status`, `GET /api/history`, `GET /api/agents`, `POST /api/agents/{name}/toggle`.

### Dashboard (`dashboard/`)

Stack: **Next.js 16**, React 19, Tailwind v4, TypeScript, Jest. App Router with pages: Overview, Agents, History, Settings. Three domain hooks — `useStatus` (5s), `useAgents` (10s), `useHistory` (15s) — built on a generic `usePolling<T>`. TypeScript types in `lib/api.ts` mirror Go response structs exactly. Styling via Tailwind v4 CSS custom properties (`--color-bg-primary`, `--color-accent-blue`, etc.).

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
