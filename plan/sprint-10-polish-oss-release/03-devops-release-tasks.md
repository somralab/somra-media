# Sprint 10 — DevOps & Release Tasks

> **Sprint goal:** Production-quality release: image publishing, versioning, upgrade path,
> and open source distribution (1.0).
>
> **Related:** Sprint 01 [`../sprint-01-foundation/05-devops-tasks.md`](../sprint-01-foundation/05-devops-tasks.md) · [`../roadmap.md`](../roadmap.md) (M6)

## Responsible Role(s)
- DevOps/Platform (primary), Tech Lead

## Dependencies
- All sprints; CI/CD (Sprint 01).

## Epics and Tasks

### Epic A: Release publishing
- [ ] A1 — Automatic versioning (SemVer) + changelog generation | Acceptance: tagged release published.
- [ ] A2 — Multi-arch image publishing (registry) + `latest`/version tags | Acceptance: users can pull.

### Epic B: Upgrade & persistence
- [ ] B1 — Upgrade/migration path validation (old version to new) | Acceptance: upgrade without data loss.
- [ ] B2 — Documented backup/restore flow | Acceptance: user can preserve data.

### Epic C: Deployment ease
- [ ] C1 — One-line install (`docker run`) + production `docker compose` example | Acceptance: <10 min install (brief success criterion).
- [ ] C2 — Version health/diagnostic output | Acceptance: easy diagnostics for support.

## Acceptance Criteria (Sprint Output)
- 1.0 image published; install/upgrade/backup flows validated and documented.

## Risks
- Upgrade/migration errors risk data loss → strict testing.

## Out of Scope
- Cloud one-click deployment (managed) — future.
