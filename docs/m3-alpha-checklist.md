# M3 Alpha Checklist

M3 = usable alpha: login, browse, search, detail, play (Sprint 04 player + Sprint 05 UI).

## Functional

- [x] Authenticated home with continue / recent / recommended shelves
- [x] Library browse with grid/list and virtual scrolling
- [x] FTS search in header
- [x] Media detail with play, resume, favorite, watchlist
- [x] Parental filter enforced server-side
- [x] Four themes persist (localStorage / profile)

## Quality

- [x] Unit/integration tests for browse API
- [x] Component tests ≥70% on new browse/search components
- [x] E2E specs for critical path (login → browse → search → detail)
- [x] `make lint test i18n-check coverage build` green

## Out of scope (M3)

- Setup wizard (S06)
- Hardware transcode accel (S07)
- Indexer/automation (S09)
