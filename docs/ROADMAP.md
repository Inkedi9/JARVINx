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

**Platform**

- **App desktop** — wrapper Wails autour du binaire Go + dashboard Next.js
- **Plugin system dynamique** — agents chargés sans recompiler
- **Multi-instance coordination** — surveillance multi-machines
- **API d'administration** REST complète
- **TLS + auth dashboard**
- **Audit trail**

**Le fil conducteur :**
Chaque agent suit le même pattern.
Chaque agent parle le même langage (interface Agent).
Tout tourne local. Tout reste privé.

---

## 📌 Décisions d'architecture prises

| Date       | Décision                                                      | Contexte                                                                                                                  |
| ---------- | ------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------- |
| 2026-05    | Interface Agent générique + BaseAgent embedding               | Composition Go plutôt qu'héritage                                                                                         |
| 2026-05    | `embed.FS` pour le dashboard                                  | Binaire auto-suffisant, zéro déploiement statique                                                                         |
| 2026-05    | Une seule dépendance externe (gopsutil)                       | Résistance dans le temps                                                                                                  |
| 2026-05-26 | V1.2 intercalée avant V1.5                                    | Corrections audit avant nouvelles features                                                                                |
| 2026-05-26 | Shell dispatch direct sans `sh -c`                            | Défense en profondeur même avec whitelist                                                                                 |
| 2026-06-01 | Store mémoire longue durée SQLite vs BBolt                    | Définir le store mémoire longue durée                                                                                     |
| 2026-06-01 | Interface `Store` + `EventLog` abstraites                     | Découpler les agents des implémentations concrètes avant SQLite                                                           |
| 2026-06-01 | `DoubleWriteStore` — JSON source de vérité + SQLite secondary | Migration sans coupure, rollback possible à chaque étape                                                                  |
| 2026-06-01 | `CycleNum` comme ID commun SQLite ↔ Qdrant                    | Éviter un UUID supplémentaire, déjà stable et incrémental                                                                 |
| 2026-06-02 | `QdrantAgent` dans le Registry plutôt qu'un subscriber Bus    | Isolation panic gratuite, conditionnel à `JARVINX_QDRANT_URL`                                                             |
| 2026-06-02 | `collReady bool` plutôt que `sync.Once` pour l'init Qdrant    | `sync.Once` fige l'erreur si Qdrant est down au démarrage                                                                 |
| 2026-06-02 | Deux circuit breakers indépendants (LLM + embedding)          | Embedding et chat ne se bloquent pas mutuellement                                                                         |
| 2026-06-02 | Injection `SimilarDecisions` en N-1 via `AgentContext`        | Cohérent avec le pattern N-1 existant, zéro couplage direct                                                               |
| 2026-06-04 | `v1.x` jusqu'au changement structurel cloud/multi-instance    | La v2.0 est réservée à un vrai saut d'architecture                                                                        |
| 2026-06-05 | App desktop Wails envisagée pour v2.0+                        | Wrapper natif autour du binaire Go + dashboard Next.js existant — embed.FS et architecture single-binary déjà compatibles |
