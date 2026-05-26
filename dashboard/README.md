# JARVINx Dashboard

> Interface web du runtime agentique JARVINx — Next.js 14, Tailwind v4, TypeScript.

---

## Stack

| Outil | Usage |
|-------|-------|
| Next.js 14 (App Router) | Framework React |
| Tailwind CSS v4 | Styling |
| TypeScript | Typage |
| Recharts | Graphiques (prévu v1.5) |
| Lucide React | Icônes |

---

## Prérequis

- Node.js 18+
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
│   ├── layout.tsx          # Layout principal — sidebar + topbar + alert banner
│   ├── page.tsx            # Overview
│   ├── agents/
│   │   └── page.tsx        # Registry des agents
│   ├── history/
│   │   └── page.tsx        # Historique des cycles
│   └── settings/
│       └── page.tsx        # Configuration runtime + workspace viewer
│
├── components/
│   ├── sidebar.tsx         # Navigation latérale
│   ├── topbar.tsx          # Barre supérieure — status runtime, heure
│   ├── alert-banner.tsx    # Bannière d'alerte — apparaît si action=alert
│   ├── runtime-cycle.tsx   # Visualisation OBSERVE→THINK→DECIDE→ACT→LEARN
│   ├── agent-list.tsx      # Liste compacte des agents avec health
│   ├── decision-feed.tsx   # Feed des dernières décisions LLM
│   ├── metrics-bar.tsx     # Barre de métrique CPU/RAM/Disk
│   └── ui/
│       └── stat-card.tsx   # Card de statistique générique
│
└── lib/
    ├── api.ts              # Client typé vers le runtime Go
    ├── hooks.ts            # useStatus, useHistory, useAgents — polling
    └── utils.ts            # cn(), formatTime(), metricColor()...
```

---

## API consommée

Le dashboard consomme l'API REST du runtime Go.

| Endpoint | Intervalle de polling | Description |
|----------|----------------------|-------------|
| `GET /api/status` | 5s | Dernier cycle + métriques live |
| `GET /api/history` | 15s | 10 derniers cycles |
| `GET /api/agents` | 10s | État du registry d'agents |

Les types TypeScript correspondants sont dans `lib/api.ts`.

---

## Pages

### Overview `/`
Vue d'ensemble du runtime — 4 stat cards (health, agents actifs, décisions, intervalle), visualisation du cycle agent, liste des agents, feed de décisions, métriques live CPU/RAM/Disk.

### Agents `/agents`
Registry complet — une card par agent avec health %, runs, erreurs, schedule, dernière exécution. Affiche les erreurs actives si présentes.

### History `/history`
Tableau de tous les cycles enregistrés avec métriques colorées (vert/amber/rouge selon seuils), action badge, analyse LLM et commande exécutée si applicable. Stats globales en haut (total, log/suggest/alert/execute).

### Settings `/settings`
Configuration runtime live (modèle, intervalle, cycle, uptime), seuils d'alerte, notifications Discord, endpoints API cliquables, et viewer workspace.yml avec syntax highlighting et badge VALID.

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

### v1.5 — Intelligence & Mémoire
- [ ] Page Memory — vector DB, historique sémantique
- [ ] Page Tools — whitelist commandes, logs d'exécution
- [ ] Graphiques sparkline CPU/RAM sur l'historique
- [ ] WebSocket — remplace le polling pour les mises à jour temps réel
- [ ] Page Workflows — configuration agents à chaud

### v2.0 — Universal Platform
- [ ] API d'administration — enable/disable agents depuis le dashboard
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