# Sprint 03 — QA Tasks

> **Sprint goal:** Verify identity, RBAC, parental controls, and watch state flows.
>
> **Related:** [`../definition-of-done.md`](../definition-of-done.md) §4 · [`04-security-tasks.md`](./04-security-tasks.md)

## Responsible Role(s)
- QA (primary)

## Dependencies
- This sprint backend + frontend outputs.

## Epics and Tasks

### Epic A: Functional tests
- [x] A1 — Login/logout/refresh e2e flow | Acceptance: critical paths pass.
- [x] A2 — RBAC matrix test (access verification per role) | Acceptance: no permission violations.
- [x] A3 — Parental control test (child profile restrictions) | Acceptance: restricted content not visible.

### Epic B: Watch state
- [x] B1 — Resume and watched state test | Acceptance: correct position preserved.

### Epic C: Security acceptance tests
- [x] C1 — Brute-force/rate-limit and unauthorized access tests | Acceptance: protections trigger.

## Acceptance Criteria (Sprint Output)
- Identity/RBAC/parental flows under test coverage; no critical/security bugs.

## Risks
- Permission edge cases → matrix test must be comprehensive.

## Out of Scope
- Playback performance — Sprint 04.
