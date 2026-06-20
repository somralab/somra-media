# Sprint 04 — Frontend Tasks (Web Player)

> **Sprint goal:** hls.js-based web video player: playback, seek, quality/audio/subtitle selection,
> resume.
>
> **Related:** [`01-media-streaming-tasks.md`](./01-media-streaming-tasks.md) · [`02-backend-tasks.md`](./02-backend-tasks.md) · [`../tech-stack.md`](../tech-stack.md)

## Responsible Role(s)
- Frontend (primary), Media Specialist (compatibility support)

## Dependencies
- Backend streaming/playback APIs.

## Epics and Tasks

### Epic A: Player core
- [x] A1 — hls.js integration + manifest loading | Acceptance: video plays.
- [x] A2 — Play/pause/seek/volume controls + keyboard shortcuts | Acceptance: basic controls work.
- [x] A3 — Fullscreen + responsive behavior | Acceptance: works well on desktop/mobile browsers.

### Epic B: Stream options
- [x] B1 — Quality (ABR) selection (automatic + manual) | Acceptance: tier switching is smooth.
- [x] B2 — Audio language and subtitle selection UI | Acceptance: connected to backend streams.

### Epic C: Resume
- [x] C1 — Resume (continue from last position) + periodic progress reporting | Acceptance: correct position.

## Acceptance Criteria (Sprint Output)
- User watches video in the browser; selects quality/audio/subtitles; resumes from last position.

## Risks
- Browser/codec compatibility → test matrix (Chrome/Firefox/Safari).

## Out of Scope
- Rich library browsing screens — Sprint 05.
