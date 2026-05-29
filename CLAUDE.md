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
go test ./...         # all tests
go test ./llm/... -v  # single package, verbose
go test ./... -cover  # with coverage
```

### Next.js Dashboard (`dashboard/`)

```powershell
cd dashboard
npm install
npm run dev     # dev server on port 3000
npm run build   # production build
npm start       # production server
```

### Environment

Copy `.env.example` to `.env` in `runtime/` before first run. Key vars:
- `DISCORD_WEBHOOK` — optional; omit to disable Discord alerts
- `JARVINX_DEBUG=true` — enables DEBUG-level logs

Dashboard: copy `.env.local` (dev) or `.env.homelab` / `.env.tailscale` for network deployments; the only variable is `NEXT_PUBLIC_RUNTIME_URL` (defaults to `http://localhost:8080`).

## Architecture

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
                         memory.State (state.json)
                         memory.Logger (logs.jsonl)
```

**Bus** is a buffered channel (size 10). Publishing is non-blocking; a warning is logged if full. Typed events: `EventObserved`, `EventDecided`, `EventExecuted`, `EventLogged`, `EventError`.

### Agent Pattern

All agents implement the `Agent` interface (`agents/agent.go`). Embed `BaseAgent` for the common state (lastRun, errorCount, enabled flag, RWMutex). `Run()` receives an `AgentContext` with snapshot, state, and logger. The Registry isolates panics — a crashing agent is disabled, not fatal.

To add a new agent: implement `Agent`, register it in `core/runtime.go` via `registry.Register()`.

### LLM Integration (`llm/`)

`OllamaClient` sends system + user prompts. The JSON parser (`parser.go`) handles malformed responses, markdown backticks, and falls back gracefully. The LLM is expected to return `{"action":"log"|"alert"|"suggest"|"execute", "reason":"...", "command":"..."}`.

### Web API (`web/`)

Go embeds the compiled dashboard into the binary via `embed.FS`. In dev, the Go server (`:8080`) serves only the API; the Next.js dev server (`:3000`) serves the UI. CORS middleware allows `:3000` → `:8080` in dev.

API endpoints: `GET /api/status`, `GET /api/history`, `GET /api/agents`, `POST /api/agents/{name}/toggle`.

### Dashboard (`dashboard/`)

App Router (Next.js). Pages: Overview, Agents, History, Settings. Polling via a generic `usePolling<T>` hook — 5s for status, 10s for agents, 15s for history. TypeScript types in `lib/api.ts` mirror Go response structs exactly. Styling via Tailwind v4 with CSS custom properties (`--color-bg-primary`, `--color-accent-blue`, etc.).

> **Important:** The dashboard uses Next.js 16 which has breaking changes vs older versions. Before editing Next.js-specific code, check `node_modules/next/dist/docs/` — see `dashboard/CLAUDE.md`.

## Known Constraints

- **Race conditions** exist in the current runtime (known, tracked). Be careful when touching shared state in `core/` without holding the appropriate mutex.
- Shell execution (`tools/shell.go`) uses a **command whitelist** — arbitrary commands are not allowed.
- Ollama must be running locally before the runtime starts; a health check runs at startup and will exit if Ollama is unreachable.
- Alert cooldown and minimum consecutive-cycle logic live in `AlertAgent` — changes there affect alert frequency directly.
