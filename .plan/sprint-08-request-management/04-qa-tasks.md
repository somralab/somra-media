# Sprint 08 — QA Tasks

> **Sprint goal:** Validate request flow, approval, and notifications.
>
> **Related:** [`../definition-of-done.md`](../definition-of-done.md) · [`01-backend-tasks.md`](./01-backend-tasks.md)

## Responsible Role(s)
- QA (primary)

## Dependencies
- This sprint's backend/frontend/notification outputs.

## Epics and Tasks

### Epic A: Request flow
- [x] A1 — Create request → approve/reject → status e2e | Acceptance: all transitions correct.
- [x] A2 — Conflict/quota/permission tests | Acceptance: policies enforced.

### Epic B: Notifications
- [x] B1 — Trigger test for each channel | Acceptance: delivered on correct event.
- [x] B2 — Preference/subscription test | Acceptance: unwanted notifications not sent.

### Epic C: Regression
- [x] C1 — Sprint 08 regression package | Acceptance: runs in CI.

## Acceptance Criteria (Sprint Output)
- Request and notification flows covered by tests; no critical bugs.

## Risks
- External channel dependency → mock + limited real testing.

## Out of Scope
- Automation/indexer tests — Sprint 09.
