# Sprint 01 — DevOps Tasks

> **Sprint goal:** Single Docker image skeleton, CI/CD pipeline, and development environment standard.
>
> **Related:** [`../tech-stack.md`](../tech-stack.md) §3–§4 · [`../definition-of-done.md`](../definition-of-done.md) §5

## Responsible Role(s)
- DevOps/Platform (primary), Tech Lead (oversight)

## Dependencies
- [`01-architecture-tasks.md`](./01-architecture-tasks.md) and [`02-backend-tasks.md`](./02-backend-tasks.md) (compilable service).

## Epics and Tasks

### Epic A: Docker image
- [x] A1 — Multi-stage Dockerfile: Go binary + frontend static + ffmpeg | Acceptance: reasonable image size, service starts.
- [x] A2 — `docker-compose.yml` example (volumes: config, media, transcode/cache) | Acceptance: runs with a single command.
- [x] A3 — Multi-architecture build (amd64 + arm64) | Acceptance: image produced for both architectures.

### Epic B: CI/CD pipeline
- [x] B1 — CI pipeline: lint → i18n-check → unit-test → integration-test → coverage-gate → build → image-build | Acceptance: DoD §5 gates applied.
- [x] B1b — `i18n-check` step: missing/unused keys + en-US/tr-TR completeness check | Acceptance: missing translation breaks PR. See [`../i18n-localization.md`](../i18n-localization.md) §6.
- [x] B1c — `coverage-gate` step: Go + frontend coverage measurement, report generation, and threshold gate (core ≥80%, critical modules ≥90%, frontend components ≥70%) | Acceptance: merge blocked below threshold; report attached to PR. See [`../definition-of-done.md`](../definition-of-done.md) §4.1.
- [x] B2 — Frontend lint/test/build integration | Acceptance: FE steps green.
- [x] B3 — Version tagging + image publish skeleton (registry decision) | Acceptance: tagged image published.

### Epic C: Development environment
- [x] C1 — Local development flow (hot reload backend + frontend dev server) | Acceptance: documented in README.
- [x] C2 — `Makefile`/task runner (build, test, lint, run) | Acceptance: standard commands work.

## Acceptance Criteria (Sprint Output)
- `docker compose up` starts the service; health endpoint responds.
- CI green through all stages; every PR passes these gates.

## Risks
- ffmpeg packaging + multi-arch build complexity → early validation.

## Out of Scope
- GPU passthrough (hardware acceleration) — Sprint 07.
