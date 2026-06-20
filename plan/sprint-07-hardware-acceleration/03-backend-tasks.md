# Sprint 07 — Backend Tasks (HW Settings & Integration)

> **Sprint goal:** Integrate hardware acceleration settings with system settings and onboarding;
> monitoring/telemetry.
>
> **Related:** Sprint 06 (smart defaults/settings) · [`01-media-streaming-tasks.md`](./01-media-streaming-tasks.md)

## Responsible Role(s)
- Backend (primary), Media Specialist (parameters)

## Dependencies
- [`01-media-streaming-tasks.md`](./01-media-streaming-tasks.md), Sprint 06 settings layer.

## Epics and Tasks

### Epic A: Settings integration
- [x] A1 — HW acceleration setting (automatic/force on/off, accelerator selection) | Acceptance: setting in API.
- [x] A2 — Add GPU detection/recommendation step to onboarding | Acceptance: wizard recommends GPU.

### Epic B: Monitoring
- [x] B1 — HW session metrics (usage, fallback rate) | Acceptance: telemetry collected.
- [x] B2 — HW error/fallback logging and diagnostics | Acceptance: troubleshooting is easier.

## Acceptance Criteria (Sprint Output)
- HW acceleration manageable from UI; detection/recommendation integrated into onboarding; telemetry available.

## Risks
- Wrong settings can break playback → safe default is "automatic + fallback".

## Out of Scope
- New codec research (AV1, etc.) — future roadmap.
