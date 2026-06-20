# Sprint 08 — Backend Tasks (Request Management — Overseerr Functionality)

> **Sprint goal:** Users create requests for missing content, approval workflow, and
> request status tracking (Overseerr/Jellyseerr functionality).
>
> **Related:** [`../architecture.md`](../architecture.md) (Request Management) · Sprint 02 (metadata search) · Sprint 03 (user/RBAC)

## Responsible Role(s)
- Backend (primary)

## Dependencies
- Sprint 02 (metadata provider search), Sprint 03 (user/permissions).

## Epics and Tasks

### Epic A: Request model
- [x] A1 — `request` schema (movie/series, requester, status, resolution/quality preference) | Acceptance: migration + CRUD.
- [x] A2 — Conflict check with existing library (already exists) | Acceptance: duplicate request blocked/flagged.

### Epic B: Approval workflow
- [x] B1 — Request state machine (pending → approved/rejected → completed) | Acceptance: state transitions controlled.
- [x] B2 — Role-based auto-approval/quota (admin policy) | Acceptance: user quota enforced.

### Epic C: Discovery & search
- [x] C1 — Search for "addable content" via provider (not in library) | Acceptance: search returns results.
- [x] C2 — Automation bridge skeleton (handoff point to connect with Sprint 09) | Acceptance: approved request can be handed off to automation via interface.

## Acceptance Criteria (Sprint Output)
- User can request content; admin approves; request status tracked; notification triggered.

## Risks
- Integration with Sprint 09 automation → handoff interface must be clearly defined.

## Out of Scope
- Actual download/grab — Sprint 09.
