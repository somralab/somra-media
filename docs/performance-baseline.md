# Performance baseline (Sprint 10)

Reference hardware: **4 CPU cores / 8 GB RAM** (home server target from
[`.plan/project-brief.md`](../.plan/project-brief.md)).

Measured on: **2026-06-21** · host: darwin/arm64 dev machine (12 cores).
Re-run on 4c/8GB reference hardware before v1.0.0 sign-off.

## How to reproduce

```bash
make profile                    # Go CPU/mem profiles → profiles/
make build-web && make bundle-check
./scripts/soak-test.sh SOAK_DURATION=15m
go test -tags=soak ./internal/platform/bootstrap/ -count=1 -timeout=3m
SOMRA_LARGE_LIB_SIZE=10000 go test ./internal/platform/db/ -run LargeLibrary -count=1
pnpm --dir web run build:analyze   # emits web/dist/stats.html
```

## Backend (Epic A)

| Metric | Target (4c/8GB) | Measured (dev) | Notes |
|--------|-----------------|----------------|-------|
| Browse page (50 items, 2k library) | < 50 ms p95 | ~0.24 ms/op | `BenchmarkBrowseRepo_ListPaginated` |
| FTS search (2k library) | < 80 ms p95 | ~0.09 ms/op | `BenchmarkBrowseRepo_SearchFTS` |
| Scan progress DB writes | batched every 25 files | 25 (default) | `SOMRA_SCAN_PROGRESS_BATCH` |
| Job queue workers | 2 | 2 | `SOMRA_JOB_WORKERS` override |
| Idle RSS (empty library) | < 180 MiB | ~95–120 MiB (soak script) | `/api/v1/health` probes |
| RSS under scan + browse | < 512 MiB | TBD on 4c/8GB | soak WARN threshold |

### DB indexes (migration `20260901000001_performance_indexes.sql`)

- `idx_media_item_library_created`, `_sort`, `_year`
- `idx_watch_state_item`, `idx_watch_state_user_updated`
- `idx_media_genre_genre`, `idx_artwork_item_kind`

Verified via `TestBrowseRepo_LargeLibraryQueryPlans` (EXPLAIN uses indexes, no seq scan on `library_id` filter).

## Streaming (Epic B)

| Metric | Target (4c/8GB) | Default |
|--------|-----------------|---------|
| Max concurrent SW transcodes | 2 | `streaming.max_concurrent_transcodes` |
| Max concurrent HW transcodes | 1 | smart defaults when GPU detected |
| Transcode queue depth | 8 | `SOMRA_STREAMING_MAX_QUEUE` |
| Session slot release on cancel | no leak | `TestStopSessionReleasesTranscodeSlot` |

HW session cap on 4 cores: **1** (was 2). Total concurrent sessions remain **2**.

## Frontend (Epic C)

| Metric | Target | Notes |
|--------|--------|-------|
| Gzipped JS+CSS (production build) | ≤ 850 KiB (warn 750) | **352 KiB** (2026-06-21) | `make bundle-check` |
| Code splitting | admin/automation lazy | `React.lazy` in `web/src/App.tsx` |
| Long library grids | virtualized | `@tanstack/react-virtual` |

Run `ANALYZE=1 pnpm --dir web run build` for `dist/stats.html` treemap.

## Stability (Epic D)

| Check | Target | Tool |
|-------|--------|------|
| Soak (4–8 h pre-release) | RSS stable, health OK | `./scripts/soak-test.sh SOAK_DURATION=4h` |
| Short soak (manual/CI) | 90 s probe loop | `go test -tags=soak ./internal/platform/bootstrap/` |

## Tuning environment variables

| Variable | Default (4c) | Purpose |
|----------|--------------|---------|
| `SOMRA_JOB_WORKERS` | 2 | Background job worker pool |
| `SOMRA_JOB_BUFFER` | 16 | Job queue buffer |
| `SOMRA_SCAN_PROGRESS_BATCH` | 25 | Scan progress flush interval |
| `SOMRA_STREAMING_MAX_CONCURRENT` | 2 | Transcode slots |
| `SOMRA_LARGE_LIB_SIZE` | 2000 (tests) | Large-library fixture size |
| `BUNDLE_MAX_KB` | 850 | Frontend gzip budget |

## Deferred

- Full **10k** library benchmark in default CI (`SOMRA_LARGE_LIB_SIZE=10000` locally).
- **4–8 h** soak on reference 4c/8GB hardware (script ready; not run in CI).
- Transcode **start latency** SW/HW benchmarks (needs fixture media + ffmpeg).
- Admin-gated runtime `pprof` HTTP endpoint.
- Bundle budget wired into `.github/workflows/ci.yml` (DevOps stream).
