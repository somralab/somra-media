# Sprint 06 — Frontend Tasks (Setup Wizard & Settings)

> **Sprint goal:** First-time setup wizard (minimum steps, maximum automation) and settings interfaces.
>
> **Related:** [`01-backend-tasks.md`](./01-backend-tasks.md) · [`../project-brief.md`](../project-brief.md) (setup success criterion: <10 min)

## Responsible Role(s)
- Frontend (primary), UX/UI (flow)

## Dependencies
- [`01-backend-tasks.md`](./01-backend-tasks.md) (detection + settings + onboarding state).

## Epics and Tasks

### Epic A: Setup wizard
- [x] A0 — First step: language selection (tr-TR/en-US, pre-selected from browser language) + set system default language | Acceptance: selected language applies instantly across the wizard. See [`../i18n-localization.md`](../i18n-localization.md) §3.
- [x] A1 — Step flow: create admin → add library → confirm auto-detection/recommendations → finish | Acceptance: working system in <10 min.
- [x] A2 — Show smart defaults with "recommend and apply" experience | Acceptance: user proceeds without manual configuration.
- [x] A3 — Show first scan progress inside the wizard | Acceptance: live feedback.

### Epic B: Settings interface
- [x] B1 — Category-based settings screens (general, library, playback, users) | Acceptance: connected to backend settings API.
- [x] B2 — "Advanced" hidden/visible mode (simple default, detail for those who want it) | Acceptance: minimum configuration philosophy.

## Acceptance Criteria (Sprint Output)
- New user completes the wizard and quickly reaches a working, optimized server.

## Risks
- Wizard too long contradicts the philosophy → keep step count to a minimum.

## Out of Scope
- GPU selection interface — Sprint 07.
