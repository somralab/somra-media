# Sprint 13 — Performance at Scale

> **Sprint goal:** Somra stays fast with a **large real-world library** on home hardware
> (4 cores / 8 GB, see [`../project-brief.md`](../project-brief.md) §8): instant browsing,
> quick playback start, seeking without waiting, and several concurrent streams.
>
> **Related:** [`../current-state.md`](../current-state.md) ·
> [`../sprint-10-polish-oss-release/01-performance-tasks.md`](../sprint-10-polish-oss-release/01-performance-tasks.md)
> (this sprint delivers its Epics A–D; Sprint 10 then only re-measures)
>
> **Estimate:** ~2 weeks · **Milestone:** M5.5 (Core experience & performance)

## Responsible Role(s)
- Backend (primary), Media Specialist, Frontend, QA

## Dependencies
- Sprint 11 (correct pipeline + data model) and Sprint 12 (new UI to measure).

## Performance budgets (targets on 4c/8GB, LAN)

| Metric | Target |
|---|---|
| Home API p95 (10k items) | < 100 ms |
| Library page / search API p95 | < 150 ms |
| Web LCP (home, warm) / INP | < 2.5 s / < 200 ms |
| JS initial load (gzip) | < 200 kB |
| Time to first frame: direct play / remux / transcode 1080p | < 1 s / < 3 s / < 5 s |
| Seek to an un-transcoded position | < 5 s |
| Concurrent 1080p SW transcodes without stalls | ≥ 2 (≥ 4 with HW) |

## Epics and Tasks

### Epic A: Measurement first
- [ ] A1 — Synthetic library generator (10k movies, 500 shows / 20k episodes, tiny media
  files) + load script; budgets above reported by a `make` target (non-blocking)
  | Acceptance: one command produces a before/after table in `docs/performance-baseline.md`.
- [ ] A2 — pprof endpoints (admin-only) and a profiling guide | Acceptance: CPU/heap profile
  of a scan and of 4 streams captured.

### Epic B: Streaming
- [ ] B1 — ABR: actually encode the ladder tiers (cap to client/player size; never SW-encode
  4K for a 1080p screen) and serve a real master playlist | Acceptance: hls.js switches
  levels; CPU for a 4K source ≤ 1080p-tier cost.
- [ ] B2 — Seek-ahead: restart ffmpeg at the requested segment when it is outside the
  produced range (segment-index mapping), keep produced segments | Acceptance: seek budget
  met.
- [ ] B3 — Cheap segment requests: no DB write per segment (in-memory session touch, batched
  flush) | Acceptance: zero writes per segment in steady state.
- [ ] B4 — Decision engine uses real bitrate, profile, bit depth and HDR from ffprobe and
  the client capability profile | Acceptance: 4K H.264 within client limits direct-plays.

### Epic C: Database & scan
- [ ] C1 — Keyset (cursor) pagination with an upper limit; parental filter applied in SQL
  before paging | Acceptance: no short pages; p95 budget met at 10k.
- [ ] C2 — Search: sanitize FTS input, prefix search, no N+1 | Acceptance: no 500 on any
  input (fuzz test); p95 met.
- [ ] C3 — Detail/shelf queries: single-item favourite/watchlist lookups; "Recommended"
  without `ORDER BY RANDOM()` | Acceptance: home p95 met.
- [ ] C4 — Scan: bounded ffprobe worker pool and batched transactions | Acceptance: 10k-file
  initial scan ≥ 3× faster than today; UI stays responsive during the scan.

### Epic D: Frontend
- [ ] D1 — Images: `srcset`/`sizes` against the Sprint 12 variants, width/height,
  `decoding=async`, blurhash placeholders, priority hint for the hero | Acceptance: no
  layout shift (CLS < 0.05).
- [ ] D2 — Prefetch detail data on hover/focus; `placeholderData` for instant transitions
  | Acceptance: detail page renders cached data in < 100 ms.
- [ ] D3 — Bundle budget enforced in `make check` (`make bundle-check` or equivalent)
  | Acceptance: `make check` fails on regression above budget.

### Epic E: Stability
- [ ] E1 — Soak: 4 concurrent streams for 2 hours + a background scan on 4c/8GB
  | Acceptance: stable memory, no leaked ffmpeg processes, no stuck sessions.

## Acceptance Criteria (Sprint Output)
- All budgets in the table are met (or the gaps are documented with owners) and recorded in
  `docs/performance-baseline.md`; `current-state.md` is updated.

## Risks
- Seek-ahead is complex (segment alignment) → keep it behind a feature flag until e2e-covered.

## Out of Scope
- Distributed transcoding, CDN, native apps.
