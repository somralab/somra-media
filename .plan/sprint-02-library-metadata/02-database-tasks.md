# Sprint 02 — Database Tasks (Media Schema)

> **Sprint goal:** Media domain schema (library, item, season/episode, file, metadata, people).
>
> **Related:** [`../architecture.md`](../architecture.md) §4 · Sprint 01 [`../sprint-01-foundation/03-database-tasks.md`](../sprint-01-foundation/03-database-tasks.md)

## Responsible Role(s)
- Backend (primary), Tech Lead (schema review)

## Dependencies
- Sprint 01 migration infrastructure.

## Epics and Tasks

### Epic A: Core media schema
- [x] A1 — `library`, `media_item` (movie/series/album), `season`, `episode`, `media_file` tables | Acceptance: migration + referential integrity.
- [x] A2 — Technical metadata tables (codec, stream information) | Acceptance: ffprobe output stored.
- [x] A3 — Visual asset (poster/backdrop) reference table | Acceptance: file/cache path stored.

### Epic B: Enrichment schema
- [x] B1 — Person (actor/director), genre, tag tables + relationships | Acceptance: many-to-many relationships.
- [x] B2 — External provider identifiers (TMDB/TVDB id mappings) | Acceptance: supports re-matching.
- [x] B3 — Multilingual text storage (locale-based for title/description; en-US + tr-TR) | Acceptance: text queryable by language for same item. See [`../i18n-localization.md`](../i18n-localization.md).

### Epic C: Indexing and performance
- [x] C1 — Indexes for search/filter | Acceptance: common queries fast.
- [x] C2 — Full-text search foundation (SQLite FTS) | Acceptance: title search works.

## Acceptance Criteria (Sprint Output)
- Schema set up via migrations; scan and metadata data written consistently.

## Risks
- Schema affects later sprints → relationships must be designed carefully.

## Out of Scope
- User/watch state tables — Sprint 03.
