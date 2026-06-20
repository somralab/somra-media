# Sprint 02 — Backend Tasks (Library & Scanning)

> **Sprint goal:** Media library definition, file scanning, file watching, and
> technical metadata extraction with ffprobe. Foundation for M2.
>
> **Related:** [`../architecture.md`](../architecture.md) §3 · [`../project-brief.md`](../project-brief.md) · [`../definition-of-done.md`](../definition-of-done.md) · Sprint 01

## Responsible Role(s)
- Backend (primary), Tech Lead (oversight)

## Dependencies
- Sprint 01: job scheduler, data layer, API gateway.

## Epics and Tasks

### Epic A: Library definition
- [x] A1 — Library concept: type (movie/series/music), source folder(s), scan settings | Acceptance: CRUD API + test.
- [x] A2 — Multi-folder/volume support | Acceptance: multiple paths can be scanned.

### Epic B: File scanning engine
- [x] B1 — Full scan job: file discovery, supported format filter | Acceptance: stable on large folders, reports progress.
- [x] B2 — Incremental scan (changed items only) | Acceptance: change detection works.
- [x] B3 — Technical metadata via ffprobe (codec, resolution, duration, audio/subtitle tracks) | Acceptance: correct parse, verified with test data.
- [x] B4 — Pre-parsing from filename/folder structure (title, year, season/episode) | Acceptance: common naming patterns resolved.

### Epic C: File watching
- [x] C1 — Filesystem watcher (add/delete/move) | Acceptance: incremental scan triggered on change.
- [x] C2 — Debounce/batch processing | Acceptance: system does not choke on bulk changes.

## Acceptance Criteria (Sprint Output)
- A library is defined and scanned; technical metadata written to DB; watching active.
- All jobs run via scheduler with progress reporting.

## Risks
- Various naming/format combinations → comprehensive test data required.

## Out of Scope
- Enriched (external provider) metadata — see [`03-metadata-providers-tasks.md`](./03-metadata-providers-tasks.md).
