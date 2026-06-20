# Sprint 04 — Backend Tasks (Playback Integration)

> **Sprint goal:** Integrate the streaming pipeline with user/watch state and the API.
>
> **Related:** [`01-media-streaming-tasks.md`](./01-media-streaming-tasks.md) · Sprint 03 watch state

## Responsible Role(s)
- Backend (primary), Media Specialist (pipeline interface)

## Dependencies
- [`01-media-streaming-tasks.md`](./01-media-streaming-tasks.md), Sprint 03 (`watch_state`, authorization).

## Epics and Tasks

### Epic A: Playback API
- [x] A1 — "Play" endpoint: authorization + playback decision + session start | Acceptance: end-to-end stream starts.
- [x] A2 — Watch progress update (periodic ping) | Acceptance: resume data is updated.

### Epic B: Session & resource management
- [x] B1 — Concurrent transcode session limit + queuing | Acceptance: overload is prevented.
- [x] B2 — Idle session termination | Acceptance: no resource leaks.

### Epic C: Telemetry
- [x] C1 — Playback/transcode metrics (session count, error rate) | Acceptance: basic metrics are collected.

## Acceptance Criteria (Sprint Output)
- Playback API works with authorization + watch state + session management.

## Risks
- Concurrent session management is critical on home hardware → limits/queue are required.

## Out of Scope
- Hardware acceleration selection — Sprint 07.
