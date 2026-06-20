# Sprint 03 — Security Tasks

> **Sprint goal:** Establish security foundations for the identity and authorization layer.
>
> **Related:** [`../definition-of-done.md`](../definition-of-done.md) §6 · [`01-backend-tasks.md`](./01-backend-tasks.md) · Sprint 10 security audit

## Responsible Role(s)
- Backend (primary), Tech Lead (oversight)

## Dependencies
- [`01-backend-tasks.md`](./01-backend-tasks.md), [`02-database-tasks.md`](./02-database-tasks.md)

## Epics and Tasks

### Epic A: Identity security
- [x] A1 — Strong password hash + policy | Acceptance: weak password rejected.
- [x] A2 — Brute-force protection (rate limit, lockout) | Acceptance: repeated failed logins limited.
- [x] A3 — Secure session/token (expiry, revocation, refresh) | Acceptance: stolen token can be revoked.

### Epic B: Authorization security
- [x] B1 — Mandatory authorization check on all protected endpoints | Acceptance: unauthorized access blocked, tested.
- [x] B2 — Input validation and injection protection | Acceptance: parameterized queries, validation.

### Epic C: Secret management
- [x] C1 — Secure storage of provider keys and secrets | Acceptance: no secrets embedded in code.

## Acceptance Criteria (Sprint Output)
- Identity/authorization layer runs with secure defaults; resilient to basic attack scenarios.

## Risks
- Missing authorization check is critical vulnerability → comprehensive testing and Sprint 10 audit.

## Out of Scope
- Full security audit/penetration testing — Sprint 10.
