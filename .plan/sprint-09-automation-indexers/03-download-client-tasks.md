# Sprint 09 — Download & Automation Tasks (*arr Functionality)

> **Sprint goal:** Download client integration + quality profiles + automatic grab
> and import (Sonarr/Radarr functionality). End-to-end automation with request management.
>
> **Related:** [`02-indexer-integration-tasks.md`](./02-indexer-integration-tasks.md) · Sprint 08 (request handoff) · Sprint 02 (library/import)

## Responsible Role(s)
- Backend (primary)

## Dependencies
- [`01-plugin-architecture-tasks.md`](./01-plugin-architecture-tasks.md), [`02-indexer-integration-tasks.md`](./02-indexer-integration-tasks.md), Sprint 08 request flow.

## Epics and Tasks

### Epic A: Download client adapters
- [x] A1 — Torrent + Usenet download client adapter interface (add, status, completed) | Acceptance: common clients can be added.
- [x] A2 — Download status monitoring (scheduler) | Acceptance: progress/completion tracked.

### Epic B: Quality profiles & decision
- [x] B1 — Quality profile definition (resolution/codec/size preferences) | Acceptance: profile-based selection.
- [x] B2 — Score indexer results by profile + automatic grab | Acceptance: best release selected.

### Epic C: Import & watchlists
- [x] C1 — Import completed download to library (rename/move + trigger scan) | Acceptance: media appears in library.
- [ ] C2 — Watchlist/monitor (automatic episode tracking for series) | Acceptance: new episodes searched automatically.
- [x] C3 — Request → approval → automatic acquisition end-to-end flow (integration with Sprint 08) | Acceptance: approved request completed automatically.

## Acceptance Criteria (Sprint Output)
- An approved request automatically completes the indexer search → quality selection → download → import → library flow.

## Risks
- Complex end-to-end flow + legal sensitivity → plugin isolation and robust error handling.

## Out of Scope
- Full Lidarr/Readarr (music/books) parity — out of scope for this plan (see brief §7).
