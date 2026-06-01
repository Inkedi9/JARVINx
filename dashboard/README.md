# JARVINx Dashboard

> Interface web du runtime agentique JARVINx — Next.js 16, React 19, Tailwind v4, TypeScript.

---

## Stack

| Outil | Usage |
|-------|-------|
| Next.js 16 (App Router) | Framework React |
| React 19 | UI |
| Tailwind CSS v4 | Styling (CSS custom properties) |
| TypeScript | Typage |
| Lucide React | Icônes |
| Jest | Tests unitaires |

---

## Prérequis

- Node.js 20+
- Le runtime Go JARVINx qui tourne sur `localhost:8080`

---

## Installation

```bash
cd dashboard
npm install
```

---

## Configuration

Crée un fichier `.env.local` à la racine de `dashboard/` :

```env
NEXT_PUBLIC_RUNTIME_URL=http://localhost:8080
```

### Environnements disponibles

```
.env.local        ← dev Windows  (localhost)
.env.homelab      ← réseau local (192.168.1.X:8080)
.env.tailscale    ← accès externe via Tailscale (100.X.X.X:8080)
```

Pour switcher d'environnement :

```bash
# Copie l'env voulu puis relance
cp .env.homelab .env.local
npm run dev
```

> Les variables `NEXT_PUBLIC_*` ne sont pas rechargées à chaud — toujours redémarrer le dev server après un changement de `.env.local`.

---

## Lancement

```bash
# Dev
npm run dev

# Build production
npm run build
npm start
```

Dashboard accessible sur `http://localhost:3000`.

---

## Structure

```
dashboard/
│
├── app/
│   ├── layout.tsx              # Layout principal — sidebar + topbar + alert banner
│   ├── page.tsx                # Overview
│   ├── agents/page.tsx         # Registry des agents
│   ├── containers/page.tsx     # Containers Docker live — filtres All/Running/Exited
│   ├── history/page.tsx        # Historique des cycles
│   ├── llm-context/page.tsx    # Tendances + contexte adaptatif LLM
│   └── settings/page.tsx       # Configuration runtime + endpoints API
│
├── components/
│   ├── sidebar.tsx             # Navigation latérale
│   ├── topbar.tsx              # Barre supérieure — status, badge Docker running/total
│   ├── alert-banner.tsx        # Bannière si action=alert
│   ├── runtime-cycle.tsx       # Visualisation OBSERVE→THINK→DECIDE→ACT
│   ├── agent-list.tsx          # Liste agents avec health
│   ├── decision-feed.tsx       # Feed décisions LLM
│   ├── metrics-bar.tsx         # Jauges CPU/RAM/Disk
│   ├── ai-analysis.tsx         # Bloc analyse IA (depuis /api/llm-context)
│   ├── daily-reporter.tsx      # Widget DailyReporter — last_sent + trigger
│   └── ui/stat-card.tsx        # Card statistique générique
│
└── lib/
    ├── api.ts                  # Types TypeScript miroir des structs Go
    ├── hooks.ts                # useStatus (5s), useHistory (15s), useAgents (10s)
    └── utils.ts                # cn(), formatTime(), metricColor()...
```

---

## API consommée

Le dashboard consomme l'API REST du runtime Go. Les types TypeScript correspondants sont dans `lib/api.ts`.

| Endpoint | Polling | Utilisé par |
|----------|---------|-------------|
| `GET /api/status` | 5s | Overview — métriques, cycle, circuit state |
| `GET /api/history` | 15s | History — tableau des cycles |
| `GET /api/agents` | 10s | Agents — registry et health |
| `GET /api/docker` | 5s | Containers — tableau live + badge topbar |
| `GET /api/llm-context` | 15s | LLM Context + bloc Analyse IA |
| `GET /api/daily-report` | 30s | Widget DailyReporter |
| `POST /api/daily-report/send` | on-demand | Trigger rapport immédiat |
| `GET /api/file` | 30s | Settings — status FileAgent |
| `GET /api/logs/status` | 30s | Settings — état logs |

---

## Pages

### Overview `/`
Vue d'ensemble — 4 stat cards, cycle agent visuel, liste agents, feed décisions LLM, bloc Analyse IA, métriques CPU/RAM/Disk live.

### Agents `/agents`
Registry complet — une card par agent avec health %, runs, erreurs, schedule, dernière exécution, bouton enable/disable à chaud.

### Containers `/containers`
Tableau Docker live avec filtres All / Running / Exited. Badge dans la topbar indique running/total en rouge si des containers sont down.

### History `/history`
Tableau des cycles avec métriques colorées (vert/amber/rouge selon seuils), action badge, analyse LLM et commande exécutée. Stats globales en haut.
Pour `execute` et `suggest` : badge confiance coloré (vert ≥ 75% / orange ≥ 50% / rouge) et sous-texte "Déclenché à CPU X% / RAM Y%" si des champs trigger sont présents.

### LLM Context `/llm-context`
Visualise le contexte adaptatif transmis au LLM : tendances CPU/RAM/Disk, action dominante, taux d'alerte, dernières alertes déclenchées.

### Settings `/settings`
Configuration runtime live (modèle, intervalle, uptime), seuils d'alerte, widget DailyReporter (dernier envoi + trigger manuel), endpoints API cliquables.

---

## Composants clés

### `AlertBanner`
Apparaît automatiquement sous la topbar quand le dernier cycle a `action: "alert"`. Se reset à chaque nouveau cycle alert, dismissable manuellement.

### `RuntimeCycle`
Visualise les 5 étapes du cycle agent avec l'étape active en surbrillance bleue et les étapes complètes en vert.

### `usePolling` (dans `hooks.ts`)
Hook générique qui fetch une endpoint toutes les N millisecondes. Gère les erreurs sans faire crasher le composant.

---

## Thème

Dark theme fixe basé sur les couleurs JARVINx.

| Variable CSS | Valeur | Usage |
|-------------|--------|-------|
| `--color-bg-primary` | `#0D1117` | Fond principal |
| `--color-bg-secondary` | `#161B22` | Cards, sidebar |
| `--color-bg-tertiary` | `#1C2128` | Inputs, hover |
| `--color-border` | `#21262D` | Bordures |
| `--color-accent-blue` | `#3B82F6` | Accent principal |

Fonts : **Inter** pour le texte, **JetBrains Mono** pour les métriques, labels et code.

---

## Roadmap

### v1.5 — Dashboard ✅
- [x] Page Containers — tableau Docker live, filtres, badge topbar
- [x] Page LLM Context — tendances + contexte adaptatif
- [x] Widget DailyReporter — last_sent + trigger manuel
- [x] Bloc Analyse IA — résumé LLM depuis `/api/llm-context`
- [x] Badge Docker topbar — running/total, rouge si containers down

### v1.6 — Couche décisionnelle ✅
- [x] History — badge confiance coloré (execute/suggest) + sous-texte trigger
- [x] Overview — statut execute guard temps réel (cooldown actif / disponible)
- [x] Types TypeScript `CycleRecord` + champs `confidence`, `trigger_cpu/ram/disk`
- [x] Types TypeScript `StatusResponse` + champ `exec_guard`

### v1.7 — Dashboard améliorations
- [ ] Graphiques sparkline CPU/RAM sur l'historique
- [ ] WebSocket — remplace le polling pour les mises à jour temps réel

### v2.0 — Universal Platform
- [ ] Multi-instance — surveiller plusieurs runtimes depuis une interface
- [ ] Auth — protection du dashboard en prod
- [ ] TLS — HTTPS natif
- [ ] Mobile — layout responsive

---

## Lien avec le runtime

Le dashboard est découplé du runtime Go — il consomme uniquement l'API REST. Pour changer l'URL du runtime, modifie `.env.local`. Le runtime peut tourner sur n'importe quelle machine accessible (même via Tailscale).

```
Windows dev        → .env.local     → localhost:8080
Réseau local       → .env.homelab   → 192.168.1.X:8080
Tailscale externe  → .env.tailscale → 100.X.X.X:8080
```