# ADR-001 — Store mémoire longue durée : SQLite

**Date :** 2026-05-31  
**Status :** Accepted  
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

Phase 1 : double write (v1.x)  
Phase 2 : bascule lecture (v1.x+1)  
Phase 3 : nettoyage + /api/query (v1.x+2)
