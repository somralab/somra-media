# Somra — Current State (Audit 2026-10-13)

> **Purpose:** what actually works today, verified against code and a local run, not against
> sprint checkboxes. On conflict about **status**, this file wins until the item is re-verified.
> Update it in the same PR whenever a status changes.
>
> **Method:** code reading, `go build` / `go test ./...` / `go test -tags=integration ./...`
> (all pass locally on macOS), and a local end-to-end run (wizard → scan → browse → play) in
> Chromium with generated H.264/AAC MKV samples. Tasks proven not done were reopened in their
> sprint files and marked **"Reopened (audit 2026-10-13)"** (26 tasks in Sprints 02, 04, 05, 07).

Legend: ✅ works · 🟡 partial / shallow · 🔴 broken · ⚪ stub · ⏸ on hold

## Summary

- **Builds and all Go tests pass locally**, but **`main` CI is red** since 2026-06-21
  (`integration-test`, `coverage-gate`), so `build`, `e2e` and `image-build` never ran.
- **No release exists:** no `v*` tag, so no GHCR image.
- **Playback is not reliable in any mode:**
  - Direct play and native HLS (Safari, recent Chromium) get `401`, because media URLs require
    an `Authorization` header that `<video src>` cannot send. Reproduced locally.
  - The ffmpeg context is cancelled as soon as the first manifest appears, so transcodes of
    longer files stop early.
- **Library data degrades on rescan:**
  - A full scan, or a changed file, creates a **new** `media_item`, which orphans its watch
    state and match.
  - TV episodes are not modelled as series → season → episode.
- **The UI is admin-panel-like.** The goal is a streaming-app experience (Sprint 12).
- **Sprint 10 issues were closed** on GitHub on 2026-06-25, but the release never happened.

## Module status

### Platform

| Area | Status | Notes |
|---|---|---|
| Build / tests (local) | ✅ | Unit and integration tests pass on macOS. |
| CI on `main` | 🔴 | `integration-test` + `coverage-gate` fail on Linux; old logs expired. |
| Release / image | 🔴 | No tag, no published image. |
| SQLite layer | 🔴 | PRAGMAs (`busy_timeout`, `foreign_keys`, `synchronous`) are applied to **one** of 8 pooled connections (`internal/platform/db/db.go` `applyPragmas`). Observed: `database is locked (SQLITE_BUSY)` on concurrent refresh. |
| HTTP | 🟡 | No compression; no `Cache-Control` for hashed assets; HTTP/1.1 only. `WriteTimeout` is 0, which is fine for streams. |
| Tests quality | 🟡 | About 23 `*coverage*` / `*_extra_test.go` files exist mainly to hit thresholds. The playback e2e only checks that a `<video>` element is visible. |

### Auth & setup

| Area | Status | Notes |
|---|---|---|
| Setup wizard | ✅ | Works end to end. The default library name "My Library" is not localized. |
| Auth (login, JWT, refresh) | 🟡 | Concurrent refresh → 500; a page reload can sign the user out. |

### Library & metadata

| Area | Status | Notes |
|---|---|---|
| Library scan | 🟡 | Works on small libraries, but serial: one ffprobe per file and about 6 commits per file. |
| Library data correctness | 🔴 | Full or changed-file scan inserts a new `media_item`. `UpsertFile` can use the wrong `LastInsertId` on update. Missing files are never removed. |
| Watcher | 🟡 | Not recursive; not started at boot. |
| Filename parser | 🟡 | Filename only; `1080p` is parsed as a year and `1920x1080` as an episode number. |
| TV model | 🔴 | Each episode becomes its own `media_item`. `season`, `episode`, `genre`, `person` and `media_stream` are never written. |
| Metadata — TMDB | 🟡 | Movies work. TV `Detail` calls the movie endpoint. Images are hotlinked (posters `w500`, backdrops `original`). Auto-match runs nightly only (50 items per library). Key is env-only (`SOMRA_TMDB_API_KEY`). No runtime, logos, trailers or episode stills. |
| Metadata — TVDB / MusicBrainz / fanart | ⚪ | Empty stubs (`internal/metadata/stubs.go`). |
| Image pipeline | ⚪ | No local cache, resize, or image endpoint. |

### Streaming & playback

| Area | Status | Notes |
|---|---|---|
| Decision engine | 🟡 | H.264 is always treated as supported (no profile, 10-bit or HDR check). Bitrate is guessed from resolution. Audio-only incompatibility forces a full video transcode. |
| Transcode lifetime | 🔴 | `runPackaging` cancels its context after the first manifest. ffmpeg runs under `exec.CommandContext`, so it is killed then (`internal/streaming/service.go` `runPackaging`). |
| Startup | 🟡 | The manifest can 404 before the first segment exists. Segments are about 10 s (no forced keyframes). |
| Seek / resume | 🔴 | Linear transcode from `-ss` with no seek-ahead. The resume offset is applied twice (server `-ss` + client `currentTime`). |
| ABR | 🔴 | A ladder is computed, but only the first tier (source resolution) is encoded. |
| Direct play | 🔴 | `<video src=…/source>` gets 401. |
| hls.js path | 🟡 | Plays in Chromium. The token is captured once, so a 401 is expected after the 15-minute access-token lifetime. |
| Native HLS (Safari / iOS) | 🔴 | `video.src = master.m3u8` → 401 (reproduced). |
| HW acceleration | 🔴 | `-hwaccel*` options come after `-i`, so ffmpeg rejects them. Every HW attempt fails, waits 30 s, then falls back to SW. |
| Concurrency | 🔴 | `SetMaxConcurrent` swaps the semaphore while slots are held. HW→SW fallback releases twice. The queue silently drops jobs and can stay stuck in `queued`. |
| Audio / subtitles | 🔴 | Audio streams are never stored, so no track list. No subtitle delivery to the player (no WebVTT, burn-in or embedded extraction). |

### Watch state

| Area | Status | Notes |
|---|---|---|
| Continue watching | 🟡 | Never marked completed. Resume can start at 0. The progress bar assumes 60 minutes. |
| Episodes | 🔴 | Episode "Play" plays the series item; no next episode. |

### Web UI

| Area | Status | Notes |
|---|---|---|
| Design system | 🟡 | Basic colour tokens and a few primitives. No icons, no type or spacing scale. The Noir primary button is white on white. |
| Shell | 🔴 | Boxed layout; 9 text nav items mixing consumer and admin; admin forms (library paths, scans) on consumer pages. |
| Home | 🟡 | Shelves exist, but no hero or backdrops, and rows are not mouse-scrollable. |
| Library grid | 🔴 | Virtualizer sizing mismatch (rows overlap); loads up to 500 items in one request. |
| Player | 🟡 | Boxed page with native controls; no audio/subtitle picker; no next episode. |
| Mobile | 🔴 | No mobile nav; search hidden below 640 px. |
| Delivery | 🟡 | About 306 kB gzip JS (all locales + full hls.js loaded eagerly). |

### On hold

| Area | Status | Notes |
|---|---|---|
| Requests / notifications | ⏸ | Skeleton. On hold. |
| Automation / indexers / download clients | ⏸ | The default build registers **stub** plugins. Real adapters are behind `-tags acquisition` and mostly untested. The import step ignores its computed path. Background jobs run every 1–2 min and add DB write contention. On hold. |

## How to keep this file honest

1. A row moves to ✅ only after a reproducible check (test, e2e, or a documented manual run).
2. When a sprint task is checked off, update the matching row here in the same PR.
