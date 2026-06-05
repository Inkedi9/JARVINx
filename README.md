<div align="center">

```
     ██╗ █████╗ ██████╗ ██╗   ██╗██╗███╗   ██╗██╗  ██╗
     ██║██╔══██╗██╔══██╗██║   ██║██║████╗  ██║╚██╗██╔╝
     ██║███████║██████╔╝██║   ██║██║██╔██╗ ██║ ╚███╔╝
██   ██║██╔══██║██╔══██╗╚██╗ ██╔╝██║██║╚██╗██║ ██╔██╗
╚█████╔╝██║  ██║██║  ██║ ╚████╔╝ ██║██║ ╚████║██╔╝ ██╗
 ╚════╝ ╚═╝  ╚═╝╚═╝  ╚═╝  ╚═══╝  ╚═╝╚═╝  ╚═══╝╚═╝  ╚═╝
Version 1.9.0
```

**Autonomous AI Runtime · Observing. Thinking. Acting. Evolving.**

![Go](https://img.shields.io/badge/Go-1.26.3-00ADD8?style=flat-square&logo=go&logoColor=white)
![Ollama](https://img.shields.io/badge/Ollama-local%20LLM-black?style=flat-square)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
![Status](https://img.shields.io/badge/status-v1.9%20stable-00E5FF?style=flat-square)

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
- Les commandes whitelistées (`uptime`, `df -h`, `free -h`) ont des équivalents Windows automatiques via `windowsSpecs` dans `tools/shell.go` (`wmic`, `cmd /C`)
- Les webhooks Discord/Slack/Gotify/Ntfy requièrent `https://` — les URLs `http://` sont rejetées au démarrage
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
| `NTFY_URL`                 | URL serveur Ntfy                           | — (désactivé)            |
| `NTFY_TOPIC`               | Topic Ntfy                                 | — (désactivé)            |
| `GOTIFY_URL`               | URL serveur Gotify                         | —                        |
| `GOTIFY_TOKEN`             | Token Gotify                               | —                        |
| `JARVINX_DAILY_REPORT`     | Active le rapport quotidien (`true/false`) | `false`                  |
| `JARVINX_REPORT_HOUR`      | Heure d'envoi du rapport (0-23)            | `8`                      |
| `JARVINX_REPORT_MINUTE`    | Minute d'envoi du rapport (0-59)           | `0`                      |
| `JARVINX_EXEC_COOLDOWN`    | Cooldown entre deux exécutions identiques  | `5m`                     |
| `JARVINX_SQLITE_PATH`      | Chemin SQLite (vide = JSON seul)           | —                        |
| `JARVINX_QDRANT_URL`       | URL Qdrant — active la mémoire sémantique  | — (opt-in)               |
| `JARVINX_EMBED_MODEL`      | Modèle Ollama pour les embeddings          | `nomic-embed-text`       |
| `JARVINX_API_TOKEN`        | Bearer token pour l'API REST (vide = pas d'auth) | —              |

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
| GET     | `/api/history/full`      | Snapshots agrégés par période (`?range=7d\|30d\|90d`) — SQLite requis |

> **Auth API (v1.9)** — si `JARVINX_API_TOKEN` est défini, toutes les routes `/api/*` requièrent le header `Authorization: Bearer <token>`. Laisser vide pour désactiver (dev local).

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
| V1.9    | Sécurité & hardening      | ✅ Released |
| V1.10   | Enrichissement agents     | 🔮 Future   |
| V1.11   | Dashboard & visibilité    | 🔮 Future   |
| V1.12   | Robustesse & tests        | 🔮 Future   |

## Roadmap

Voir [ROADMAP.md](docs/ROADMAP.md) pour le détail des versions passées et à venir.

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
