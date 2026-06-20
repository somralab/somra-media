# Sprint 09 — Indexer Integration Tasks (Prowlarr Functionality)

> **Sprint goal:** Torrent + Usenet indexer integration (Prowlarr functionality): search, capability
> query, and result normalization — via the plugin framework.
>
> **Related:** [`01-plugin-architecture-tasks.md`](./01-plugin-architecture-tasks.md) · [`../project-brief.md`](../project-brief.md) §4 (content acquisition: later sprint)

## Responsible Role(s)
- Backend (primary)

## Dependencies
- [`01-plugin-architecture-tasks.md`](./01-plugin-architecture-tasks.md) (plugin contract).

## Epics and Tasks

### Epic A: Indexer abstraction
- [ ] A1 — Common search/capability interface for Torrent + Usenet | Acceptance: type-agnostic result model.
- [ ] A2 — Indexer definition/schema (Torznab/Newznab-like compatibility) | Acceptance: common protocols supported.

### Epic B: Search & normalization
- [ ] B1 — Parallel search across multiple indexers + result merging | Acceptance: deduplicated results.
- [ ] B2 — Result parsing (quality, resolution, size, seeders) | Acceptance: fields ready for scoring.

### Epic C: Management
- [ ] C1 — Indexer add/test/disable API | Acceptance: connection test works.

## Acceptance Criteria (Sprint Output)
- Multiple indexers can be added and searched; results normalized and tested.

## Risks
- Protocol/indexer diversity + legal sensitivity → compatibility + isolation.

## Out of Scope
- Grab/download execution — [`03-download-client-tasks.md`](./03-download-client-tasks.md).
