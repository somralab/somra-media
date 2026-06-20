# Sprint 04 — QA Tasks

> **Sprint goal:** Verify playback and transcode correctness, compatibility, and resilience.
>
> **Related:** [`../definition-of-done.md`](../definition-of-done.md) §4 · [`01-media-streaming-tasks.md`](./01-media-streaming-tasks.md)

## Responsible Role(s)
- QA (primary), Media Specialist (test media)

## Dependencies
- This sprint's streaming + player deliverables.

## Epics and Tasks

### Epic A: Compatibility matrix
- [x] A1 — Codec/container test set (H.264/H.265/AAC/AC3/MKV/MP4, etc.) | Acceptance: matrix results are reported.
- [x] A2 — Browser compatibility test (Chrome/Firefox/Safari) | Acceptance: critical formats play.

### Epic B: Functional tests
- [x] B1 — Direct play vs transcode decision accuracy | Acceptance: no unnecessary transcode.
- [x] B2 — Seek/subtitle/audio selection/resume e2e | Acceptance: flows pass.

### Epic C: Resilience & performance
- [x] C1 — Concurrent session/limit stress test | Acceptance: limit is enforced, no crash.
- [x] C2 — Transcode resource/cleanup test | Acceptance: no process/disk leaks.

## Acceptance Criteria (Sprint Output)
- Playback flows are under test coverage; compatibility matrix is documented; no critical bugs.

## Risks
- Wide format variety → automation + sample media pool required.

## Out of Scope
- Hardware acceleration tests — Sprint 07.
