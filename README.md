<div align="center">

```
     ██╗ █████╗ ██████╗ ██╗   ██╗██╗███╗   ██╗██╗  ██╗
     ██║██╔══██╗██╔══██╗██║   ██║██║████╗  ██║╚██╗██╔╝
     ██║███████║██████╔╝██║   ██║██║██╔██╗ ██║ ╚███╔╝
██   ██║██╔══██║██╔══██╗╚██╗ ██╔╝██║██║╚██╗██║ ██╔██╗
╚█████╔╝██║  ██║██║  ██║ ╚████╔╝ ██║██║ ╚████║██╔╝ ██╗
 ╚════╝ ╚═╝  ╚═╝╚═╝  ╚═╝  ╚═══╝  ╚═╝╚═╝  ╚═══╝╚═╝  ╚═╝
Version 1.4.0
```

**Autonomous AI Runtime · Observing. Thinking. Acting. Evolving.**

![Go](https://img.shields.io/badge/Go-1.26.3-00ADD8?style=flat-square&logo=go&logoColor=white)
![Ollama](https://img.shields.io/badge/Ollama-local%20LLM-black?style=flat-square)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
![Status](https://img.shields.io/badge/status-v1.4%20stable-00E5FF?style=flat-square)

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
- **Discord Webhooks** — alertes temps réel
- **HTML/CSS/JS** — dashboard web servi directement par Go

---

## Architecture

```
jarvinx/
│
├── cmd/
│   └── main.go              # Point d'entrée — config + lancement
│
├── core/
│   ├── runtime.go           # Assemblage + shutdown propre (SIGINT/SIGTERM)
│   ├── bus.go               # Bus d'événements (channels Go)
│   ├── scheduler.go         # Ticker — émet les cycles
│   ├── orchestrator.go      # Cerveau — dispatch observe/think/act
│   └── cli.go               # Interface CLI interactive
│
├── agents/
│   ├── agent.go             # Interface Agent + BaseAgent + AgentContext
│   ├── registry.go          # Registry — lifecycle, enable/disable, panic isolation
│   ├── system_agent.go      # Agent LLM — analyse + décision JSON
│   ├── alert_agent.go       # Agent alertes — seuils + Discord
│   ├── alert_test.go        # Tests AlertAgent
│   └── registry_test.go     # Tests Registry
│
├── llm/
│   ├── ollama.go            # Client HTTP Ollama + retries
│   ├── parser.go            # Parser JSON robuste + fallback + validation
│   ├── parser_test.go       # Tests parser (8 cas)
│   └── prompt.go            # Prompt builder (system + user + historique)
│
├── tools/
│   ├── system.go            # Métriques CPU/RAM/Disk — détection OS auto
│   └── shell.go             # Executor whitelist de commandes
│
├── memory/
│   ├── state.go             # Persistance state.json — historique cycles
│   └── logger.go            # Logger JSONL — logs.jsonl / alerts.jsonl
│
├── web/
│   ├── server.go            # HTTP server — API REST
│   ├── embed.go             # embed.FS — fichiers statiques dans le binaire
│   └── static/
│       ├── index.html       # Dashboard HTML
│       ├── style.css        # Styles dark theme
│       └── app.js           # Logique dashboard
│
├── jxlog/
│   ├── log.go               # Structured logging slog — niveaux INFO/WARN/ERROR/DEBUG
│   └── log_test.go          # Tests handler et helpers
│
└── config/
    ├── config.go            # Configuration centralisée
    └── env.go               # Auto-load .env au démarrage
```

### Agent loop

```
┌─────────────────────────────────────────────────────────┐
│                    JARVINX RUNTIME                       │
│                                                          │
│  Scheduler ──tick──► Bus ──► Orchestrator                │
│     (15s)          (chan)        │                       │
│                                 ├── AlertAgent           │
│                                 │   └── Discord Webhook  │
│                                 │                        │
│                                 ├── SystemAgent (LLM)    │
│                                 │   └── Ollama API       │
│                                 │                        │
│                                 ├── Executor (whitelist) │
│                                 │                        │
│                                 └── Memory               │
│                                     ├── state.json       │
│                                     └── logs.jsonl       │
│                                                          │
│  WebServer ──────────────────── Dashboard :8080          │
│  CLI ────────────────────────── stdin interactif         │
└─────────────────────────────────────────────────────────┘
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

Crée un fichier `.env` à la racine :

```env
# Discord webhook (optionnel — alertes désactivées si absent)
DISCORD_WEBHOOK=https://discord.com/api/webhooks/TON_ID/TON_TOKEN
```

### 4. Lancer Ollama et puller un modèle

```bash
ollama pull llama3.1:8b
ollama serve
```

### 5. Lancer JARVINx

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
| `llm`           | 26 tests — parser JSON, markdown, fallback, uppercase, malformed | Parser robuste        |
| `agents`        | 44 tests — seuils, cooldown, enable/disable, panic isolation     | AlertAgent + Registry |
| `tools`         | 8 tests — whitelist, timeout, commandes valides                  | Shell executor        |
| `config`        | 22 tests — seuils, intervalle, port, champs vides                | Validation config     |
| `jxlog`         | 9 tests — niveaux, filtrage debug, nil safety                    | Logger structuré      |
| `memory`        | 8 tests                                                          |
| `web`           | 12 tests                                                         |
| `core`          | 17 tests                                                         |
| `dashboard/lib` | 18 tests                                                         |
| **Total**       | **150 tests**                                                    |

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

Le dashboard est accessible à `http://localhost:8080` dès le lancement.

**Fonctionnalités :**

- Métriques CPU / RAM / Disk en temps réel (refresh 5s)
- Dernière décision de l'agent avec analyse et raison
- Console style macOS avec logs live
- Historique des 10 derniers cycles avec badges d'action
- Agent loop visuel (Observe → Think → Decide → Act → Sleep)
- Runtime info (modèle, intervalle, cycle, uptime)

**Endpoints API :**

```
GET /                   → Dashboard HTML
GET /api/status         → Dernier cycle + métriques actuelles
GET /api/history        → 10 derniers cycles complets
GET /api/agents         → Registry agents + statuts
POST /api/agents/toggle → Active/désactive un agent à chaud
```

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

| Fichier        | Format       | Description                                                 |
| -------------- | ------------ | ----------------------------------------------------------- |
| `state.json`   | JSON indenté | Historique cycles + snapshots — persiste entre redémarrages |
| `logs.jsonl`   | JSONL        | Une ligne par observation                                   |
| `alerts.jsonl` | JSONL        | Historique de toutes les alertes déclenchées                |

---

## Changelog

| Version | Theme                     | Status         |
| ------- | ------------------------- | -------------- |
| V1.0    | Stable & Deployable       | ✅ Released    |
| V1.1    | Hardening & Corrections   | ✅ Released    |
| V1.2    | Correction & Robustesse   | ✅ Released    |
| V1.3    | Intelligence & Mémoire    | ✅ Released    |
| V1.4    | Robustesse Runtime        | 🔄 In progress |
| V1.5    | Dashboard                 | 📋 Planned     |
| V1.6    | Couche décisionnelle      | 📋 Planned     |
| V1.x    | Mémoire sémantique Qdrant | 🔮 Future      |

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

### v1.4 — Robustesse runtime

- [ ] **Supprimer** webhookURL mort DockerAgent
- [ ] **Documenter Docker socket** README + TECHNICAL
- [ ] **Validation** URLs webhooks au démarrage
- [ ] **Validation** paths FileAgent
- [ ] **Exposer** dry_run dans StatusResponse
- [ ] **GET** /api/logs/status
- [ ] **Circuit breaker** OllamaClient
- [ ] **Bus dispatcher** goroutine dédiée
- [ ] **Race conditions** CLAUDE.md
- [ ] **Test intégration** end-to-end
- [ ] **Store mémoire** longue durée SQLite vs BBolt design doc

### V1.5 — Dashboard

- [ ] **Quick win** badge header Brancher `/api/docker` — page Agents ou widget
- [ ] **Nouveau endpoint** `GET /api/file` - FileAgent status + paths + tailles - `GET /api/daily-report` + `POST /api/daily-report/send`
- [ ] **Features Dashboard** Page "LLM Context" — tendances + prompt adaptatif - Widget DailyReporter — last_sent + trigger manuel - `GET /api/llm-context` — expose `BuildAdaptiveContext()`

### V1.6 — Couche décisionnelle

[ 📋 Planned ]

### v1.8 — Mémoire sémantique

- [ ] **Vector DB** — intégration Qdrant local pour mémoire sémantique longue durée
- [ ] **Mémoire contextuelle** — JARVINx se souvient des événements passés similaires et les cite dans ses décisions
- [ ] **Embedding** des décisions passées
- [ ] **Recherche sémantique** — "qu'est-ce qui s'est passé la dernière fois que le CPU a spiké ?"
- [ ] **Contexte enrichi pour le LLM** — événements similaires passés

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
