# ADR-001 — Store mémoire longue durée : SQLite

**Date :** 2026-05-31  
**Status :** Implemented — V1.7 (juin 2026)  
**Décideurs :** JARVINx project

## Contexte

`memory/State` est limité à 20 snapshots et 20 cycles en mémoire.
Insuffisant pour la mémoire sémantique (Qdrant), les rapports quotidiens précis,
et les requêtes temporelles.

## Options considérées

| Option           | Pour                         | Contre                         |
| ---------------- | ---------------------------- | ------------------------------ |
| BBolt            | Pure Go, simple              | Abandonné, pas de SQL          |
| SQLite (modernc) | SQL complet, Pure Go, mature | Légèrement plus lourd          |
| PostgreSQL       | SQL avancé                   | Trop lourd, dépendance externe |

## Décision

**SQLite via `modernc.org/sqlite`** — Pure Go, pas de CGO, SQL complet.

## Conséquences

- Dépendance externe ajoutée : `modernc.org/sqlite`
- Migration en 3 phases (double write → bascule → nettoyage)
- Interface `Store` abstraite — agents non impactés
- Prérequis pour v1.x Qdrant

## Schema

Voir `memory/schema.sql` (créé en phase 1).

## Migration

Phase 1 : double write (v1.7) ✅  
Phase 2 : bascule lecture (v1.7) ✅  
Phase 3 : nettoyage + /api/query — à venir

## Implémentation V1.7

- `memory/store.go` — interfaces `Store`, `EventLog`, `HistoryReader` + struct `SnapshotBucket`
- `memory/sqlite_store.go` — `*SQLiteStore` : WAL, `synchronous=NORMAL`, index timestamp, `SnapshotBuckets()`, `TotalSnapshots()`
- `memory/double_write_store.go` — `*DoubleWriteStore` : écritures fan-out, lectures depuis SQLite
- `memory/noop_store.go` — `NoopStore` : fallback fail-silencieux
- Config : `JARVINX_SQLITE_PATH` (env var, vide = désactivé)
- API : `GET /api/history/full?range=7d|30d|90d`
- Dashboard : graphes Recharts AreaChart + sélecteur période (page History)
