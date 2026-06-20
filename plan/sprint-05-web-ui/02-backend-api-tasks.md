# Sprint 05 — Backend API Tasks (Browsing & Discovery)

> **Sprint goal:** Efficient APIs for library browsing, discovery shelves, search, and filtering.
>
> **Related:** [`01-frontend-tasks.md`](./01-frontend-tasks.md) · Sprint 02 schema/FTS

## Responsible Role(s)
- Backend (primary)

## Dependencies
- Sprint 02 (media/FTS), Sprint 03 (user/watch state/parental controls).

## Epics and Tasks

### Epic A: Browsing APIs
- [x] A1 — Paginated/filtered library list endpoint | Acceptance: fast, uses indexes.
- [x] A2 — Item/season/episode detail endpoints | Acceptance: required data in a single call.

### Epic B: Discovery shelves
- [x] B1 — "Continue", "recently added", recommendation shelf endpoints (user-specific) | Acceptance: based on watch state.
- [x] B2 — Parental control filter application (server-side) | Acceptance: restricted content is not returned.

### Epic C: Search
- [x] C1 — FTS-based search endpoint (debounce-friendly) | Acceptance: low latency.

## Acceptance Criteria (Sprint Output)
- Performant APIs ready to meet frontend browsing/discovery/search needs.

## Risks
- Server-side parental filter consistency → must be applied on every endpoint.

## Out of Scope
- Smart recommendations/ML — out of scope for this plan (simple rule-based shelves).
