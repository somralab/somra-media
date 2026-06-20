# Sprint 01 — Database Tasks

> **Sprint goal:** SQLite data layer foundation, migration infrastructure, and schema versioning.
>
> **Related:** [`../tech-stack.md`](../tech-stack.md) · [`../architecture.md`](../architecture.md) §4 · [`../definition-of-done.md`](../definition-of-done.md)

## Responsible Role(s)
- Backend (primary), Tech Lead (schema review)

## Dependencies
- [`01-architecture-tasks.md`](./01-architecture-tasks.md) Epic A3 (migration tool decision).

## Epics and Tasks

### Epic A: Data access layer
- [x] A1 — SQLite connection management (WAL mode, connection pool, pragma settings) | Acceptance: WAL active, concurrent read test passes.
- [x] A2 — Repository/access pattern skeleton | Acceptance: CRUD tested for sample table.
- [x] A3 — Transaction helpers | Acceptance: rollback/commit tests.

### Epic B: Migration and schema versioning
- [x] B1 — Migration infrastructure (up/down) setup | Acceptance: `up`/`down` work, version table maintained.
- [x] B2 — Schema versioning and automatic migration on application startup | Acceptance: schema auto-updates on upgrade.
- [x] B3 — Seed and test data infrastructure | Acceptance: isolated test DB setup.

### Epic C: Backup/resilience foundation
- [x] C1 — DB file location and volume strategy (persistence) | Acceptance: volume compatible with [`05-devops-tasks.md`](./05-devops-tasks.md).
- [x] C2 — Integrity check and corruption recovery notes | Acceptance: basic `PRAGMA integrity_check` flow.

## Acceptance Criteria (Sprint Output)
- Application applies migrations on startup; sample repository tests pass.
- Data layer is compatible with [`../architecture.md`](../architecture.md) §4.

## Risks
- Schema decisions are early; frequent migrations in later sprints → migration discipline required.

## Out of Scope
- Domain tables (media, users, etc.) — added in relevant sprints.
