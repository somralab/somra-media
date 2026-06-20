# Sprint 03 — Frontend Tasks (Identity & User UI)

> **Sprint goal:** Login/logout, user management, profile, and parental control interfaces.
>
> **Related:** Sprint 01 SPA skeleton · [`01-backend-tasks.md`](./01-backend-tasks.md)

## Responsible Role(s)
- Frontend (primary)

## Dependencies
- This sprint backend identity/RBAC APIs.

## Epics and Tasks

### Epic A: Identity flows
- [x] A1 — Login/logout screens + session management (token storage, refresh) | Acceptance: secure, protected routes.
- [x] A2 — Protected route / permission-based UI display | Acceptance: unauthorized user sees limited UI.

### Epic B: User management (admin)
- [x] B1 — User list/create/edit + role assignment | Acceptance: connected to RBAC API.
- [x] B2 — Parental control settings UI | Acceptance: rating limit configurable.

### Epic C: Profile
- [x] C1 — Profile editing (language selection tr-TR/en-US, **theme selection** Cinematic/Aurora/Noir/Minimal, avatar, preferences) | Acceptance: language and theme changes apply instantly to UI and persist; browser language auto-preselect; theme from `localStorage` when unauthenticated migrates to profile on login. See [`../i18n-localization.md`](../i18n-localization.md) §3 and [`04-frontend-tasks.md`](../sprint-01-foundation/04-frontend-tasks.md) C1b.

## Acceptance Criteria (Sprint Output)
- User logs in, admin manages users, parental controls configured.

## Risks
- Token storage security → secure practices (aligned with Sprint 03 security tasks).

## Out of Scope
- Player/browsing screens — Sprint 05.
