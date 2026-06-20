# Sprint 01 — Backend Tasks

> **Sprint goal:** Core service skeleton backend components: application lifecycle,
> job scheduler skeleton, shared utilities.
>
> **Related:** [`../architecture.md`](../architecture.md) · [`../tech-stack.md`](../tech-stack.md) · [`../definition-of-done.md`](../definition-of-done.md) · [`01-architecture-tasks.md`](./01-architecture-tasks.md)

## Responsible Role(s)
- Backend (primary), Tech Lead (oversight)

## Dependencies
- [`01-architecture-tasks.md`](./01-architecture-tasks.md) Epic A/B (module boundaries, bootstrap).

## Epics and Tasks

### Epic A: Application core
- [x] A1 — Application start/stop lifecycle (graceful shutdown, signal handling) | Acceptance: clean shutdown on SIGTERM, tested.
- [x] A2 — Configuration loading + validation (env + defaults) | Acceptance: meaningful error on invalid config.
- [x] A3 — Shared error types and wrapping helpers | Acceptance: standard error response format.

### Epic B: Job Scheduler skeleton
- [x] B1 — Periodic + one-shot job execution infrastructure | Acceptance: sample job runs and is logged.
- [x] B2 — Job status tracking (running/success/error) and concurrency protection | Acceptance: overlapping runs of the same job are prevented.
- [x] B3 — Job queue API skeleton (scanning/refresh will connect in later sprints) | Acceptance: interface contract defined.

### Epic C: Shared infrastructure
- [x] C1 — Structured logger integration | Acceptance: consistent across all modules.
- [x] C2 — Health/diagnostics information collection skeleton | Acceptance: `/api/v1/health` enriched.
- [x] C3 — Backend i18n skeleton: locale-aware message catalog + locale negotiation (user/system/`Accept-Language`/en-US) + API error response with key + localized message | Acceptance: sample error message returns en-US/tr-TR. See [`../i18n-localization.md`](../i18n-localization.md) §4.3.

## Acceptance Criteria (Sprint Output)
- Scheduler safely runs a sample periodic job.
- All backend code complies with DoD §3 standards; tests are green.

## Risks
- Scheduler design will carry all async work in later sprints → interface must be defined early and solidly.

## Out of Scope
- Real business logic (scanning, metadata) — Sprint 02.
