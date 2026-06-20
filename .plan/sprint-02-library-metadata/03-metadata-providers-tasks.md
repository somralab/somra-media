# Sprint 02 — Metadata Provider Tasks

> **Sprint goal:** Fetch rich information from external metadata providers (TMDB/TVDB/MusicBrainz/fanart),
> matching, and image download.
>
> **Related:** [`../tech-stack.md`](../tech-stack.md) §5 · [`../architecture.md`](../architecture.md) §3 · [`01-backend-tasks.md`](./01-backend-tasks.md) · [`../i18n-localization.md`](../i18n-localization.md)

## Responsible Role(s)
- Backend (primary)

## Dependencies
- [`01-backend-tasks.md`](./01-backend-tasks.md) (pre-parsing output), [`02-database-tasks.md`](./02-database-tasks.md) (schema).

## Epics and Tasks

### Epic A: Provider abstraction
- [x] A1 — Common `MetadataProvider` interface (search, detail, images) | Acceptance: providers pluggable.
- [x] A2 — API key management + rate limit + cache | Acceptance: limit exceeded prevented, results cached.

### Epic B: Provider integrations
- [x] B1 — TMDB (movie + series) | Acceptance: correct match rate, tested.
- [x] B2 — TVDB (series) and MusicBrainz (music) basic integration | Acceptance: basic fields fetched.
- [x] B3 — fanart.tv / image provider (poster/backdrop/logo) | Acceptance: images downloaded and cached.

### Epic C-lang: Multilingual metadata
- [x] CL1 — Language parameter in provider queries (user/system locale: en-US/tr-TR) | Acceptance: description/title fetched in preferred language, falls back to en-US if missing. See [`../i18n-localization.md`](../i18n-localization.md) §2.
- [x] CL2 — Multilingual metadata storage/cache strategy | Acceptance: TR+EN text can be stored for same item.

### Epic C: Matching
- [x] C1 — Pre-parsing + provider result matching algorithm (scoring) | Acceptance: common cases match correctly.
- [x] C2 — Manual correction/re-match API | Acceptance: wrong match can be corrected.
- [x] C3 — Periodic metadata refresh job (scheduler) | Acceptance: updates fetched.

## Acceptance Criteria (Sprint Output)
- Scanned items matched with rich metadata + images; manual correction possible.

## Risks
- Provider rate limits and match accuracy → cache + scoring important.

## Out of Scope
- Subtitle download automation — Sprint 06.
