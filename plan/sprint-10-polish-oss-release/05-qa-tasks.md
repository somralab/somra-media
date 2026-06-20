# Sprint 10 — QA Tasks (Release Acceptance)

> **Sprint goal:** Comprehensive regression, acceptance, and release quality gate for 1.0 release.
>
> **Related:** [`../definition-of-done.md`](../definition-of-done.md) · [`../roadmap.md`](../roadmap.md) (M6) · [`../project-brief.md`](../project-brief.md) (success criteria)

## Responsible Role(s)
- QA (primary), entire team

## Dependencies
- All previous sprints.

## Epics and Tasks

### Epic A: Full regression
- [ ] A1 — Combined run of all sprint regression packages | Acceptance: green.
- [ ] A2 — End-to-end scenario matrix (install → usage → automation) | Acceptance: critical flows pass.

### Epic B: Release acceptance criteria
- [ ] B1 — Brief success criteria validation (install time, optimization, integrity, performance) | Acceptance: all criteria met.
- [ ] B2 — Multi-platform/browser final check | Acceptance: target environments work.
- [ ] B3 — Upgrade/restore acceptance test | Acceptance: data preserved.
- [ ] B4 — i18n release gate: en-US + tr-TR 100% key completeness, hardcoded text scan, pseudo-locale overflow/length test | Acceptance: both languages complete, no overflow. See [`../i18n-localization.md`](../i18n-localization.md) §6.

### Epic C: Release gate
- [ ] C1 — 1.0 release checklist (code, docs, security, image) | Acceptance: all items complete.

## Acceptance Criteria (Sprint Output)
- M6 (1.0) release quality gate passed; no known critical/high bugs; product ready for release.

## Risks
- Bug pile-up in final phase → continuous regression throughout sprint.

## Out of Scope
- Post-release maintenance/2.0 planning — planned separately.
