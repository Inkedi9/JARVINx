# JARVINx — Documentation Technique

> Pour les contributeurs, développeurs d'agents, et intégrateurs.

---

## Sommaire

1. [Architecture globale](#architecture-globale)
2. [Flux de données complet](#flux-de-données-complet)
3. [Core — le runtime](#core--le-runtime)
4. [Système d'agents](#système-dagents)
5. [Intégration LLM (Ollama)](#intégration-llm-ollama)
6. [Mémoire & persistance](#mémoire--persistance)
7. [Web server & API](#web-server--api)
8. [Outils système](#outils-système)
9. [Écrire un nouvel agent](#écrire-un-nouvel-agent)
10. [Tests](#tests)
11. [Build & déploiement](#build--déploiement)
12. [Limites connues & roadmap technique](#limites-connues--roadmap-technique)

---

## Architecture globale

JARVINx est découpé en 7 packages Go indépendants avec des responsabilités strictement séparées.

```
jarvinx/
├── cmd/        Point d'entrée unique — config + boot
├── core/       Runtime, bus, scheduler, orchestrator, CLI
├── agents/     Interface Agent, Registry, agents concrets
├── llm/        Client Ollama, parser JSON, prompt builder
├── memory/     Persistance state.json, logger JSONL
├── tools/      Métriques système, exécuteur shell whitelist
├── web/        HTTP server, embed.FS, API REST
└── config/     Config centralisée, chargement .env
└── jxlog/      Structured logging — handler slog custom, niveaux, couleurs
```

**Principe de dépendance :** les couches basses (`tools`, `memory`, `config`) ne dépendent de rien d'autre dans le projet. Les couches hautes (`core`, `agents`, `web`) importent les couches basses, jamais l'inverse.

```
core / agents / web
       ↓
     memory
       ↓
     tools / config
```

### Dépendances externes

JARVINx a **une seule** dépendance externe :

```
github.com/shirou/gopsutil/v3
```

Utilisée uniquement dans `tools/system.go` pour la collecte cross-platform de métriques CPU/RAM/Disk.

---

## Flux de données complet

Voici ce qui se passe à chaque cycle (15s par défaut) :

```
1. Scheduler.tick()
      │
      ▼
2. tools.CollectSnapshot()    ← bloque 1s (mesure CPU)
      │ memory.Snapshot{cpu, ram, disk}
      ▼
3. Bus.Publish(EventObserved)
      │
      ▼
4. Orchestrator.handleObserved(snap)
      ├─ Logger.Write(entry)            ← logs.jsonl
      ├─ State update (lastSnap)
      └─ si décision précédente = execute → tools.ExecuteCommand()
      │
      ▼
5. Registry → agents en parallèle (goroutines séparées)
      │
      ├─ SystemAgent.Run()
      │     ├─ llm.BuildPrompt(snap, history)
      │     ├─ ollama.Query(prompt)        ← HTTP vers Ollama
      │     ├─ parser.Parse(response)      ← JSON extraction
      │     └─ state.AddCycle(decision)    ← state.json
      │
      └─ AlertAgent.Run()
            ├─ Analyze(snap)               ← seuils + cooldown
            ├─ logAlert() → alerts.jsonl
            └─ sendDiscord()               ← HTTP webhook (si configuré)
```

---

## Core — le runtime

### `core/runtime.go` — Assembleur

`Runtime` est le point d'entrée principal. Il instancie tous les composants et les relie.

```go
type Runtime struct {
    cfg        *config.Config
    bus        *Bus
    scheduler  *Scheduler
    orch       *Orchestrator
    registry   *agents.Registry
    state      *memory.State
    logger     *memory.Logger
    webServer  *web.Server
    cli        *CLI
}
```

Ordre d'initialisation dans `Start()` :

1. `memory.NewState()` + `memory.NewLogger()` — persistance
2. `agents.NewRegistry()` + registration des agents
3. `core.NewBus()` + `core.NewScheduler()` + `core.NewOrchestrator()`
4. `web.NewServer()` → lancé dans goroutine séparée
5. `registry.Start(ctx)` + `scheduler.Start(ctx)` + `orchestrator.Start(ctx)`
6. `cli.Start()` — bloquant, dans goroutine séparée

Shutdown propre via `signal.NotifyContext(ctx, SIGINT, SIGTERM)` : quand le signal arrive, le context est annulé et toutes les goroutines en `select { case <-ctx.Done() }` s'arrêtent.

### `core/bus.go` — Bus d'événements

Bus central basé sur un channel Go bufferisé.
**Bus** — pub/sub avec fan-out. `Subscribe(name)` retourne un canal dédié par consommateur. `Publish()` broadcast à tous les subscribers simultanément. Buffer plein → warning + drop sans bloquer les autres subscribers. `Unsubscribe(name)` ferme le canal proprement.

```go
type EventType string

const (
    EventObserved EventType = "observed"
    EventExecuted EventType = "executed"
    EventError    EventType = "error"
)

type Event struct {
    Type    EventType
    Payload any
}
```

`Subscribe()` retourne un `<-chan Event`. L'orchestrateur s'y abonne et dispatch selon le type.

> **Limite actuelle :** `bufferSize=10` + drop silencieux si le buffer est plein. Sous charge haute, des cycles peuvent être perdus sans log explicite.

### `core/scheduler.go` — Horloge

Émet un tick `EventObserved` à chaque intervalle. Thread-safe : `SetInterval()` utilise un `sync.RWMutex` pour modifier l'intervalle à chaud depuis la CLI.

```go
func (s *Scheduler) SetInterval(d time.Duration) {
    s.mu.Lock()
    s.interval = d
    s.mu.Unlock()
    // Relance le ticker avec le nouvel intervalle
}
```

### `core/orchestrator.go` — Dispatcher

Reçoit les événements du bus et coordonne les agents. `TryLock()` sur le mutex principal garantit qu'un seul cycle tourne à la fois — si le cycle précédent n'est pas terminé, le tick est ignoré.

```go
func (o *Orchestrator) handleObserved(snap memory.Snapshot) {
    if !o.mu.TryLock() {
        fmt.Println("[ ORCHESTRATOR ] Cycle précédent en cours — tick ignoré")
        return
    }
    defer o.mu.Unlock()
    // ...
}
```

`AgentContext()` expose un snapshot thread-safe aux agents via `snapMu RWMutex`.

### `core/cli.go` — Interface interactive

Scanner stdin en boucle. Commandes : `help`, `status`, `history [n]`, `interval <s>`, `clear`.

---

## Système d'agents

### Interface `Agent`

Tout agent doit implémenter cette interface ([agents/agent.go](../agents/agent.go)) :

```go
type Agent interface {
    Name()     string
    Schedule() time.Duration
    Run(ctx context.Context, actx AgentContext) error
    Status()   AgentStatus
    IsEnabled() bool
    Enable()
    Disable()
}
```

### `BaseAgent` — implémentation commune

`BaseAgent` fournit les méthodes `Enable/Disable/Status/IsEnabled` avec protection mutex intégrée. Il suffit de l'embedder dans ton agent :

```go
type MonAgent struct {
    agents.BaseAgent
    // ... tes champs
}
```

`BaseAgent` maintient automatiquement :

- `runCount` — nombre de runs réussis
- `errorCount` — nombre d'erreurs
- `lastRun` — timestamp du dernier run
- `lastError` — message de la dernière erreur

### `AgentContext`

Passé à `Run()` à chaque cycle, contient tout ce dont un agent peut avoir besoin :

```go
type AgentContext struct {
    Snapshot memory.Snapshot   // Métriques du cycle courant
    State    *memory.State     // Accès à l'historique complet
    Logger   *memory.Logger    // Pour écrire dans logs.jsonl
}
```

### `Registry`

Le Registry gère le lifecycle de tous les agents enregistrés.

```go
registry := agents.NewRegistry()
registry.Register(monAgent)
registry.Start(ctx, orchestrator.AgentContext)
```

`Start()` lance chaque agent dans sa propre goroutine avec son propre ticker. Le cycle de chaque agent est indépendant des autres.

**Isolation des panics :** chaque `Run()` est enveloppé dans un `defer recover()`. Un panic dans un agent n'affecte pas les autres.

```go
defer func() {
    if rec := recover(); rec != nil {
        fmt.Printf("[ REGISTRY ] Panic récupéré dans %s : %v\n", a.Name(), rec)
    }
}()
```

**Enable/Disable à chaud :**

```go
registry.Disable("alert")  // L'AlertAgent ne sera plus appelé
registry.Enable("alert")   // Réactivation
```

### Agents existants

#### `SystemAgent` (`agents/system_agent.go`)

Appelle Ollama avec les métriques + historique, parse la décision JSON, persiste le cycle dans `state.json`.

- Schedule : 15s (suit l'intervalle global)
- Timeout : 60s (context timeout sur l'appel LLM)
- Fallback : action `"log"` si le LLM ne répond pas ou retourne un JSON invalide

#### `AlertAgent` (`agents/alert_agent.go`)

Analyse les seuils, gère le cooldown anti-spam, envoie les embeds Discord.

- Schedule : 15s
- Seuils : CPU 85%, RAM 90%, Disk 85% (configurables)
- Logique CPU/RAM : N cycles consécutifs requis (`AlertMinCycles`)
- Logique Disk : alerte directe avec cooldown uniquement

### DockerAgent (`agents/docker_agent.go`)

Surveille les containers Docker via l'API REST Docker (socket Unix ou TCP Windows).
Schedule : 30s. Pas de dépendance externe — `net/http` standard avec transport custom.

- Détecte les transitions `running → exited` (crash)
- Détecte les redémarrages `exited → running`
- WatchList optionnelle — vide = surveille tout
- Désactivable via `JARVINX_DOCKER_ENABLED=false`

Windows : Docker Desktop doit exposer le port TCP 2375 dans ses settings.

**Sécurité socket Docker**

`tools/docker.go` se connecte à Docker via :

- Linux/macOS : `unix:///var/run/docker.sock`
- Windows : `tcp://localhost:2375` (Docker Desktop, TCP non-TLS)

Accéder au socket Docker = privilèges root effectifs sur l'hôte. À documenter explicitement dans toute politique de déploiement. Ne jamais exposer JARVINx sur un réseau public sans auth si DockerAgent est actif.

### FileAgent (`agents/file_agent.go`)

Scanne les dossiers configurés via `filepath.Walk`.
Schedule : 5 minutes.

- Détecte les fichiers dépassant `FileMaxSizeMB`
- Détecte la croissance rapide d'un dossier entre deux cycles
- Nécessite `JARVINX_FILE_WATCH` — désactivé si vide

### Multi-webhook (`agents/notifier.go`)

Interface `Notifier` — `Name() string` + `Send(alert Alert) error`.
`NotifierDispatcher` broadcast les alertes à tous les notifiers enregistrés.
Un échec sur un notifier n'affecte pas les autres.

Notifiers disponibles : `DiscordNotifier`, `SlackNotifier`, `NtfyNotifier`, `GotifyNotifier`.

Ajouter un notifier custom :

```go
type MyNotifier struct{}
func (n *MyNotifier) Name() string       { return "myservice" }
func (n *MyNotifier) Send(a Alert) error { /* ... */ return nil }

dispatcher.Register(&MyNotifier{})
```

### DailyReporter (`agents/daily_report.go`)

Goroutine indépendante — tick toutes les minutes, envoie à l'heure configurée.
Protection anti-double envoi via `lastSent` (cooldown 2 minutes).
Activé via `JARVINX_DAILY_REPORT=true`.

### Prompt adaptatif (`llm/context_builder.go`)

`BuildAdaptiveContext(cycles, snapshots)` analyse les N derniers cycles et snapshots :

- Action dominante sur la période
- Taux d'alerte (%)
- Tendances CPU/RAM/Disk : stable / en hausse / en baisse / critique
- Dernières alertes déclenchées

`BuildAdaptiveSystemPrompt(base, ctx)` enrichit le system prompt avec ce contexte.
Utilisé automatiquement par `SystemAgent` — aucune config requise.

---

## Intégration LLM (Ollama)

### Client HTTP (`llm/ollama.go`)

```go
func (c *Client) Query(ctx context.Context, prompt Prompt) (string, error)
```

- 3 tentatives avec 2s de délai fixe entre chaque
- Timeout global : 90s (context passé depuis SystemAgent)
- Format de requête : `POST /api/generate` avec `stream: false`
- Parse la réponse : extrait `response` du JSON Ollama

### Parser JSON robuste (`llm/parser.go`)

Le LLM ne retourne pas toujours un JSON propre. Le parser gère :

1. **JSON direct** — cas nominal
2. **JSON dans des backticks markdown** — ` ```json {...} ``` `
3. **JSON enveloppé dans du texte** — extraction via regex `\{[^{}]*\}`
4. **Action en majuscules** — `"LOG"` → `"log"`
5. **Champs manquants** — `analysis` et `reason` ont des valeurs par défaut
6. **Fallback total** — si rien ne marche, retourne `action: "log"` avec un message d'erreur
7. **Bus** — pub/sub avec fan-out. `Subscribe(name)` retourne un canal dédié par consommateur. `Publish()` broadcast à tous les subscribers simultanément. Buffer plein → warning + drop sans bloquer les autres subscribers. `Unsubscribe(name)` ferme le canal proprement.

```go
type Decision struct {
    Analysis string `json:"analysis"`
    Action   string `json:"action"`   // "log" | "alert" | "suggest" | "execute"
    Command  string `json:"command"`  // seulement si action=execute
    Reason   string `json:"reason"`
}
```

Actions valides : `log`, `alert`, `suggest`, `execute`. Toute autre valeur est normalisée vers `log`.

> **Limite actuelle :** la regex `\{[^{}]*\}` ne gère pas le JSON imbriqué. Si un futur agent retourne une structure nested, le parser échouera et tombera sur le fallback.

### Prompt builder (`llm/prompt.go`)

**System prompt** (statique) — définit le rôle de JARVINx, le format JSON attendu, les règles d'action et les commandes autorisées.

**User prompt** (dynamique) — construit à chaque cycle avec :

- L'historique des 5 derniers snapshots (timestamps + CPU/RAM/Disk)
- L'observation courante (valeurs absolues + pourcentages)

---

## Mémoire & persistance

### `memory.State` (`memory/state.go`)

Persistance dans `state.json`. Chargé au démarrage, sauvegardé après chaque cycle.

```go
type State struct {
    CycleNum int            `json:"cycle_num"`
    History  []Snapshot     `json:"history"`   // max 20 entrées
    Cycles   []CycleRecord  `json:"cycles"`    // non borné — voir limites
}
```

**`Snapshot`** — métriques brutes d'un cycle :

```go
type Snapshot struct {
    Timestamp   time.Time
    CPUPercent  float64
    MemUsed     uint64    // en MB
    MemTotal    uint64    // en MB
    MemPercent  float64
    DiskUsed    uint64    // en GB
    DiskTotal   uint64    // en GB
    DiskPercent float64
}
```

**`CycleRecord`** — décision LLM associée à un cycle :

```go
type CycleRecord struct {
    CycleNum  int
    Timestamp time.Time
    Action    string
    Analysis  string
    Reason    string
    Command   string
}
```

Méthodes principales :

- `state.AddSnapshot(snap)` — ajoute + rotate si > 20 entrées
- `state.AddCycle(record)` — ajoute sans borne actuelle
- `state.Last(n)` — retourne les N derniers snapshots
- `state.LastCycles(n)` — retourne les N derniers cycles

### `memory.Logger` (`memory/logger.go`)

Append-only vers `logs.jsonl`. Chaque appel à `Write()` ouvre, écrit et ferme le fichier.

```go
type LogEntry struct {
    Timestamp   time.Time
    CPUPercent  float64
    MemUsed     uint64
    MemTotal    uint64
    MemPercent  float64
    DiskUsed    uint64
    DiskTotal   uint64
    DiskPercent float64
}
```

---

## Web server & API

### Architecture (`web/server.go`)

HTTP server standard library Go, sans framework externe. Les fichiers statiques sont intégrés dans le binaire via `embed.FS` ([web/embed.go](../web/embed.go)).

```go
//go:embed static
var StaticFiles embed.FS
```

Routes :

```
GET /                → index.html (embed.FS)
GET /static/*        → fichiers statiques (CSS, JS)
GET /api/status      → StatusResponse JSON
GET /api/history     → HistoryResponse JSON (10 derniers cycles)
```

### `GET /api/status` — Réponse complète

```go
type StatusResponse struct {
    Online    bool                `json:"online"`
    Model     string              `json:"model"`
    Interval  string              `json:"interval"`
    CycleNum  int                 `json:"cycle_num"`
    Uptime    string              `json:"uptime"`
    LastCycle *memory.CycleRecord `json:"last_cycle,omitempty"`
}
```

### Frontend (`web/static/`)

- `index.html` — structure du dashboard
- `style.css` — dark theme, variables CSS, animations
- `app.js` — polling toutes les 5s vers `/api/status` et `/api/history`, mise à jour du DOM

---

## Outils système

### Métriques (`tools/system.go`)

```go
func CollectSnapshot() (memory.Snapshot, error)
```

Utilise `gopsutil/v3` :

- `cpu.Percent(1*time.Second, false)` — bloque 1s pour la mesure différentielle
- `mem.VirtualMemory()` — RAM totale et utilisée
- `disk.Usage(diskPath)` — espace disque du path configuré

Détection automatique du path disque selon l'OS :

```go
var diskPath = "/"        // Linux / macOS
// Windows : "C:\\"       // via build tag ou détection runtime
```

### Exécuteur shell whitelist (`tools/shell.go`)

```go
func ExecuteCommand(cmd string) CommandResult
```

**Whitelist stricte** — seules ces commandes sont autorisées :

```go
var allowedCommands = map[string]bool{
    "docker ps":    true,
    "docker stats": true,
    "uptime":       true,
    "df -h":        true,
    "free -h":      true,
}
```

Sur Windows, les commandes Unix sont mappées vers leurs équivalents PowerShell via `windowsAliases` :

- `uptime` → `(Get-Date) - (gcim Win32_OperatingSystem).LastBootUpTime`
- `df -h` → `Get-PSDrive`
- `free -h` → `Get-CimInstance Win32_OperatingSystem | Select FreePhysicalMemory,TotalVisibleMemorySize`

Toute commande non listée retourne une erreur `command not allowed` sans exécution.

### Sécurité

**CORS** — whitelist explicite via `allowedOrigins map[string]bool` dans `web/server.go`.
Origins autorisées par défaut : `http://localhost:3000`, `http://localhost:8080`.
Origins supplémentaires configurables via `JARVINX_ALLOWED_ORIGINS=url1,url2` dans `.env`.
Requêtes sans header `Origin` (curl, scripts) passent sans restriction.
Preflight OPTIONS retourne 403 si l'origin n'est pas dans la whitelist.

**Shell executor** — dispatch direct sans shell intermédiaire.
Chaque commande whitelistée est mappée vers un `CommandSpec{bin, args}` — `exec.Command("df", "-h")` et non `sh -c "df -h"`. Injection shell structurellement impossible même si la whitelist est contournée.

**Fichiers** — permissions `0600` sur `state.json`, `logs.jsonl`, `alerts.jsonl`. World-readable uniquement sur Windows (pas de permissions Unix).

### Health check Ollama (`llm/health.go`)

Appelé au démarrage avant le lancement du runtime.

```go
func CheckOllama(baseURL string, model string) HealthStatus
```

- Ping sur `/api/tags` avec timeout 5s
- Vérifie que le modèle configuré est dans la liste des modèles installés
- Retourne `HealthStatus{Online, Models, Error}`
- Si offline → `os.Exit(1)` avec message d'aide
- Si modèle manquant → warning mais démarrage quand même

---

## Logging structuré (jxlog)

JARVINx utilise un handler `slog` custom qui produit des logs colorés avec niveaux.

```go
jxlog.Info("REGISTRY", "Agent enregistré : system")
jxlog.Warn("AGENT", "Fallback après 3 tentatives")
jxlog.Error("STATE", fmt.Sprintf("Save failed: %v", err))
jxlog.Debug("ORCHESTRATOR", "Cycle précédent en cours — tick ignoré")
```

**Niveaux :**

- `DEBUG` — filtré par défaut, activé via `JARVINX_DEBUG=true`
- `INFO` — logs normaux du runtime
- `WARN` — situations dégradées non critiques
- `ERROR` — erreurs qui nécessitent attention

**Couleurs par tag :**

- `REGISTRY`, `SCHEDULER`, `WEB` → cyan
- `ORCHESTRATOR`, `CLI` → bleu
- `SYSTEM AGENT`, `AGENT` → magenta
- `ALERT` → rouge
- `EXEC` → jaune
- `STATE` → gris

## Écrire un nouvel agent

Voici le guide complet pour ajouter un agent. L'exemple : un `NetworkAgent` qui surveille la connectivité réseau.

### Étape 1 — Créer le fichier

```go
// agents/network_agent.go
package agents

import (
    "context"
    "fmt"
    "net/http"
    "time"
)

type NetworkAgent struct {
    BaseAgent
    targetURL  string
    httpClient *http.Client
}

func NewNetworkAgent(targetURL string) *NetworkAgent {
    return &NetworkAgent{
        BaseAgent:  NewBaseAgent("network", 30*time.Second), // schedule indépendant
        targetURL:  targetURL,
        httpClient: &http.Client{Timeout: 5 * time.Second},
    }
}

func (a *NetworkAgent) Run(ctx context.Context, actx AgentContext) error {
    resp, err := a.httpClient.Get(a.targetURL)
    if err != nil {
        a.recordError(err)
        return fmt.Errorf("network check failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 400 {
        err := fmt.Errorf("HTTP %d from %s", resp.StatusCode, a.targetURL)
        a.recordError(err)
        return err
    }

    a.recordSuccess()
    fmt.Printf("[ NETWORK ] %s → %d OK\n", a.targetURL, resp.StatusCode)
    return nil
}
```

### Étape 2 — Enregistrer dans le runtime

Dans [core/runtime.go](../core/runtime.go), ajoute l'agent au registry :

```go
// Après les agents existants
networkAgent := agents.NewNetworkAgent("https://1.1.1.1")
registry.Register(networkAgent)
```

### Étape 3 — (optionnel) Exposer via l'API

Si tu veux que le status de l'agent apparaisse dans `/api/status`, il est automatiquement inclus via `registry.Statuses()` — aucune modification nécessaire.

### Règles à respecter

1. **Toujours embedder `BaseAgent`** — ne pas réimplémenter `Enable/Disable/Status/IsEnabled`
2. **Appeler `recordSuccess()` ou `recordError()`** — les stats sont utilisées pour le monitoring
3. **Respecter le context** — utilise `ctx` dans tes appels HTTP/IO pour répondre au shutdown
4. **Ne pas bloquer indéfiniment** — pose un timeout sur tout appel réseau ou IO
5. **Pas de commandes shell** — utilise `tools.ExecuteCommand()` avec la whitelist, ou ajoute une commande à la whitelist si nécessaire

### Schedule indépendant

Chaque agent a son propre schedule. Un agent de backup peut tourner toutes les heures sans affecter le cycle de 15s du SystemAgent :

```go
NewBaseAgent("backup", 1*time.Hour)
```

---

## Tests

### Lancer les tests

```bash
# Tous les packages
go test ./...

# Avec détail
go test ./agents/... -v
go test ./llm/... -v

# Avec coverage
go test ./... -cover

# Avec détecteur de race conditions (recommandé)
go test ./... -race
```

### Structure des tests existants

#### `agents/alert_test.go` — 14 tests AlertAgent

Couvre : seuils CPU/RAM/Disk, cooldown anti-spam, reset sur descente, niveaux warning/critical, dispatch sans webhook.

Pattern utilisé : construction d'un `AlertAgent` avec des seuils bas (10%) pour déclencher facilement des alertes dans les tests.

#### `agents/registry_test.go` — 4 tests Registry

Couvre : register + get, enable/disable, agent skippé si désactivé, isolation panic.

#### `llm/parser_test.go` — 8 tests Parser

Couvre : JSON direct, JSON avec backticks markdown, action uppercase, champs manquants, JSON malformé, fallback.

### Écrire un test pour ton agent

```go
// agents/network_agent_test.go
package agents

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestNetworkAgent_OK(t *testing.T) {
    // Serveur HTTP de test
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(200)
    }))
    defer srv.Close()

    agent := NewNetworkAgent(srv.URL)
    actx := AgentContext{} // vide pour ce test

    err := agent.Run(context.Background(), actx)
    if err != nil {
        t.Fatalf("expected no error, got: %v", err)
    }

    status := agent.Status()
    if status.RunCount != 1 {
        t.Errorf("expected RunCount=1, got %d", status.RunCount)
    }
}

func TestNetworkAgent_Failure(t *testing.T) {
    agent := NewNetworkAgent("http://localhost:0") // port invalide

    err := agent.Run(context.Background(), AgentContext{})
    if err == nil {
        t.Fatal("expected error, got nil")
    }

    status := agent.Status()
    if status.ErrorCount != 1 {
        t.Errorf("expected ErrorCount=1, got %d", status.ErrorCount)
    }
}
```

---

## Build & déploiement

### Build & version

```bash
# Dev normal
go run cmd/main.go

# Dev dry-run
go run cmd/main.go --dry-run

# Ou via env
JARVINX_DRY_RUN=true go run cmd/main.go
```

La version est injectée au build via `-ldflags "-X main.Version=x.y.z"`.
En dev (`go run`), la variable vaut `"dev"`.

```bash
make build          # jarvinx v1.2.0
make build-linux    # jarvinx-linux v1.2.0
```

`govulncheck` doit être lancé avant chaque release :

```bash
cd runtime && govulncheck ./...
# Expected: No vulnerabilities found.
```

### Commandes Makefile

```bash
make run          # go run cmd/main.go
make build        # go build -o jarvinx (Linux/macOS) ou jarvinx.exe (Windows)
make build-linux  # GOOS=linux GOARCH=amd64 go build -o jarvinx-linux
make test         # go test ./...
make clean        # rm -f jarvinx jarvinx.exe jarvinx-linux
```

### Cross-compilation

Go supporte la cross-compilation sans toolchain externe.

```bash
# Windows → Linux
GOOS=linux GOARCH=amd64 go build -o jarvinx-linux cmd/main.go

# Windows → macOS ARM (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o jarvinx-macos cmd/main.go

# Windows → Windows (depuis Linux)
GOOS=windows GOARCH=amd64 go build -o jarvinx.exe cmd/main.go
```

### Binaire auto-suffisant

Grâce à `embed.FS`, le binaire compilé contient :

- Le runtime Go
- Le dashboard HTML/CSS/JS
- Aucune dépendance externe à déployer

Le seul prérequis sur la machine cible est **Ollama**.

### Variables d'environnement

| Variable                  | Requis                                 | Description                      |
| ------------------------- | -------------------------------------- | -------------------------------- |
| `DISCORD_WEBHOOK`         | Non                                    | URL webhook Discord pour alertes |
| `JARVINX_DEBUG`           | Active les logs DEBUG (`true/false`)   | Non                              |
| `JARVINX_ALLOWED_ORIGINS` | Origins CORS supplémentaires (virgule) | Non                              |

## Configuration via variables d'environnement

Toutes les valeurs de `config/config.go` sont surchargeables via env vars sans recompiler. Ordre de priorité :

Valeur par défaut (config.go)
↓ surchargée par .env
↓ surchargée par variables d'environnement système
↓ surchargée par flags CLI (--dry-run uniquement)

`config.LoadEnv(".env")` charge le fichier `.env`.
`cfg.FromEnv()` applique les variables d'environnement sur la config.
Les valeurs invalides (ex: `JARVINX_CPU_THRESHOLD=abc`) sont ignorées avec un warning — jamais de crash.

## Rotation des logs

`memory.NewLoggerWithRotation(filepath, maxBytes, maxBackups)` — rotation automatique par taille.

Comportement :

- Quand `logs.jsonl` dépasse `maxBytes` → rotation avant l'écriture
- `logs.jsonl` → `logs.jsonl.1` → `logs.jsonl.2` → `logs.jsonl.3` (max)
- Le backup le plus vieux est supprimé quand `maxBackups` est atteint
- `os.Rename` atomique — pas de corruption possible

Configurable via env :

```env
JARVINX_LOG_MAX_MB=10
JARVINX_LOG_MAX_BACKUPS=3
```

## Mode dry-run

Activé via `--dry-run` flag ou `JARVINX_DRY_RUN=true`.

Ce qui est simulé :

- Exécution de commandes shell → log `[DRY-RUN] commande simulée`
- Envoi d'alertes Discord → log `[DRY-RUN] Discord alert simulée`

Ce qui tourne normalement en dry-run :

- Observation système (CPU/RAM/Disk)
- Appel LLM Ollama — décisions réelles
- Écriture dans state.json et logs.jsonl
- Dashboard web

Utile pour : tester une nouvelle config de seuils, valider un déploiement, débugger sans effets de bord.

---

## Limites connues & roadmap technique

### Limites actuelles (v1.0)

**Concurrence**

- `memory/state.go` : `History` et `Cycles` accédés sans mutex — utiliser `go test -race` pour confirmer. Fix : `sync.RWMutex` sur `State`
- `agents/alert_agent.go` : `AlertState` modifié sans lock — même risque. Fix : mutex sur `AlertState`

**Scalabilité**

- `Cycles []CycleRecord` dans `state.json` croît sans borne — après des semaines de run continu sur 15s, le fichier peut devenir volumineux
- `logs.jsonl` n'est jamais rotaté — même problème

**Robustesse**

- Bus drop silencieux : pas de métrique sur les cycles perdus
- Fallback LLM vers `action: "log"` masque les pannes d'Ollama sans alerte explicite
- Pas de timeout sur l'exécution des commandes shell whitelistées

**Dashboard**

- Polling toutes les 5s — remplacement futur par WebSocket pour les mises à jour en temps réel
- Pas d'authentification — acceptable sur réseau local uniquement

### Roadmap technique v1.1 ✅ (corrections appliquées)

```
[x] Mutex sur State (race condition)
[x] Mutex sur AlertState et Logger (race condition)
[x] Timeout sur tools.ExecuteCommand() — context.WithTimeout 10s
[x] Cap sur state.Cycles (garder les N derniers)
[x] Validation des valeurs de Config au démarrage — fail fast
[x] Health check Ollama au démarrage — ping + vérification modèle
[x] Structured logging (jxlog) — remplace fmt.Printf
[x] Couleurs ANSI terminal — métriques, décisions, alertes, banner
[x] 58 tests unitaires — parser, health, alertes, registry, shell, config, logger
```

### Roadmap technique v1.5

```
[ ] Structured logging (slog) — remplace fmt.Printf
[ ] Config seuils via variables d'environnement
[ ] Health check Ollama au démarrage (fail fast)
[ ] Rotation automatique des logs (taille max ou date)
[ ] Qdrant client pour mémoire vectorielle longue durée
[ ] Interface Notifier (Discord, Slack, Ntfy, Gotify)
```

### Roadmap technique v2.0

```
[ ] WebSocket pour le dashboard (remplace polling)
[ ] API d'administration REST (enable/disable agents à chaud)
[ ] Plugin system (chargement d'agents sans recompilation)
[ ] Multi-instance coordination
[ ] TLS sur le web server
[ ] Prometheus metrics endpoint (/metrics)
[ ] Config hot-reload (SIGHUP)
```
