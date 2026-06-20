# Sprint 01 — QA Tasks

> **Sprint goal:** Establish test strategy, prepare automation skeleton, and start DoD
> validation process.
>
> **Related:** [`../definition-of-done.md`](../definition-of-done.md) §4 · [`../project-brief.md`](../project-brief.md)

## Responsible Role(s)
- QA (primary), all developers (test writing)

## Dependencies
- CI pipeline ([`05-devops-tasks.md`](./05-devops-tasks.md) Epic B).

## Epics and Tasks

### Epic A: Test strategy
- [x] A1 — Test pyramid and coverage policy documented | Acceptance: unit/integration/e2e boundaries clear.
- [x] A2 — Bug/issue management process and severity levels | Acceptance: critical/high/medium/low defined.
- [x] A3 — Operational definition of coverage standard ([`../definition-of-done.md`](../definition-of-done.md) §4.1): measurement tool, critical module list, report format | Acceptance: core ≥80%, critical ≥90%, frontend components ≥70 thresholds enforceable.

### Epic B: Automation skeleton
- [x] B1 — Backend integration test harness (with isolated DB) | Acceptance: sample test runs in CI.
- [x] B2 — E2E test harness setup (for web flows) | Acceptance: `health` page smoke test passes.

### Epic C: DoD validation
- [x] C1 — Sprint closure checklist (DoD §1–§2) | Acceptance: applied every sprint.
- [x] C2 — i18n acceptance criteria checklist (no hardcoded text, en-US+tr-TR complete) | Acceptance: i18n validated every sprint. See [`../i18n-localization.md`](../i18n-localization.md) §6.

## Acceptance Criteria (Sprint Output)
- Test harnesses run in CI; smoke tests green.
- Test strategy and bug process documented.

## Risks
- Missing early test infrastructure accumulates technical debt → foundation laid this sprint.

## Out of Scope
- Feature-specific comprehensive test scenarios — in relevant sprints.
