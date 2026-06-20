# Sprint 08 — Frontend Tasks (Request UI)

> **Sprint goal:** Content discovery/request creation, request tracking, and admin approval interfaces.
>
> **Related:** [`01-backend-tasks.md`](./01-backend-tasks.md) · Sprint 05 (browsing UI pattern)

## Responsible Role(s)
- Frontend (primary)

## Dependencies
- [`01-backend-tasks.md`](./01-backend-tasks.md) request APIs.

## Epics and Tasks

### Epic A: Request creation
- [x] A1 — Discovery/search (content not in library) + "request" flow | Acceptance: user creates request.
- [x] A2 — Quality/resolution preference selection | Acceptance: sent to backend.

### Epic B: Request tracking
- [x] B1 — "My Requests" screen + status display | Acceptance: real-time status.

### Epic C: Admin approval
- [x] C1 — Pending requests + approve/reject interface | Acceptance: status updated.
- [x] C2 — Quota/policy settings interface | Acceptance: tied to backend policy.

## Acceptance Criteria (Sprint Output)
- User creates and tracks requests; admin approves/rejects.

## Risks
- Status synchronization → live updates via WS/SSE.

## Out of Scope
- Automation settings screens — Sprint 09.
