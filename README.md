<div align="center">

```
     ██╗ █████╗ ██████╗ ██╗   ██╗██╗███╗   ██╗██╗  ██╗
     ██║██╔══██╗██╔══██╗██║   ██║██║████╗  ██║╚██╗██╔╝
     ██║███████║██████╔╝██║   ██║██║██╔██╗ ██║ ╚███╔╝
██   ██║██╔══██║██╔══██╗╚██╗ ██╔╝██║██║╚██╗██║ ██╔██╗
╚█████╔╝██║  ██║██║  ██║ ╚████╔╝ ██║██║ ╚████║██╔╝ ██╗
 ╚════╝ ╚═╝  ╚═╝╚═╝  ╚═╝  ╚═══╝  ╚═╝╚═╝  ╚═══╝╚═╝  ╚═╝
Version 1.8.0
```

**Autonomous AI Runtime · Observing. Thinking. Acting. Evolving.**

![Go](https://img.shields.io/badge/Go-1.26.3-00ADD8?style=flat-square&logo=go&logoColor=white)
![Ollama](https://img.shields.io/badge/Ollama-local%20LLM-black?style=flat-square)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
![Status](https://img.shields.io/badge/status-v1.8%20stable-00E5FF?style=flat-square)

_Your system. My mission._

</div>

---

// Copyright 2026 Inkedi9

// Licensed under the Apache License, Version 2.0

// https://www.apache.org/licenses/LICENSE-2.0

---

## Documentation

| Document                                     | Description                                                         |
| -------------------------------------------- | ------------------------------------------------------------------- |
| [Manuel Utilisateur](docs/USER_GUIDE.md)     | Installation, configuration, CLI, dashboard, alertes, dépannage     |
| [Documentation Technique](docs/TECHNICAL.md) | Architecture, écrire un agent, API, tests, build, roadmap technique |
| [SQLite Store](docs/ADR-001-sqlite-store.md) | SQLite comme store mémoire longue durée, plan de migration          |

---

## What is JARVINx?

JARVINx est un **runtime agentique local** écrit en Go. Il observe ton système en continu, envoie l'état à un LLM local via Ollama, reçoit une décision structurée, et agit — le tout en autonomie complète.

Ce n'est pas un chatbot. Ce n'est pas un dashboard passif. C'est un système qui **pense et agit** sur ta machine, sans cloud, sans abonnement, sans dépendance externe.

```
observe → think → decide → act → log → repeat
```

**Stack :**

- **Go** — runtime concurrent, goroutines natives, binaire unique
- **Ollama** — LLM local (llama3.1, qwen2.5, mistral...)
- **Next.js 16 / React 19 / Tailwind v4** — dashboard web (TypeScript)
- **Discord / Slack / Ntfy / Gotify** — alertes multi-canal

---

## Architecture

```
jarvinx/
│
├── runtime/                     # Backend Go
│   ├── cmd/main.go              # Point d'entrée — config + lancement
│   ├── core/                    # Runtime, Bus, Scheduler, Orchestrator, CLI
│   ├── agents/                  # Interface Agent, Registry, tous les agents
│   │   ├── system_agent.go      # LLM → décision JSON
│   │   ├── alert_agent.go       # Seuils + NotifierDispatcher
│   │   ├── docker_agent.go      # Surveillance containers (30s)
│   │   ├── file_agent.go        # Surveillance fichiers lourds (5min)
│   │   ├── qdrant_agent.go      # Mémoire sémantique — embed + upsert + search (opt-in v1.8)
│   │   ├── daily_report.go      # Rapport quotidien (goroutine indépendante)
│   │   └── notifier.go          # Discord / Slack / Ntfy / Gotify
│   ├── llm/                     # Client Ollama, parser JSON, prompt adaptatif, embedder, circuit breaker
│   ├── memory/                  # Store/EventLog interfaces, state.json, SQLiteStore (historique illimité),
│   │                            #   DoubleWriteStore, logs.jsonl / alerts.jsonl avec rotation
│   ├── tools/                   # Métriques gopsutil, shell whitelist, Docker, filesystem
│   ├── web/                     # HTTP server, CORS, embed.FS (build Next.js)
│   ├── config/                  # Config centralisée + chargement .env
│   └── jxlog/                   # Structured logging slog custom
│
└── dashboard/                   # Frontend Next.js 16
    ├── app/                     # App Router — pages: Overview, Agents, Containers,
    │                            #   History, LLM Context, Settings
    ├── components/              # Composants React (metrics-bar, decision-feed, etc.)
    └── lib/                     # api.ts (types Go miroir), hooks.ts (usePolling)
```

### Agent loop

```
┌─────────────────────────────────────────────────────────────┐
│                      JARVINX RUNTIME                         │
│                                                              │
│  Scheduler ──tick──► Bus ──► Orchestrator                    │
│     (15s)          (chan)        │                           │
│                                 ├── SystemAgent (LLM)        │
│                                 │   └── Ollama API → JSON    │
│                                 │                            │
│                                 ├── AlertAgent               │
│                                 │   └── NotifierDispatcher   │
│                                 │       Discord/Slack/Ntfy/  │
│                                 │       Gotify               │
│                                 │                            │
│                                 ├── DockerAgent (30s)        │
│                                 │   └── crash/restart detect │
│                                 │                            │
│                                 ├── FileAgent (5min)         │
│                                 │   └── fichiers volumineux  │
│                                 │                            │
│                                 ├── QdrantAgent (15s)        │
│                                 │   └── embed + RAG (opt-in) │
│                                 │                            │
│                                 └── Memory                   │
│                                     ├── state.json           │
│                                     └── logs.jsonl           │
│                                         alerts.jsonl         │
│                                                              │
│  DailyReporter ────────── rapport quotidien (goroutine)      │
│  WebServer ─────────────── API REST :8080 + embed dashboard  │
│  Dashboard ─────────────── Next.js :3000 (dev)               │
│  CLI ───────────────────── stdin interactif                  │
└─────────────────────────────────────────────────────────────┘
```

---

## Prérequis

| Outil  | Version | Usage                 |
| ------ | ------- | --------------------- |
| Go     | 1.21+   | Runtime + compilation |
| Ollama | latest  | LLM local             |
| Git    | any     | Versioning            |

**Modèles Ollama recommandés :**

| Modèle             | RAM requise | Usage                         |
| ------------------ | ----------- | ----------------------------- |
| `llama3.1:8b`      | ~6 GB       | Recommandé — bon équilibre    |
| `qwen2.5:7b`       | ~5 GB       | Rapide, très bon en JSON      |
| `qwen2.5-coder:7b` | ~5 GB       | Si tu ajoutes des agents code |
| `mistral:7b`       | ~5 GB       | Alternative légère            |

---

## Installation

### 1. Cloner le projet

```bash
git clone https://github.com/Inkeki9/JARVINx.git
cd JARVINx
```

### 2. Installer les dépendances Go

```bash
go mod tidy
```

### 3. Configurer les variables d'environnement

Crée un fichier `runtime/.env` (voir la section Configuration pour toutes les variables) :

```env
# Modèle Ollama
JARVINX_MODEL=llama3.1:8b

# Discord webhook (optionnel — alertes désactivées si absent)
DISCORD_WEBHOOK=https://discord.com/api/webhooks/TON_ID/TON_TOKEN
```

### 4. Lancer Ollama et puller un modèle

```bash
ollama pull llama3.1:8b
ollama serve
```

### 5. Lancer JARVINx

```powershell
# Runtime Go (API :8080)
cd runtime
.\run.ps1           # Windows — charge .env et lance
# ou
make run

# Dashboard Next.js (UI :3000) — dans un second terminal
cd dashboard
npm install
npm run dev
```

Dashboard accessible sur `http://localhost:3000`. En production, `make build` produit un binaire Go qui embarque le build Next.js sur `:8080`.

---

## Déploiement

### Windows

**Option A — PowerShell (recommandé pour le dev)**

Crée `run.ps1` à la racine :

```powershell
# Charge le .env
Get-Content .env | ForEach-Object {
    if ($_ -match '^([^=]+)=(.*)$') {
        [System.Environment]::SetEnvironmentVariable($matches[1], $matches[2])
    }
}

# Lance JARVINx
go run cmd/main.go
```

```powershell
.\run.ps1
```

**Option B — variables manuelles**

```powershell
$env:DISCORD_WEBHOOK="https://discord.com/api/webhooks/..."
go run cmd/main.go
```

**Option C — binaire compilé**

```powershell
go build -o jarvinx.exe cmd/main.go
$env:DISCORD_WEBHOOK="..."
.\jarvinx.exe
```

**Notes Windows :**

- Le chemin disque par défaut est `C:\` — modifiable dans `config/config.go`
- Les commandes whitelistées (`uptime`, `df -h`, `free -h`) ont des équivalents Windows automatiques via `windowsAliases` dans `tools/shell.go`
- Ollama doit tourner en arrière-plan (`ollama serve`)

---

### Linux / macOS

**Option A — lancement direct**

```bash
# Charge le .env et lance
export $(cat .env | xargs)
go run cmd/main.go
```

**Option B — binaire compilé**

```bash
go build -o jarvinx cmd/main.go
chmod +x jarvinx
export DISCORD_WEBHOOK="https://discord.com/api/webhooks/..."
./jarvinx
```

**Option C — cross-compilation depuis Windows vers Linux**

```powershell
$env:GOOS="linux"
$env:GOARCH="amd64"
go build -o jarvinx-linux cmd/main.go
```

Transfère le binaire sur ta machine Linux et lance :

```bash
chmod +x jarvinx-linux
DISCORD_WEBHOOK="..." ./jarvinx-linux
```

**Option D — systemd service (Linux production)**

Crée `/etc/systemd/system/jarvinx.service` :

```ini
[Unit]
Description=JARVINx Autonomous Agent Runtime
After=network.target ollama.service

[Service]
Type=simple
User=ton_user
WorkingDirectory=/opt/jarvinx
ExecStart=/opt/jarvinx/jarvinx
Environment=DISCORD_WEBHOOK=https://discord.com/api/webhooks/...
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable jarvinx
sudo systemctl start jarvinx
sudo systemctl status jarvinx
```

**Notes Linux/macOS :**

- Le chemin disque par défaut est `/` — détecté automatiquement
- `free -h`, `df -h`, `uptime` fonctionnent nativement
- `docker ps` disponible si Docker est installé

---

## Configuration

Toute la configuration est dans `config/config.go` :

```go
func Default() *Config {
    return &Config{
        // Runtime
        Interval: 15 * time.Second,   // Fréquence d'observation

        // LLM
        OllamaURL: "http://localhost:11434",
        Model:     "llama3.1:8b",

        // Fichiers
        LogFile:   "logs.jsonl",
        StateFile: "state.json",
        AlertFile: "alerts.jsonl",

        // Web
        WebPort: 8080,

        // Seuils d'alerte
        CPUAlertThreshold:  85.0,  // % CPU
        RAMAlertThreshold:  90.0,  // % RAM
        DiskAlertThreshold: 85.0,  // % Disk

        // Comportement alertes
        AlertCooldown:  5,   // cycles entre deux alertes identiques
        AlertMinCycles: 2,   // cycles consécutifs avant d'alerter
    }
}
```

### Variables d'environnement

| Variable                   | Description                                | Défaut                   |
| -------------------------- | ------------------------------------------ | ------------------------ |
| `DISCORD_WEBHOOK`          | URL webhook Discord                        | —                        |
| `JARVINX_DEBUG`            | Logs DEBUG (`true/false`)                  | `false`                  |
| `JARVINX_DRY_RUN`          | Mode simulation (`true/false`)             | `false`                  |
| `JARVINX_ALLOWED_ORIGINS`  | Origins CORS supplémentaires (virgule)     | —                        |
| `JARVINX_MODEL`            | Modèle Ollama                              | `llama3.1:8b`            |
| `JARVINX_OLLAMA_URL`       | URL Ollama                                 | `http://localhost:11434` |
| `JARVINX_INTERVAL`         | Intervalle d'observation (`15s`, `1m`)     | `15s`                    |
| `JARVINX_CPU_THRESHOLD`    | Seuil alerte CPU (%)                       | `85`                     |
| `JARVINX_RAM_THRESHOLD`    | Seuil alerte RAM (%)                       | `90`                     |
| `JARVINX_DISK_THRESHOLD`   | Seuil alerte Disk (%)                      | `85`                     |
| `JARVINX_ALERT_COOLDOWN`   | Cycles entre deux alertes identiques       | `5`                      |
| `JARVINX_ALERT_MIN_CYCLES` | Cycles consécutifs avant alerte CPU/RAM    | `2`                      |
| `JARVINX_PORT`             | Port dashboard web                         | `8080`                   |
| `JARVINX_LOG_MAX_MB`       | Taille max logs.jsonl en MB                | `10`                     |
| `JARVINX_LOG_MAX_BACKUPS`  | Nombre de backups logs                     | `3`                      |
| `JARVINX_DOCKER_ENABLED`   | Active le DockerAgent (`true/false`)       | `true`                   |
| `JARVINX_DOCKER_WATCH`     | Containers à surveiller (virgule)          | tous                     |
| `JARVINX_FILE_ENABLED`     | Active le FileAgent (`true/false`)         | `true`                   |
| `JARVINX_FILE_WATCH`       | Dossiers à surveiller (virgule)            | —                        |
| `JARVINX_FILE_MAX_MB`      | Taille max fichier avant alerte (MB)       | `500`                    |
| `SLACK_WEBHOOK`            | URL webhook Slack                          | —                        |
| `NTFY_URL`                 | URL serveur Ntfy                           | `https://ntfy.sh`        |
| `NTFY_TOPIC`               | Topic Ntfy                                 | `jarvinx`                |
| `GOTIFY_URL`               | URL serveur Gotify                         | —                        |
| `GOTIFY_TOKEN`             | Token Gotify                               | —                        |
| `JARVINX_DAILY_REPORT`     | Active le rapport quotidien (`true/false`) | `false`                  |
| `JARVINX_REPORT_HOUR`      | Heure d'envoi du rapport (0-23)            | `8`                      |
| `JARVINX_REPORT_MINUTE`    | Minute d'envoi du rapport (0-59)           | `0`                      |
| `JARVINX_EXEC_COOLDOWN`    | Cooldown entre deux exécutions identiques  | `5m`                     |
| `JARVINX_SQLITE_PATH`      | Chemin SQLite (vide = JSON seul)           | —                        |
| `JARVINX_QDRANT_URL`       | URL Qdrant — active la mémoire sémantique  | — (opt-in)               |
| `JARVINX_EMBED_MODEL`      | Modèle Ollama pour les embeddings          | `nomic-embed-text`       |

---

## Tests

```bash
# Tous les tests
go test ./...

# Par package
go test ./llm/... -v
go test ./agents/... -v

# Avec coverage
go test ./... -cover
```

| Package         | Tests                                                            | Couverture            |
| --------------- | ---------------------------------------------------------------- | --------------------- |
| `llm`           | 36 tests — parser JSON, markdown, fallback, uppercase, malformed | Parser robuste        |
| `agents`        | 44 tests — seuils, cooldown, enable/disable, panic isolation     | AlertAgent + Registry |
| `tools`         | 8 tests — whitelist, timeout, commandes valides                  | Shell executor        |
| `config`        | 29 tests — seuils, intervalle, port, champs vides                | Validation config     |
| `jxlog`         | 9 tests — niveaux, filtrage debug, nil safety                    | Logger structuré      |
| `memory`        | 11 tests                                                         |
| `web`           | 15 tests                                                         |
| `core`          | 23 tests                                                         |
| `dashboard/lib` | 18 tests                                                         |
| **Total**       | **~193 tests**                                                   |

**Ce qui est testé :**

- Parser LLM — 8 cas dont JSON malformé, backticks markdown, action uppercase, champs manquants
- AlertAgent — seuils CPU/RAM/Disk, cooldown anti-spam, reset sur descente, niveaux warning/critical
- Registry — register, enable/disable, agent skippé si désactivé, isolation panic, status RunCount
- Health check Ollama — serveur de test httptest, modèle manquant, offline, erreur 500
- Shell executor — whitelist stricte, timeout context, commandes valides
- Config validation — seuils hors range, intervalle invalide, port, champs vides, erreurs multiples
- jxlog — niveaux INFO/WARN/ERROR/DEBUG, filtrage, nil safety, CaptureOutput

---

## Dashboard web

**Dev :** `http://localhost:3000` (Next.js) · API Go sur `http://localhost:8080`
**Prod :** `http://localhost:8080` (binaire Go embarque le build Next.js)

**Pages :**

| Page        | URL            | Description                                                                        |
| ----------- | -------------- | ---------------------------------------------------------------------------------- |
| Overview    | `/`            | Métriques live, cycle agent, feed décisions, analyse IA, statut execute guard      |
| Agents      | `/agents`      | Registry — health, runs, erreurs, enable/disable à chaud                           |
| Containers  | `/containers`  | Tableau Docker live, filtres All/Running/Exited, badge topbar                      |
| History     | `/history`     | Tableau cycles — badge confiance coloré (execute/suggest), trigger info, métriques |
| LLM Context | `/llm-context` | Tendances CPU/RAM/Disk, action dominante, taux d'alerte                            |
| Settings    | `/settings`    | Config runtime, seuils, endpoints API cliquables                                   |

**Endpoints API :**

| Méthode | Endpoint                 | Description                                                      |
| ------- | ------------------------ | ---------------------------------------------------------------- |
| GET     | `/api/status`            | État runtime, uptime, circuit breaker, dernier cycle, exec_guard |
| GET     | `/api/history`           | 10 derniers cycles (plus récent en premier)                      |
| GET     | `/api/agents`            | Liste agents avec runs/erreurs/enabled                           |
| POST    | `/api/agents/toggle`     | Active/désactive un agent — body: `{"name": "..."}`              |
| GET     | `/api/docker`            | Containers avec running/exited counts                            |
| GET     | `/api/logs/status`       | Taille et état des fichiers de logs                              |
| GET     | `/api/file`              | Status FileAgent + chemins surveillés                            |
| GET     | `/api/daily-report`      | Horaire rapport + dernier/prochain envoi                         |
| POST    | `/api/daily-report/send` | Déclenche un rapport immédiat                                    |
| GET     | `/api/llm-context`       | Contexte adaptatif transmis au LLM                               |

---

## CLI interactive

Pendant que JARVINx tourne, tape directement dans le terminal :

```
help                   Liste les commandes
status                 Dernier snapshot système
history [n]            N derniers cycles (défaut: 5)
interval <secondes>    Change l'intervalle à chaud (min: 5s)
clear                  Efface le terminal
```

---

## Commandes whitelistées

JARVINx n'exécute **que** les commandes de cette liste — aucun shell arbitraire :

| Commande       | Description                     |
| -------------- | ------------------------------- |
| `docker ps`    | Liste les containers actifs     |
| `docker stats` | Stats containers (no-stream)    |
| `uptime`       | Temps de fonctionnement système |
| `df -h`        | Espace disque                   |
| `free -h`      | Mémoire disponible              |

Pour ajouter une commande : `tools/shell.go` → `allowedCommands`.

---

## Sécurité — Docker socket

Le `DockerAgent` communique avec Docker via son socket Unix `/var/run/docker.sock` sur Linux/macOS, ou via TCP `localhost:2375` sur Windows (Docker Desktop).

**⚠️ Avertissement important :** accéder au socket Docker équivaut à avoir les droits `root` sur l'hôte. Sur Linux, l'utilisateur qui lance JARVINx doit appartenir au groupe `docker` :

```bash
sudo usermod -aG docker $USER
```

En contexte containerisé, monter `/var/run/docker.sock` donne un accès root effectif au host — ne jamais faire ça en production multi-tenant.

**Windows :** Docker Desktop doit exposer le daemon sur TCP. Dans Docker Desktop → Settings → General → coche "Expose daemon on tcp://localhost:2375 without TLS".

---

## Alertes Discord

JARVINx envoie des embeds Discord structurés quand un seuil est dépassé.

**Logique anti-spam :**

- CPU / RAM : alerte seulement après `AlertMinCycles` cycles **consécutifs** au-dessus du seuil — les pics isolés (chargement LLM) n'alertent pas
- Disk : alerte à chaque `AlertCooldown` cycles si le seuil est dépassé
- Cooldown configurable entre deux alertes identiques

**Niveaux :**

- `⚠️ WARNING` — seuil dépassé, surveillance requise (jaune)
- `🚨 CRITICAL` — situation dégradée persistante (rouge)

---

## Fichiers générés

| Fichier        | Format       | Description                                                                 |
| -------------- | ------------ | --------------------------------------------------------------------------- |
| `state.json`   | JSON indenté | Historique cycles + snapshots — persiste entre redémarrages (20 derniers)   |
| `logs.jsonl`   | JSONL        | Une ligne par observation                                                   |
| `alerts.jsonl` | JSONL        | Historique de toutes les alertes déclenchées                                |
| `jarvinx.db`   | SQLite       | Historique illimité (snapshots + cycles) — activé via `JARVINX_SQLITE_PATH` |

---

## Changelog

| Version | Theme                     | Status      |
| ------- | ------------------------- | ----------- |
| V1.0    | Stable & Deployable       | ✅ Released |
| V1.1    | Hardening & Corrections   | ✅ Released |
| V1.2    | Correction & Robustesse   | ✅ Released |
| V1.3    | Intelligence & Mémoire    | ✅ Released |
| V1.4    | Robustesse Runtime        | ✅ Released |
| V1.5    | Dashboard                 | ✅ Released |
| V1.6    | Couche décisionnelle      | ✅ Released |
| V1.7    | Mémoire historique SQLite | ✅ Released |
| V1.8    | Mémoire sémantique Qdrant | ✅ Released |
| V1.9    | Sécurité & hardening      | 🔮 Future   |
| V1.10   | Enrichissement agents     | 🔮 Future   |
| V1.11   | Dashboard & visibilité    | 🔮 Future   |
| V1.12   | Robustesse & tests        | 🔮 Future   |

## Roadmap

### v1.0 — Stable & Deployable ✅

- [x] Makefile — `make run`, `make build`, `make build-linux`
- [x] `.env` chargé automatiquement au démarrage
- [x] Chemin disque détecté automatiquement selon l'OS
- [x] Parser LLM robuste — retries, validation schema, fallback
- [x] Interface Agent générique — BaseAgent, Registry, panic isolation
- [x] 26 tests unitaires — parser, alertes, registry
- [x] embed.FS — dashboard en fichiers HTML/CSS/JS séparés
- [x] Shutdown propre — SIGINT/SIGTERM, context annulable
- [x] README complet

### v1.1 — Hardening v1 ✅

- [x] Fix race conditions (State + AlertAgent)
- [x] Structured logging slog / jxlog
- [x] Timeout shell executor
- [x] Validation config au démarrage
- [x] go test -race clean
- [x] Health check Ollama au démarrage — fail fast avec message clair si le LLM est absent

### V1.2 — Correction & Robustesse ✅

- [x] Permissions 0600 — state.json + alerts.jsonl
- [x] CORS whitelist explicite
- [x] Shell dispatch direct sans sh -c
- [x] AlertAgent recordError → recordSuccess
- [x] schedule_ms → nanosecondes corrigées
- [x] Scheduler → time.Ticker + context
- [x] Migration logging complète vers jxlog
- [x] Bus → vrai pub/sub fan-out
- [x] Endpoint toggle agents
- [x] Version unifiée via ldflags
- [x] handleIndex → 404 propre
- [x] govulncheck
- [x] Tests core/ (Orchestrator, Bus, Scheduler)
- [x] Backoff exponentiel polling dashboard
- [x] Tests dashboard hooks + composants

### v1.3 — Intelligence & Mémoire ✅

- [x] **Config via env vars** — (seuils overridables)
- [x] **Rotation des logs** — logs.jsonl sans borne = disk full sur machine faible
- [x] **Mode `--dry-run`** — pour tester sans que l'agent exécute des commandes réelles
- [x] **DockerAgent** — surveillance des containers, détection de crashes, suggestion de restarts
- [x] **FileAgent** — surveillance de dossiers, détection de fichiers lourds, analyse d'espace
- [x] **Multi-webhook** — support Slack, Ntfy, Gotify en plus de Discord
- [x] **Rapport quotidien** — résumé automatique envoyé à heure fixe
- [x] **Prompt adaptatif** — le system prompt évolue selon l'historique des décisions

### v1.4 — Robustesse runtime ✅

- [x] **Supprimer** webhookURL mort DockerAgent
- [x] **Documenter Docker socket** README + TECHNICAL
- [x] **Validation** URLs webhooks au démarrage — `url.Parse()` + schéma
- [x] **Validation** paths FileAgent — blocklist chemins sensibles
- [x] **Exposer** dry_run dans StatusResponse
- [x] **GET** /api/logs/status — taille, backups, rotation
- [x] **Circuit breaker** OllamaClient — open/half-open/closed
- [x] **Bus dispatcher** Bus goroutine dédiée — `Publish()` non-bloquant hors verrou
- [x] CI `-race` clean — 0 data race détectée
- [x] **Test intégration** end-to-end — 6 nouveaux tests
- [x] **Store mémoire** longue durée SQLite vs BBolt design doc

### V1.5 — Dashboard ✅

- [x] **Badge Docker** — topbar running/total, rouge si containers down
- [x] **Page Containers** — tableau live, filtres All/Running/Exited, branché sur `/api/docker`
- [x] **Page LLM Context** — tendances CPU/RAM/Disk, action dominante, taux d'alerte
- [x] **Widget DailyReporter** — last_sent + trigger manuel depuis le dashboard
- [x] **Bloc Analyse IA** — résumé LLM de l'état global, affiché si Ollama connecté
- [x] **Nouveaux endpoints** — `/api/file`, `/api/daily-report`, `/api/daily-report/send`, `/api/llm-context`

### V1.6 — Couche décisionnelle ✅

- [x] **P0 Cooldown execute** `executeGuard` empêche la même commande de s'exécuter en boucle (défaut 5min, `JARVINX_EXEC_COOLDOWN`)
- [x] **P1 Seuils dynamiques** `BuildSystemPrompt` reçoit les seuils réels de config — plus de valeurs hardcodées dans le prompt LLM
- [x] **P2 Score de confiance** `Decision.Confidence` : `execute` rétrogradé en `suggest` si confidence < 0.75 ; absent = 0.5
- [x] **P2 Verify avant execute** `shouldExecute()` re-vérifie les métriques trigger avant d'agir ; annulation si normalisé (±5pt)
- [x] **P2 Audit trail** `TriggerCPU/RAM/Disk` + `Confidence` persistés dans `CycleRecord` / `state.json`
- [x] **P3 OS dans le prompt** `runtime.GOOS` injecté dans `BuildSystemPrompt` ; note traduction automatique Windows
- [x] **Dashboard History** Badge confiance coloré (vert ≥ 0.75 / orange ≥ 0.5 / rouge) + sous-texte "Déclenché à CPU X%"
- [x] **Dashboard Overview** Statut execute guard en temps réel — cooldown actif ou disponible
- [x] **API exec_guard** `last_cmd` + `cooldown_remaining_seconds` dans `GET /api/status`

### V1.7 — Mémoire historique SQLite ✅

- [x] **Phase 0 — Interface Store** — `memory.Store` + `memory.EventLog` ; `AgentContext` découplé des types concrets ; `Add()`/`AddCycle()` retournent `error`
- [x] **Phase 1 — SQLite double write** — `SQLiteStore` (historique illimité, WAL, pure Go via `modernc.org/sqlite`) ; `DoubleWriteStore` (JSON source de vérité + SQLite secondary fail-silencieux) ; `NoopStore` fallback ; config `JARVINX_SQLITE_PATH`
- [x] **Phase 2 — Bascule lecture + Dashboard** — lecture depuis SQLite (`LastCycles(5760)` < 3s validé en CI) ; `GET /api/history/full?range=7d|30d|90d` avec agrégation SQL par bucket heure/6h/jour ; graphes CPU/RAM/Disk (Recharts AreaChart) + sélecteur période sur la page History

## 🧠 V1.8 — Mémoire sémantique Qdrant ✅

> Les décisions passées alimentent les décisions futures.

- [x] `JARVINX_QDRANT_URL` config — Activation opt-in — `QdrantAgent` enregistré seulement si la var est définie
- [x] `QdrantAgent` dans le Registry — vectorise chaque décision LLM via Ollama embeddings
- [x] Embedding des décisions — Texte : `"[{action}] {analysis}. {reason}"` + metadata (confidence, triggers, CycleNum)
- [x] CircuitBreaker embedding — Fail-silencieux si Ollama sous charge — le cycle 15s n'est jamais bloqué
- [x] Contexte enrichi — décisions similaires passées injectées dans le prompt LLM via `SimilarDecisionsProvider`
- [x] Corrélation / patterns — Backlog non daté — à affiner quand le RAG tourne

## 🔒 V1.9 — Sécurité & hardening

> Consolider avant d'ouvrir les features. Aucune nouvelle fonctionnalité — uniquement sécurité et dette critique.

- [ ] **Auth API** — middleware Bearer token sur tous les `/api/*` via `JARVINX_API_TOKEN`
- [ ] **Versioning `state.json`** — champ `version int` + `migrateFrom()` dans `memory/state.go`
- [ ] **CLI → `memory.Store`** — remplacer `*memory.State` par l'interface dans `core/cli.go`
- [ ] **Fix DailyReporter double send** — `sync.Mutex` dans `send()` lui-même
- [ ] **Rate limiting POST** — token bucket 1 req/s sur `/api/daily-report/send` et `/api/agents/toggle`
- [ ] **`validateFilePath`** — préfixes sensibles

- [ ] **Sanitisation prompt injection Qdrant** — encadrer `SimilarDecisions` comme `[HISTORICAL DATA]` dans `context_builder.go`, tronquer 200 chars, strip newlines
- [ ] **Headers HTTP sécurité** — `X-Frame-Options`, `X-Content-Type-Options`, `Content-Security-Policy` dans `corsMiddleware`
- [ ] **QdrantAgent point ID** — remplacer `CycleNum` par `hash(instance_id:cycle_num)` pour éviter collisions au reset SQLite
- [ ] **Body limit XS** — `http.MaxBytesReader(w, r.Body, 4096)` sur les POST
- [ ] **`memory/state_test.go`** — cap à 20, `Save()`/`Load()` roundtrip, `CurrentCycle()`. Compatible Windows
- [ ] **`core/orchestrator_test.go`** — N-1 pattern end-to-end, executeGuard cooldown, shouldExecute annulation

- [ ] **Jitter retry LLM** — `rand.Intn(1000)ms` dans `DefaultRetryConfig`
- [ ] **Token Gotify dans header** — `X-Gotify-Key` au lieu de query param URL
- [ ] **Ntfy désactivé par défaut** — `NtfyURL: ""` pour éviter fuite métriques sur topic public
- [ ] **Webhooks HTTPS uniquement** — rejeter `http://` dans `validateWebhookURL`
- [ ] **Corriger TECHNICAL.md** — `/api/chat`, `Add()`, cap 20 cycles, noms variables réels, commandes Windows

## ⚡ V1.10 — Enrichissement agents

> Améliorer la qualité des décisions LLM sans changer l'architecture.

System Agent — nouvelles métriques gopsutil :

- [ ] **Swap** — `mem.SwapMemory()` dans `tools/system.go` + `Snapshot`. Effort XS
- [ ] **Load average** — `load.Avg()`, fail-silent sur Windows. Effort XS
- [ ] **Top 5 processus** — `process.Processes()` avec timeout 2s. Effort M
- [ ] **Réseau delta** — débit MB/s via delta `net.IOCounters()`. Effort S

Docker Agent — sans appel stats supplémentaire

- [ ] **Health check** — parser `"unhealthy"` dans le champ `Status` déjà retourné par Docker. Effort XS
- [ ] **Restart count** — regex `"Restarted N times"` dans `Status`. Effort XS
- [ ] **Container age** — champ `Created` (unix timestamp) déjà dans la réponse brute. Effort XS

File Agent

- [ ] **Taux de croissance MB/min** — diviser `growth` par durée du cycle. Effort XS
- [ ] **ModTime récent** — `fi.ModTime()` gratuit dans `filepath.Walk`. Effort XS
- [ ] **Inodes Linux** — `disk.Usage(path).InodesUsed` via gopsutil, fail-silent sur Windows. Effort XS

Prompt adaptatif — zero appel réseau

- [ ] **Heure + période** — nuit/matin/journée/soirée + weekend dans `context_builder.go`. Effort XS
- [ ] **Corrélations CPU/RAM** — diagnostic pré-calculé : "RAM monte CPU stable → memory leak potentiel". Effort XS
- [ ] **Streak stabilité** — cycles depuis dernière alerte injecté dans le prompt. Effort XS
- [ ] **Forecast N cycles** — projection linéaire vers le seuil configuré. Effort S

## 📊 V1.11 — Dashboard & visibilité

> Exposer ce qui existe déjà mais n'est pas visible.

- [ ] **Circuit breaker visible** — `circuit_state` dans topbar ou Overview. Zero nouveau endpoint, 20 lignes
- [ ] **Page Alerts** — exposer `alerts.jsonl` via `GET /api/alerts` + page `/alerts` avec filtres Warning/Critical
- [ ] **Corrélations colorées LLM Context** — remplacer le texte de tendance par des badges avec flèches colorées
- [ ] **Sparklines History** — mini-courbe SVG inline par ligne de la table History
- [ ] **Forecast card Overview** — afficher le forecast RAM/CPU si disponible dans `/api/llm-context`
- [ ] **Live terminal execute** — afficher le résultat de la dernière commande exécutée dans l'Overview

## 🧪 V1.12 — Robustesse & tests

> Fermer les trous de couverture et préparer l'architecture pour la v2.0.

Tests manquants critiques

- [ ] **`memory/sqlite_store_test.go`** — `SnapshotBuckets` buckets 1h/6h/24h, AVG correct, slice vide
- [ ] **`dashboard/lib/__tests__/hooks.test.ts`** — cleanup unmount, backoff exponentiel, cancelled flag
- [ ] **`web/server_test.go`** — `/api/history/full`, `/api/history`, `/api/agents`, `/api/docker`

Architecture

- [ ] **Bus configurable** — `JARVINX_BUS_BUFFER` env var + `drops_count` dans `/api/status`
- [ ] **Transactions SQLite** — `Add()` et `AddCycle()` dans une transaction pour cohérence garantie

### Vision v2.0 — Universal Agent Platform

> Au-delà du homelab. JARVINx devient une plateforme
> d'orchestration agentique généraliste.

**Infrastructure**

- Proxmox, VMware, Kubernetes — surveillance et self-healing
- Réseau local — scan, anomalies, intrusions
- NAS, stockage — optimisation automatique

**Productivity**

- Agent email — triage, résumés, réponses suggérées
- Agent calendrier — optimisation de planning
- Agent fichiers — organisation, déduplication, archivage

**Development**

- Agent CI/CD — analyse des pipelines, détection de régressions
- Agent code review — suggestions automatiques
- Agent logs — parsing intelligent, détection d'erreurs

**Data & Research**

- Agent web research — veille automatique sur des topics
- Agent RSS/news — résumés quotidiens personnalisés
- Agent knowledge base — construction d'une base de connaissance locale

**Le fil conducteur :**
Chaque agent suit le même pattern.
Chaque agent parle le même langage (interface Agent).
Tout tourne local. Tout reste privé.

- Ajouts recommandés :
  · Plugin system dynamique
  · Multi-instance coordination
  · API d'administration REST complète
  · TLS + auth dashboard
  · Audit trail

---

## 📌 Décisions d'architecture prises

| Date       | Décision                                                      | Contexte                                                        |
| ---------- | ------------------------------------------------------------- | --------------------------------------------------------------- |
| 2026-05    | Interface Agent générique + BaseAgent embedding               | Composition Go plutôt qu'héritage                               |
| 2026-05    | `embed.FS` pour le dashboard                                  | Binaire auto-suffisant, zéro déploiement statique               |
| 2026-05    | Une seule dépendance externe (gopsutil)                       | Résistance dans le temps                                        |
| 2026-05-26 | V1.2 intercalée avant V1.5                                    | Corrections audit avant nouvelles features                      |
| 2026-05-26 | Shell dispatch direct sans `sh -c`                            | Défense en profondeur même avec whitelist                       |
| 2026-06-01 | Store mémoire longue durée SQLite vs BBolt                    | Définir le store mémoire longue durée                           |
| 2026-06-01 | Interface `Store` + `EventLog` abstraites                     | Découpler les agents des implémentations concrètes avant SQLite |
| 2026-06-01 | `DoubleWriteStore` — JSON source de vérité + SQLite secondary | Migration sans coupure, rollback possible à chaque étape        |
| 2026-06-01 | `CycleNum` comme ID commun SQLite ↔ Qdrant                    | Éviter un UUID supplémentaire, déjà stable et incrémental       |
| 2026-06-02 | `QdrantAgent` dans le Registry plutôt qu'un subscriber Bus    | Isolation panic gratuite, conditionnel à `JARVINX_QDRANT_URL`   |
| 2026-06-02 | `collReady bool` plutôt que `sync.Once` pour l'init Qdrant    | `sync.Once` fige l'erreur si Qdrant est down au démarrage       |
| 2026-06-02 | Deux circuit breakers indépendants (LLM + embedding)          | Embedding et chat ne se bloquent pas mutuellement               |
| 2026-06-02 | Injection `SimilarDecisions` en N-1 via `AgentContext`        | Cohérent avec le pattern N-1 existant, zéro couplage direct     |
| 2026-06-04 | `v1.x` jusqu'au changement structurel cloud/multi-instance    | La v2.0 est réservée à un vrai saut d'architecture              |

## Contribuer

Le projet est en développement actif. Les PRs sont bienvenues sur :

- Nouveaux agents
- Support de nouveaux modèles Ollama
- Améliorations du dashboard
- Corrections de bugs Windows/Linux

---

## License

Apache 2.0 License

---

<div align="center">

**JARVINx — AI Runtime that thinks, acts and evolves.**

_Built with Go · Powered by Ollama · No cloud required_

</div>
