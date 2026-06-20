# Sprint 05 — Frontend Tasks (Library Browsing & Discovery)

> **Sprint goal:** Rich library browsing, detail pages, search/filter, and home page discovery.
> Combined with the Sprint 04 player, completes M3 (usable alpha).
>
> **Related:** Sprint 02 (metadata), Sprint 04 (player) · [`03-design-tasks.md`](./03-design-tasks.md) · [`02-backend-api-tasks.md`](./02-backend-api-tasks.md)

## Responsible Role(s)
- Frontend (primary), UX/UI (design)

## Dependencies
- Sprint 02 metadata + Sprint 03 user + Sprint 04 playback APIs.

## Epics and Tasks

### Epic A: Home page & discovery
- [x] A1 — Home page: "continue watching", "recently added", recommendation shelves | Acceptance: user-specific shelves.
- [x] A2 — Library view (grid/list, poster, lazy load, virtual list) | Acceptance: large library is smooth.

### Epic B: Detail pages
- [x] B1 — Movie/series/season/episode detail pages (metadata, cast, images) | Acceptance: rich presentation.
- [x] B2 — "Play" / "continue" / favorite / watchlist actions | Acceptance: connected to backend.

### Epic C: Search & filter
- [x] C1 — Quick search (FTS) + result preview | Acceptance: instant results.
- [x] C2 — Filter/sort (genre, year, watch status) | Acceptance: combined filters.

### Epic D: Responsiveness & state
- [x] D1 — Loading/empty/error states + skeleton | Acceptance: consistent UX.
- [x] D2 — Content filtering by parental controls (UI) | Acceptance: child profile sees restricted content.
- [x] D3 — All text delivered via i18n keys (en-US + tr-TR) and text length/overflow resilience | Acceptance: no hardcoded text, layout holds across TR↔EN. See [`../i18n-localization.md`](../i18n-localization.md) §5.
- [x] D4 — Four themes (Cinematic default/Aurora/Noir/Minimal) applied to all screens + quick theme switcher (menu/settings) | Acceptance: theme changes instantly, consistent across all screens, selection persists (sessionless `localStorage`, session profile). See [`04-frontend-tasks.md`](../sprint-01-foundation/04-frontend-tasks.md) C1b and [`03-design-tasks.md`](./03-design-tasks.md) Epic D.

## Acceptance Criteria (Sprint Output)
- User browses the library, searches, views details, and plays content (end-to-end alpha flow).

## Risks
- Large library performance → virtual list + pagination required.

## Out of Scope
- Setup wizard — Sprint 06.
