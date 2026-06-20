# Sprint 01 — Architecture Tasks

> **Sprint goal:** A running skeleton service, clear module boundaries, API contract, and
> CI/CD foundation. By the end of this sprint, `docker run` yields an empty but running Somra core (M1).
>
> **Related:** [`../project-brief.md`](../project-brief.md) · [`../architecture.md`](../architecture.md) · [`../tech-stack.md`](../tech-stack.md) · [`../definition-of-done.md`](../definition-of-done.md) · [`../i18n-localization.md`](../i18n-localization.md)

## Responsible Role(s)
- Tech Lead / Architect (primary), Backend (support)

## Dependencies
- None (initial sprint). Outputs are the foundation for **all** subsequent sprints.

## Epics and Tasks

### Epic A: Implementation and validation of closed decisions
> All technology/architecture decisions are made (see [`../tech-stack.md`](../tech-stack.md) §7, [`../architecture.md`](../architecture.md) §8). This epic **implements/validates** decisions; it does not reopen debate.
- [x] A1 — HTTP router (`go-chi/chi`) integration + skeleton | Acceptance: routing + middleware work.
- [x] A2 — Session/identity foundation (JWT access + revocable refresh token) skeleton | Acceptance: contract ready for Sprint 03.
- [x] A3 — Migration (`pressly/goose`) + scheduler (`robfig/cron/v3`) integration | Acceptance: sample migration + cron job run.
- [x] A4 — OpenAPI 3.1 design-first contract skeleton + FE type generation pipeline | Acceptance: `/api/v1/health` types generated from spec.
- [x] A5 — i18n architecture implementation: libraries (`i18next`/`react-i18next`, `go-i18n/v2`), key standard (`domain.context.key`), locale negotiation | Acceptance: working skeleton. See [`../i18n-localization.md`](../i18n-localization.md).

### Epic B: Module skeleton
- [x] B1 — Monorepo directory structure and module boundaries (API, identity, library, metadata, streaming, settings, jobs) | Acceptance: aligns exactly with [`../architecture.md`](../architecture.md) §3.
- [x] B2 — Dependency injection / application bootstrap skeleton | Acceptance: service starts/stops cleanly (graceful shutdown).
- [x] B3 — Configuration layer (environment variables + defaults) | Acceptance: "convention over configuration" principle applied.
- [x] B4 — Structured logging + error handling standard | Acceptance: all modules use shared logger.

### Epic C: API Gateway foundation
- [x] C1 — HTTP server + `/api/v1/health` and `/api/v1/version` endpoints | Acceptance: returns 200, has tests.
- [x] C2 — Middleware chain (request log, recover, CORS, rate-limit skeleton) | Acceptance: verified with unit tests.
- [x] C3 — WebSocket/SSE infrastructure skeleton (for real-time events) | Acceptance: sample event broadcast works.

## Acceptance Criteria (Sprint Output)
- Service compiles as a single binary and starts; health/version endpoints respond.
- Architecture decisions are documented; open decision list is closed.
- [`../definition-of-done.md`](../definition-of-done.md) §1–§2 are met.

## Risks
- Early wrong architecture decisions are costly → decisions are documented and reviewed.

## Out of Scope
- Business logic (scanning, playback, etc.) — later sprints. Skeleton only.
