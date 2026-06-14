# Sprint 02 — Definition of Done Checklist

## Backend

- [x] Library CRUD API (movie/tv/music, multi-root paths)
- [x] Full + incremental scan jobs with progress reporting
- [x] ffprobe technical metadata → DB
- [x] Filename/folder pre-parse (title, year, SxxEyy)
- [x] Filesystem watch with debounce → incremental scan
- [x] MetadataProvider interface + TMDB + TVDB/MusicBrainz/fanart stubs
- [x] Match scoring + manual rematch API
- [x] Localized metadata storage (en-US + tr-TR)
- [x] SSRF guard on outbound provider calls
- [x] Path traversal guard on library roots

## Frontend

- [x] Library list/create UI
- [x] Multi path input
- [x] Scan trigger + SSE progress
- [x] Scan history view
- [x] Media item list with poster/title/year
- [x] Manual rematch candidate preview
- [x] i18n library.* en-US + tr-TR

## QA

- [x] Naming fixture set + parser tests
- [x] SSRF + match unit tests
- [x] DB migration + repository tests
- [x] OpenAPI spec updated

## Gates

Run before merge:

```bash
make lint test i18n-check coverage build
```

## Demo

1. Mount media under `deploy/media/` (or any host path).
2. Create a library pointing at that path via UI or API.
3. Trigger full scan; observe SSE progress and scan history.
4. Set `SOMRA_TMDB_API_KEY` for live TMDB matching (optional).
