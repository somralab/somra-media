# Sprint 04 — Media & Streaming Tasks

> **Sprint goal:** Direct play + software (CPU) transcode, HLS/DASH packaging, adaptive
> bitrate (ABR), and subtitle/audio track handling. Core of M3 (first end-to-end playback).
>
> **Related:** [`../architecture.md`](../architecture.md) §3 (Streaming) · [`../project-brief.md`](../project-brief.md) (transcode decision: software first) · [`../tech-stack.md`](../tech-stack.md)

## Responsible Role(s)
- Media/Streaming Specialist (primary), Backend (support)

## Dependencies
- Sprint 02 (media/technical metadata), Sprint 03 (authorization — who can play what).

## Epics and Tasks

### Epic A: Playback decision
- [x] A1 — Client capability profile intake (supported codec/container) | Acceptance: matching against client profile.
- [x] A2 — Direct play / direct stream / transcode decision engine | Acceptance: unnecessary transcode is avoided.

### Epic B: Transcode pipeline (software)
- [x] B1 — ffmpeg process management (start/monitor/terminate, resource limits) | Acceptance: no leftover (zombie) processes.
- [x] B2 — CMAF (fMP4) segment generation + HLS manifest (primary); DASH manifest optional from the same segments | Acceptance: output playable in browser with hls.js. See [`../tech-stack.md`](../tech-stack.md) §2.
- [x] B3 — Adaptive bitrate (multiple quality tiers) | Acceptance: tiers are produced, switching works.
- [x] B4 — Transcode session management + segment cache/cleanup | Acceptance: disk does not bloat; cleaned up when session closes.

### Epic C: Audio & subtitles
- [x] C1 — Audio track/language selection + downmix when needed | Acceptance: multiple audio streams can be selected.
- [x] C2 — Subtitle handling (embedded extraction, external file, burn-in when needed) | Acceptance: subtitles are displayed.

### Epic D: Streaming endpoints
- [x] D1 — Streaming API endpoints (manifest, segment, session) + authorization control | Acceptance: unauthorized streams are blocked.
- [x] D2 — Seek and resume support | Acceptance: forward/rewind works.

## Acceptance Criteria (Sprint Output)
- A media file plays end-to-end in the browser (via direct play or CPU transcode).
- Audio/subtitle selection and seek work; transcode sessions are managed cleanly.

## Risks
- **Highest technical risk.** Wide codec/container variety → compatibility: compatibility matrix and tests are critical.
- ffmpeg process/resource management errors affect the system → strict limits required.

## Out of Scope
- Hardware acceleration — Sprint 07 (this sprint is CPU only).
