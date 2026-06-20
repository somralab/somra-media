# Sprint 06 — Backend Tasks (Smart Defaults & Settings)

> **Sprint goal:** Implement the "minimum configuration / maximum optimization" philosophy via
> system detection, smart default generation, and centralized settings management.
>
> **Related:** [`../project-brief.md`](../project-brief.md) (success criteria) · [`../architecture.md`](../architecture.md) (Settings & Onboarding) · [`02-frontend-wizard-tasks.md`](./02-frontend-wizard-tasks.md)

## Responsible Role(s)
- Backend (primary), Media Specialist (transcode profile recommendations)

## Dependencies
- Sprint 01 (settings layer), Sprint 02 (library), Sprint 04 (transcode profiles).

## Epics and Tasks

### Epic A: System detection
- [x] A1 — Hardware detection (CPU, memory, available GPU presence) | Acceptance: system profile is derived.
- [x] A2 — Storage/directory detection and validation (media/cache paths) | Acceptance: read/write permissions are verified.

### Epic B: Smart defaults
- [x] B1 — Transcode profile/concurrency recommendation based on hardware (CPU-based; GPU expands in Sprint 07) | Acceptance: reasonable defaults are produced.
- [x] B2 — Recommended library scan/refresh schedule | Acceptance: default job schedule.

### Epic C: Centralized settings management
- [x] C1 — Settings schema + API (category-based, validated) | Acceptance: settings managed from a single place.
- [x] C2 — Setup state (onboarding completed or not) state machine | Acceptance: manages first-time setup flow.
- [x] C3 — System default language setting (tr-TR/en-US) | Acceptance: used when no user preference exists; follows language negotiation priorities. See [`../i18n-localization.md`](../i18n-localization.md) §3.

## Acceptance Criteria (Sprint Output)
- System detects itself, produces smart defaults; settings managed via centralized API.

## Risks
- Wrong defaults cause poor experience → conservative but optimized defaults + override.

## Out of Scope
- GPU-based optimization details — Sprint 07.
