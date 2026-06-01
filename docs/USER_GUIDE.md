# JARVINx — Manuel Utilisateur

> Version 1.0 | Pour les questions : ouvre une issue sur [GitHub](https://github.com/Inkeki9/JARVINx)

---

## Sommaire

1. [Qu'est-ce que JARVINx ?](#quest-ce-que-jarvinx-)
2. [Prérequis](#prérequis)
3. [Installation](#installation)
4. [Configuration](#configuration)
5. [Lancement](#lancement)
6. [CLI interactive](#cli-interactive)
7. [Dashboard web](#dashboard-web)
8. [Alertes Discord](#alertes-discord)
9. [Fichiers générés](#fichiers-générés)
10. [Dépannage](#dépannage)

---

## Qu'est-ce que JARVINx ?

JARVINx est un **runtime agentique local** : il observe ton système en continu, envoie les métriques à un LLM local (Ollama), reçoit une décision structurée, et agit — sans cloud, sans abonnement.

```
observe → think → decide → act → log → repeat (toutes les 15s)
```

Ce que JARVINx fait concrètement :

- Lit CPU, RAM et espace disque toutes les 15 secondes
- Envoie ces données à un LLM local (llama3.1, qwen2.5, mistral...)
- Reçoit une décision : `log` / `alert` / `suggest` / `execute`
- Exécute les commandes autorisées si la décision est `execute`
- Alerte sur Discord si un seuil est dépassé
- Affiche tout sur un dashboard web à `http://localhost:8080`

Ce que JARVINx **ne fait pas** :

- Pas de connexion cloud
- Pas de commandes shell arbitraires (whitelist stricte)
- Pas de modification de fichiers système

---

## Prérequis

| Outil  | Version minimale | Vérification       |
| ------ | ---------------- | ------------------ |
| Go     | 1.21+            | `go version`       |
| Ollama | latest           | `ollama --version` |
| Git    | any              | `git --version`    |

### Modèles Ollama recommandés

| Modèle             | RAM requise | Recommandation                     |
| ------------------ | ----------- | ---------------------------------- |
| `llama3.1:8b`      | ~6 GB       | Meilleur équilibre qualité/vitesse |
| `qwen2.5:7b`       | ~5 GB       | Très rapide, excellent en JSON     |
| `qwen2.5-coder:7b` | ~5 GB       | Si tu ajoutes des agents code      |
| `mistral:7b`       | ~5 GB       | Alternative légère                 |

> **Note RAM :** Ollama charge le modèle entier en mémoire. Sur 8 GB de RAM, préfère `qwen2.5:7b` pour laisser de la place au système.

---

## Installation

### Étape 1 — Cloner le projet

```bash
git clone https://github.com/Inkedi9/JARVINx.git
cd JARVINx
```

### Étape 2 — Installer les dépendances Go

```bash
go mod tidy
```

Tu dois voir `go: downloading github.com/shirou/gopsutil/v3 ...` — c'est la seule dépendance externe.

### Étape 3 — Préparer Ollama

```bash
# Puller le modèle (une seule fois)
ollama pull llama3.1:8b

# Vérifier que le modèle est présent
ollama list
```

Ollama doit tourner en arrière-plan. Sur Windows il démarre automatiquement après installation. Sur Linux/macOS :

```bash
ollama serve
```

### Étape 4 — Configurer l'environnement (optionnel)

Crée un fichier `.env` à la racine du projet si tu veux les alertes Discord :

```env
# Discord webhook (optionnel)
DISCORD_WEBHOOK=https://discord.com/api/webhooks/TON_ID/TON_TOKEN

# Logs debug (optionnel — false par défaut)
JARVINX_DEBUG=false
```

Sans ce fichier, JARVINx fonctionne normalement — les alertes Discord sont simplement désactivées.

---

## Configuration

Tout est dans [config/config.go](../config/config.go). Modifie directement les valeurs par défaut :

```go
func Default() *Config {
    return &Config{
        // Fréquence d'observation — minimum 5s recommandé
        Interval: 15 * time.Second,

        // LLM — URL Ollama et modèle
        OllamaURL: "http://localhost:11434",
        Model:     "llama3.1:8b",

        // Fichiers de données
        LogFile:   "logs.jsonl",
        StateFile: "state.json",
        AlertFile: "alerts.jsonl",

        // Dashboard
        WebPort: 8080,

        // Seuils d'alerte (en %)
        CPUAlertThreshold:  85.0,
        RAMAlertThreshold:  90.0,
        DiskAlertThreshold: 85.0,

        // Anti-spam alertes
        AlertCooldown:  5,   // Minimum N cycles entre deux alertes identiques
        AlertMinCycles: 2,   // Cycles consécutifs au-dessus du seuil avant d'alerter
    }
}
```

### Option A — Variables d'environnement (recommandé)

Ajoute dans ton `.env` ce dont tu as besoin :

```env
# Modèle LLM
JARVINX_MODEL=qwen2.5:7b

# Fréquence d'observation
JARVINX_INTERVAL=30s

# Seuils d'alerte (%)
JARVINX_CPU_THRESHOLD=80
JARVINX_RAM_THRESHOLD=85
JARVINX_DISK_THRESHOLD=85

# Rotation des logs
JARVINX_LOG_MAX_MB=10
JARVINX_LOG_MAX_BACKUPS=3
```

Aucune recompilation nécessaire. Les valeurs invalides sont ignorées avec un warning.

### Option B — Modifier config.go

Pour des changements permanents ou des valeurs non exposées via env, modifie directement `runtime/config/config.go` et recompile.

### Paramètres importants expliqués

**`Interval`** — Fréquence des cycles d'observation.

- `15s` : bon équilibre réactivité / charge LLM
- `30s` : si ton LLM est lent ou ta machine limitée
- `5s` : minimum autorisé (limite imposée par la CLI)

**`Model`** — Le modèle Ollama à utiliser. Doit correspondre exactement au nom affiché par `ollama list`.

**`AlertMinCycles`** — Nombre de cycles **consécutifs** au-dessus du seuil avant d'envoyer une alerte CPU ou RAM. Avec `2` cycles à 15s d'intervalle, une alerte CPU nécessite 30s de dépassement continu. Évite les faux positifs lors du chargement du LLM lui-même.

**`AlertCooldown`** — Nombre minimum de cycles entre deux alertes sur la même métrique. Avec `5` cycles à 15s, même si le CPU reste haut, tu ne recevras une alerte que toutes les 75 secondes maximum.

### Changer le modèle

```go
Model: "qwen2.5:7b",  // ou "mistral:7b", "llama3.2:3b", etc.
```

Assure-toi d'avoir pullé le modèle : `ollama pull qwen2.5:7b`

### Changer le port du dashboard

```go
WebPort: 9090,  // Dashboard sur http://localhost:9090
```

### Surveiller un autre disque

Le chemin disque est détecté automatiquement (`/` sur Linux/macOS, `C:\` sur Windows). Pour surveiller un autre volume, modifie [tools/system.go](../tools/system.go) — cherche `diskPath`.

---

## Lancement

### Windows

**Option recommandée — PowerShell avec .env**

Utilise le script inclus :

```powershell
.\run.ps1
```

Ce script charge automatiquement les variables du `.env` puis lance JARVINx.

**Alternative — variables manuelles**

```powershell
$env:DISCORD_WEBHOOK="https://discord.com/api/webhooks/..."
go run cmd/main.go
```

**Binaire compilé (production)**

```powershell
# Compiler
go build -o jarvinx.exe cmd/main.go

# Lancer
$env:DISCORD_WEBHOOK="..."
.\jarvinx.exe
```

**Ou via Makefile (si Make est installé)**

```powershell
make run          # Lancer en dev
make build        # Compiler pour Windows
make build-linux  # Cross-compiler pour Linux
```

### Linux / macOS

```bash
# Avec .env
export $(cat .env | xargs)
go run cmd/main.go

# Ou via Makefile
make run
```

**Service systemd (Linux production)**

```bash
# Compiler le binaire
go build -o jarvinx cmd/main.go

# Créer le service
sudo nano /etc/systemd/system/jarvinx.service
```

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

### Mode dry-run

Lance JARVINx sans qu'il exécute de commandes ou envoie d'alertes Discord :

```bash
# Via flag
go run cmd/main.go --dry-run

# Via env
JARVINX_DRY_RUN=true .\run.ps1
```

Tu verras une bannière jaune au démarrage. Utile pour :

- Tester une nouvelle config de seuils
- Valider un déploiement sur un nouveau serveur
- Débugger sans risque

### Sortie attendue au démarrage

```
[ JARVINX ] Vérification Ollama...
[ OLLAMA ] ✓ En ligne · modèle 'llama3.1:8b' disponible · 5 modèle(s) installé(s)
[ OK   ] Discord webhook chargé
[ REGISTRY ] Agent enregistré : system (schedule: 15s)
[ REGISTRY ] Agent enregistré : alert (schedule: 15s)
╔══════════════════════════════════════════════╗
║           JARVINx — RUNTIME v1.2            ║
╚══════════════════════════════════════════════╝
  Modèle     : llama3.1:8b
  Intervalle : 15s
  Seuils     : CPU 85% · RAM 90% · Disk 85%

[ WEB ] Dashboard → http://localhost:8080
[ ORCHESTRATOR ] En écoute sur le bus...
[ REGISTRY ] Démarrage agent : system
[ REGISTRY ] Démarrage agent : alert
[ CLI ] Prêt — tape 'help' pour les commandes
```

Si tu vois `connection refused` à l'adresse Ollama, vérifie qu'Ollama est bien lancé.

---

## CLI interactive

La CLI fonctionne en parallèle du runtime — tape directement dans le terminal pendant que JARVINx tourne.

### Commandes disponibles

> Les commandes sont exécutées directement sans shell intermédiaire (`exec("df", "-h")` et non `sh -c "df -h"`). L'injection shell est structurellement impossible.

#### `help`

Affiche la liste des commandes disponibles.

```
  Commandes disponibles :
  ─────────────────────────────────────────
  status              État du dernier cycle
  history [n]         Derniers N snapshots (défaut: 5)
  interval <secondes> Change l'intervalle de tick
  clear               Efface l'écran
  help                Cette aide
  ─────────────────────────────────────────
```

#### `status`

Affiche le dernier snapshot système observé.

```
  ┌─[ DERNIER SNAPSHOT ]─────────────────────┐
  │ Timestamp : 2025-05-20 14:32:17
  │ CPU       : 12.4%
  │ RAM       : 4823 MB / 16384 MB (29.4%)
  │ DISK      : 187 GB / 512 GB (36.5%)
  └──────────────────────────────────────────┘
```

#### `history [n]`

Affiche les N derniers snapshots sous forme de tableau. Par défaut : 5.

```
history 10       # Afficher les 10 derniers snapshots
```

```
  Historique — 10 derniers snapshots :
  ──────────────────────────────────────────────────────
  Heure       CPU       RAM                 Disk
  ──────────────────────────────────────────────────────
  14:32:17    12.4%     4823/16384 MB       187/512 GB
  14:32:02    11.8%     4820/16384 MB       187/512 GB
  ...
```

#### `interval <secondes>`

Change l'intervalle d'observation **à chaud**, sans redémarrer JARVINx.

```
interval 30    # Observer toutes les 30 secondes
interval 5     # Observer toutes les 5 secondes (minimum)
```

> **Attention :** Un intervalle très court (5-10s) augmente la charge sur Ollama. Si le LLM ne répond pas assez vite, des cycles seront ignorés.

#### `clear`

Efface le terminal. Utile quand les logs s'accumulent.

---

## Agents disponibles

### SystemAgent

Actif par défaut. Envoie les métriques à Ollama toutes les 15s et prend une décision.
Non désactivable — c'est le cerveau de JARVINx.

### AlertAgent

Surveille les seuils CPU/RAM/Disk. Envoie des alertes via les notifiers configurés.
Configure les seuils via env vars : `JARVINX_CPU_THRESHOLD`, `JARVINX_RAM_THRESHOLD`, `JARVINX_DISK_THRESHOLD`.

### DockerAgent

Surveille tes containers Docker. Actif si Docker est accessible.

```env
# Surveille tous les containers
JARVINX_DOCKER_ENABLED=true

# Surveille seulement certains containers
JARVINX_DOCKER_WATCH=nginx,postgres,redis
```

### FileAgent

Surveille des dossiers et détecte les fichiers volumineux.

```env
JARVINX_FILE_WATCH=/var/log,/home/user/downloads
JARVINX_FILE_MAX_MB=500
```

Désactivé si `JARVINX_FILE_WATCH` est vide.

---

## Dashboard web

Le dashboard est accessible à `http://localhost:8080` dès le démarrage de JARVINx.

### Sections du dashboard

**Runtime Info** (bandeau supérieur)

- Modèle LLM actif
- Intervalle de cycle
- Numéro du cycle courant
- Uptime depuis le démarrage

**Métriques système** (3 jauges)

- CPU % avec barre de progression
- RAM utilisée / totale en MB avec %
- Disk utilisé / total en GB avec %
- Actualisé toutes les 5 secondes

**Dernière décision LLM**

- Action décidée par le LLM (`log` / `alert` / `suggest` / `execute`)
- Analyse courte de l'état du système
- Raison de la décision

**Agent Loop** (animation visuelle)

- Visualise le cycle : Observe → Think → Decide → Act → Sleep
- L'étape active est mise en surbrillance

**Console logs** (style macOS)

- Derniers logs du runtime en temps réel
- Code couleur selon le type d'événement

**Historique des 10 derniers cycles** (tableau)

- Timestamp, CPU, RAM, Disk, action décidée
- Badges colorés selon l'action (log=gris, alert=rouge, suggest=orange, execute=bleu)

### API REST

Le dashboard consomme une API interne que tu peux aussi interroger directement.

#### `GET /api/status`

Retourne l'état actuel du runtime.

```json
{
  "online": true,
  "model": "llama3.1:8b",
  "interval": "15s",
  "cycle_num": 42,
  "uptime": "2h 15m 30s",
  "last_cycle": {
    "cycle_num": 42,
    "timestamp": "2025-05-20T14:32:17Z",
    "cpu": 12.4,
    "mem_percent": 29.4,
    "disk_percent": 36.5,
    "action": "log",
    "analysis": "Système stable, pas de tendance préoccupante",
    "reason": "Toutes les métriques en dessous des seuils"
  }
}
```

#### `GET /api/history`

Retourne les 10 derniers cycles, du plus récent au plus ancien.

```json
{
  "cycles": [ ... ],
  "total": 42
}
```

---

## Notifications

JARVINx supporte plusieurs canaux simultanément. Configure ceux dont tu as besoin dans `.env` :

```env
# Discord
DISCORD_WEBHOOK=https://discord.com/api/webhooks/...

# Slack
SLACK_WEBHOOK=https://hooks.slack.com/services/...

# Ntfy (self-hosted ou ntfy.sh)
NTFY_URL=https://ntfy.sh
NTFY_TOPIC=mon-jarvinx

# Gotify (self-hosted)
GOTIFY_URL=https://gotify.example.com
GOTIFY_TOKEN=mon-token
```

Tous les canaux configurés reçoivent les alertes simultanément.
Un échec sur un canal n'affecte pas les autres.

### Rapport quotidien

```env
JARVINX_DAILY_REPORT=true
JARVINX_REPORT_HOUR=8
JARVINX_REPORT_MINUTE=0
```

Envoie un résumé quotidien via tous les notifiers configurés à l'heure définie.

---

## Alertes Discord

### Configuration du webhook

1. Dans ton serveur Discord, va dans les paramètres d'un salon texte
2. `Intégrations` → `Webhooks` → `Nouveau webhook`
3. Copie l'URL du webhook
4. Colle-la dans ton `.env` :

```env
# Discord webhook (optionnel)
DISCORD_WEBHOOK=https://discord.com/api/webhooks/TON_ID/TON_TOKEN

# Logs debug (optionnel)
JARVINX_DEBUG=false

# Origins CORS supplémentaires — pour accès homelab ou Tailscale (optionnel)
# JARVINX_ALLOWED_ORIGINS=http://192.168.1.X:3000,http://100.X.X.X:3000
```

### Ce qui déclenche une alerte

| Métrique | Condition                                               | Niveau   |
| -------- | ------------------------------------------------------- | -------- |
| CPU      | ≥ 85% pendant **2 cycles consécutifs** (30s par défaut) | CRITICAL |
| RAM      | ≥ 90% pendant **2 cycles consécutifs** (30s par défaut) | CRITICAL |
| Disk     | ≥ 85% pendant un cycle (peut se répéter selon cooldown) | WARNING  |

### Format des alertes Discord

JARVINx envoie des embeds structurés :

```
🚨 critical — CPU
CPU à 91.3% depuis 3 cycles consécutifs

Valeur : 91.3%   Seuil : 85.0%   Cycles : 3

JARVINx · Autonomous Agent Runtime — 14:32:17
```

### Anti-spam

Le système évite les alertes en rafale :

- **CPU / RAM** : alerte seulement après N cycles consécutifs au-dessus du seuil (défaut : 2). Un pic isolé ne déclenche pas d'alerte.
- **Cooldown** : minimum 5 cycles (75s) entre deux alertes sur la même métrique, même si le seuil reste dépassé.
- **Reset** : dès que la métrique repasse sous le seuil, le compteur de cycles repart à zéro.

---

## Fichiers générés

JARVINx génère trois fichiers dans le répertoire de lancement :

### `state.json`

Historique des cycles + snapshots. Persiste entre les redémarrages — JARVINx reprend le comptage de cycles là où il en était.

```json
{
  "cycle_num": 42,
  "history": [
    {
      "timestamp": "2025-05-20T14:32:17Z",
      "cpu_percent": 12.4,
      "mem_used": 4823,
      "mem_total": 16384,
      "mem_percent": 29.4,
      "disk_used": 187,
      "disk_total": 512,
      "disk_percent": 36.5
    }
  ],
  "cycles": [
    {
      "cycle_num": 42,
      "timestamp": "2025-05-20T14:32:17Z",
      "action": "log",
      "analysis": "Système stable",
      "reason": "Métriques dans les normes",
      "command": ""
    }
  ]
}
```

### `logs.jsonl`

Une ligne JSON par cycle observé. Format JSONL (JSON Lines) — chaque ligne est un JSON valide indépendant.

```jsonl
{"timestamp":"2025-05-20T14:32:17Z","cpu_percent":12.4,"mem_used":4823,"mem_total":16384,"mem_percent":29.4,"disk_used":187,"disk_total":512,"disk_percent":36.5}
{"timestamp":"2025-05-20T14:32:32Z","cpu_percent":15.1,"mem_used":4850,...}
```

**Lire les logs :**

```bash
# Dernière ligne (dernier cycle)
tail -1 logs.jsonl | jq .

# Tous les cycles où le CPU dépasse 80%
cat logs.jsonl | jq 'select(.cpu_percent > 80)'
```

### `alerts.jsonl`

Historique de toutes les alertes déclenchées.

```jsonl
{
  "timestamp": "2025-05-20T14:32:17Z",
  "level": "critical",
  "metric": "CPU",
  "value": 91.3,
  "threshold": 85,
  "message": "CPU à 91.3% depuis 3 cycles consécutifs",
  "cycles_above": 3
}
```

---

## Dépannage

### JARVINx démarre mais ne produit pas de décision LLM

**Cause probable :** Ollama n'est pas joignable.

```bash
# Vérifier qu'Ollama tourne
curl http://localhost:11434/api/tags

# Si ça ne répond pas, lancer Ollama
ollama serve
```

### JARVINx refuse de démarrer avec "configuration invalide"

La validation de config a détecté un problème. Exemples d'erreurs :

- `CPUAlertThreshold doit être <= 100` — seuil > 100%
- `Interval trop court` — interval < 5s
- `WebPort invalide` — port < 1024 ou > 65535

Corrige les valeurs dans `config/config.go` et relance.

### Ollama est lancé mais le modèle n'est pas trouvé

Le modèle configuré n'est pas installé. Lance :

```bash
ollama pull llama3.1:8b
```

Ou change le modèle dans `config/config.go` pour un modèle présent dans `ollama list`.

### Le LLM répond très lentement ou time out

**Causes et solutions :**

1. Le modèle est trop lourd pour ta RAM → passe à `qwen2.5:7b` ou `llama3.2:3b`
2. Le premier appel est lent (chargement en RAM) → normal, les suivants seront plus rapides
3. Si le timeout persiste → augmente `Interval` à 30s pour donner plus de temps au LLM

### Les alertes Discord ne s'envoient pas

1. Vérifie que `DISCORD_WEBHOOK` est chargé : au démarrage tu dois voir `[ OK ] Discord webhook chargé`
2. Vérifie l'URL du webhook — elle doit commencer par `https://discord.com/api/webhooks/`
3. Teste le webhook manuellement :

```bash
curl -H "Content-Type: application/json" \
  -d '{"content": "test JARVINx"}' \
  https://discord.com/api/webhooks/TON_ID/TON_TOKEN
```

### Le dashboard ne s'affiche pas

1. Vérifie que le port 8080 n'est pas déjà utilisé :

```bash
# Windows
netstat -ano | findstr :8080

# Linux
lsof -i :8080
```

2. Change le port dans `config/config.go` : `WebPort: 9090`

### "Cycle précédent en cours — tick ignoré"

Ce message est normal. Il indique que le LLM a mis plus de 15s pour répondre, donc un cycle a été sauté. Solutions :

- Augmente `Interval` : `interval 30` dans la CLI
- Utilise un modèle plus rapide

### Erreur `connection refused` au démarrage

Ollama n'est pas lancé. Lance `ollama serve` dans un terminal séparé avant de démarrer JARVINx.

### Les métriques disque semblent fausses sur Windows

Le chemin disque par défaut est `C:\`. Si ton disque principal est une autre lettre, modifie [tools/system.go](../tools/system.go) — cherche `diskPath`.

### Les seuils d'alerte ne semblent pas pris en compte

Vérifie l'ordre de priorité — une variable d'environnement système écrase le `.env` :

```bash
# Vérifie ce qui est actif
echo $env:JARVINX_CPU_THRESHOLD  # PowerShell
echo $JARVINX_CPU_THRESHOLD       # Linux/macOS
```

La config active est affichée au démarrage :
[ JARVINX ] Modèle : llama3.1:8b | Intervalle : 15s | CPU : 85% RAM : 90% Disk : 85%

### logs.jsonl grossit trop vite

Configure la rotation dans `.env` :

```env
JARVINX_LOG_MAX_MB=5
JARVINX_LOG_MAX_BACKUPS=2
```

Les anciens logs sont archivés en `logs.jsonl.1`, `logs.jsonl.2`.

### Le circuit breaker est ouvert — LLM bloqué

Si `/api/status` retourne `"circuit_state": "open"`, Ollama a eu trop d'échecs consécutifs.

JARVINx va automatiquement tester à nouveau après 30 secondes (`half-open`).
Pour forcer le reset : redémarre JARVINx.

Vérifie qu'Ollama tourne :

```bash
ollama serve
curl http://localhost:11434/api/tags
```

### logs.jsonl est plein

Configure la rotation dans `.env` :

```env
JARVINX_LOG_MAX_MB=5
JARVINX_LOG_MAX_BACKUPS=3
```

Consulte l'état des logs via l'API :

```bash
curl http://localhost:8080/api/logs/status
```
