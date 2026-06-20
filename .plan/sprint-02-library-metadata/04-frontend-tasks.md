# Sprint 02 — Frontend Tasks (Library Management)

> **Sprint goal:** Library definition, scan triggering, and management UI showing scan/metadata status
> (no playback yet).
>
> **Related:** Sprint 01 [`../sprint-01-foundation/04-frontend-tasks.md`](../sprint-01-foundation/04-frontend-tasks.md) · [`01-backend-tasks.md`](./01-backend-tasks.md)

## Responsible Role(s)
- Frontend (primary)

## Dependencies
- Sprint 01 SPA skeleton + API client; this sprint backend library/scan APIs.

## Epics and Tasks

### Epic A: Library management
- [x] A1 — Create/edit/delete library screens | Acceptance: connected to CRUD API, with validation.
- [x] A2 — Folder/path selection UI | Acceptance: multiple paths can be added.

### Epic B: Scan monitoring
- [x] B1 — Scan trigger + real-time progress (WS/SSE) | Acceptance: progress updates live.
- [x] B2 — Scan history / job status view | Acceptance: successful/failed jobs listed.

### Epic C: Metadata preview
- [x] C1 — Basic list of matched items + metadata preview | Acceptance: poster + title + year displayed.
- [x] C2 — Manual re-match UI | Acceptance: uses backend correction API.

## Acceptance Criteria (Sprint Output)
- User defines library, scans, and views results from the UI.

## Risks
- Real-time progress UX → WS/SSE integration must be solid.

## Out of Scope
- Rich browsing/player screens — Sprint 05.
