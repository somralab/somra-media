# Sprint 11 — Playback & Data Foundation

> **Sprint goal:** A real user can install, log in, scan a movie **and** a TV library, and
> **watch a movie or a whole season to the end in any major browser** without errors. Rescans
> never corrupt data. Fix the blockers from the 2026-10 audit before any redesign work.
>
> **Related:** [`../current-state.md`](../current-state.md) (audit findings) ·
> [`../sprint-12-streaming-experience/01-tasks.md`](../sprint-12-streaming-experience/01-tasks.md) ·
> [`../definition-of-done.md`](../definition-of-done.md)
>
> **Estimate:** ~2 weeks · **Milestone:** M5.5 (Core experience & performance)

## Responsible Role(s)
- Backend (primary), Media Specialist, Frontend, QA

## Dependencies
- Sprints 02, 04, 05 outputs (reopened tasks listed in `current-state.md`). Nothing from
  Sprints 08–09 (on hold).

## Epics and Tasks

### Epic A: Media request authentication (blocker)
- [ ] A1 — Media URLs (playlists, segments, direct-play `source`, subtitles) accept a
  short-lived, session-scoped signed token in the query string (or a path-scoped HttpOnly
  cookie). A plain `<video src>` must work | Acceptance: direct play and native HLS return
  200; the token is bound to session + user, expires with the session, and is redacted in logs.
- [ ] A2 — The player prefers hls.js whenever MSE is available, and uses native HLS only where
  MSE is missing (iOS Safari) | Acceptance: verified in Chrome, Firefox, Safari macOS and iOS
  Safari.
- [ ] A3 — hls.js reads the current access token per request (no token captured once)
  | Acceptance: 60-minute playback with 15-minute access tokens has no 401.
- [ ] A4 — Playback e2e asserts real progress (`currentTime > 3 s`, no `video.error`, and a
  seek to 80 % works) for direct play, remux and transcode fixtures longer than 2 minutes
  | Acceptance: the e2e fails on today's `main` and passes after the fixes.

### Epic B: Transcode pipeline correctness (blocker)
- [ ] B1 — ffmpeg lifetime is owned by the playback session, not by `runPackaging`'s
  10-minute context that is cancelled after the first manifest | Acceptance: a 2-hour file
  transcodes to the end; ffmpeg stops on session stop or idle reap.
- [ ] B2 — The manifest endpoint waits (bounded) for the first segment or returns a retryable
  status. `WaitForManifest` fails fast when ffmpeg has exited | Acceptance: no fatal hls.js
  error on start; clear error when the source is unreadable.
- [ ] B3 — Force keyframes to the segment length (`-force_key_frames` / GOP) so 4-second
  segments are real | Acceptance: segment durations within ±0.5 s of target.
- [ ] B4 — Resume applies the offset once (server `-ss` with matching timeline, or client
  seek, not both) | Acceptance: resume lands within ±2 s.
- [ ] B5 — Audio-only incompatibility (AC3/DTS/5.1) uses video copy + audio transcode
  | Acceptance: H.264 + AC3 MKV plays with no video re-encode.
- [ ] B6 — Fix HW ffmpeg argument order (`-hwaccel*` before `-i`) and the GPU scale filters
  (`scale_cuda` / `scale_qsv`). HW failure falls back to SW immediately, not after 30 s
  | Acceptance: VAAPI or QSV verified on at least one real machine; fallback in < 2 s.
- [ ] B7 — ProcessManager: concurrency-safe resizing of the limit, one release per acquire,
  the queue returns `ErrQueueFull` and drains on failure too | Acceptance: race detector
  clean (`go test -race`); 10 parallel play requests never leave sessions stuck in `queued`.

### Epic C: Database & session stability
- [x] C1 — Apply PRAGMAs on **every** pooled connection (DSN `_pragma=` parameters)
  | Acceptance: a test opens N connections and asserts every PRAGMA value.
- [ ] C2 — Refresh-token rotation tolerates concurrent refresh (grace window, or single-flight
  on the client) | Acceptance: 20 parallel refreshes → no 500 and no forced logout.
- [ ] C3 — The session survives a full page reload | Acceptance: e2e reloads `/` and stays
  signed in.

### Epic D: Library data correctness
- [ ] D1 — A rescan or a changed file updates the existing `media_item` (stable file → item
  identity) instead of inserting a new one. Fix `UpsertFile` id resolution (no
  `LastInsertId` after an upsert) | Acceptance: 3 full rescans keep item ids, watch state and
  matches unchanged.
- [ ] D2 — Remove or mark missing files. The watcher is recursive and starts at boot
  | Acceptance: add, delete and move in a nested folder are reflected within 1 minute.
- [ ] D3 — TV model: series → season → episode rows (`season` / `episode` tables populated).
  TMDB TV uses `/tv/{id}` and `/tv/{id}/season/{n}` | Acceptance: a show library lists one
  series with correct seasons and episodes, with titles and overviews.
- [ ] D4 — Store audio and subtitle streams from ffprobe (`media_stream`) and runtime;
  expose runtime in the API | Acceptance: the API returns tracks and duration for every file.
- [ ] D5 — Parser: use folder names; stop treating `1080p` / `1920x1080` as year or episode
  | Acceptance: `testdata/library/naming` cases pass, with new regression cases added.
- [ ] D6 — TMDB key configurable in Settings (not env-only); auto-match runs after each scan
  | Acceptance: new items get metadata within minutes of a scan.

### Epic E: Watch progress you can trust
- [ ] E1 — Wait for watch state before starting a session and honour the start position
  | Acceptance: "Resume" starts within ±2 s of the saved position.
- [ ] E2 — Mark completed at ≥ 90 %, flush progress on exit or tab close (`sendBeacon`), and
  call stop on unmount | Acceptance: finished items leave "Continue watching"; no orphan
  sessions.
- [ ] E3 — Episode Play starts the episode. A next-episode endpoint exists | Acceptance:
  S01E02 is returned after S01E01.

### Epic F: CI & honesty
- [ ] F1 — Make `main` green: re-run CI, fix `integration-test` and `coverage-gate` on Linux
  | Acceptance: all gates green on `main`.
- [ ] F2 — Replace coverage-padding tests in packages touched this sprint with behaviour tests
  (gradual refactor rule in `AGENTS.md`) | Acceptance: no `*coverage*` test files left in
  `internal/streaming`, `internal/library` and `internal/auth`.
- [ ] F3 — Background jobs of on-hold modules (automation, requests) do not run when nothing
  is configured | Acceptance: an idle server writes nothing to the DB.

### Epic G: Quick performance wins
- [ ] G1 — gzip/brotli for API JSON and SPA assets; `Cache-Control: immutable` for hashed
  assets | Acceptance: JS over the wire ≤ 50 % of today's size.
- [ ] G2 — Lazy-load hls.js and locale bundles (only the active locale) | Acceptance: main
  chunk < 120 kB gzip.

## Acceptance Criteria (Sprint Output)
- Fresh install → wizard → scan a movie library and a TV library → play a 2-hour movie and
  binge two episodes in Chrome, Firefox and Safari. Resume and "Continue watching" behave
  correctly; rescans keep data intact; CI is green. Every fixed row in
  [`../current-state.md`](../current-state.md) is updated.

## Risks
- Signed media URLs touch auth: needs security review (token scope, expiry, log redaction).
- The TV model is the largest item. If it slips, D3 moves to the start of Sprint 12, but before
  the Sprint 12 detail page.

## Out of Scope
- Visual redesign (Sprint 12). ABR ladder and seek-ahead (Sprint 13). Requests / automation /
  indexers (on hold).
