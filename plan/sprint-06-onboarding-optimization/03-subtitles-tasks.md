# Sprint 06 — Subtitle Automation Tasks (Bazarr Functionality)

> **Sprint goal:** Automatic subtitle search/download (Bazarr-like): missing subtitle detection,
> provider integration, and matching.
>
> **Related:** [`../architecture.md`](../architecture.md) (Metadata/plugin) · Sprint 04 (subtitle playback) · [`../project-brief.md`](../project-brief.md) (scope)

## Responsible Role(s)
- Backend (primary)

## Dependencies
- Sprint 02 (library/item), Sprint 04 (subtitle handling).

## Epics and Tasks

### Epic A: Subtitle provider integration
- [x] A1 — Common interface + integration for open subtitle provider(s) | Acceptance: search/download works.
- [x] A2 — Language preference and quality/match scoring | Acceptance: correct subtitle is selected.

### Epic B: Automation
- [x] B1 — Missing subtitle detection (based on user language preference) | Acceptance: gaps are reported.
- [x] B2 — Periodic automatic download job (scheduler) | Acceptance: subtitles arrive automatically for new content.
- [x] B3 — Manual subtitle search/upload | Acceptance: user can override.

### Epic C: UI integration
- [x] C1 — Subtitle management on detail page (coordinated with frontend) | Acceptance: subtitle status is visible.

## Acceptance Criteria (Sprint Output)
- System detects missing subtitles and downloads them automatically; manual management is possible.

## Risks
- Provider rate limits/licensing → cache + compatible provider selection.

## Out of Scope
- Subtitle sync/AI translation — out of scope for this plan.
