# Sprint 03 — Backend Tasks (Identity & RBAC)

> **Sprint goal:** Multi-user, authentication, RBAC, profiles, parental controls, and
> watch state (resume) backend.
>
> **Related:** [`../architecture.md`](../architecture.md) §3/§5 · [`../project-brief.md`](../project-brief.md) (Identity decision) · [`04-security-tasks.md`](./04-security-tasks.md)

## Responsible Role(s)
- Backend (primary), Tech Lead (session strategy)

## Dependencies
- Sprint 01 (API/session decision, data layer), Sprint 02 (media items — for watch state).

## Epics and Tasks

### Epic A: Authentication
- [x] A1 — User registration/login, password hash (strong algorithm), session/token management | Acceptance: secure flow, tested.
- [x] A2 — Session refresh/logout, multi-device sessions | Acceptance: sessions manageable.
- [x] A3 — Admin creation flow on first setup | Acceptance: compatible with Sprint 06 onboarding.

### Epic B: RBAC and profiles
- [x] B1 — Role/permission model (admin, user, child) + authorization middleware | Acceptance: protected endpoints filtered by permission.
- [x] B2 — User profiles (avatar, language preference `tr-TR`/`en-US`, **UI theme** `cinematic`/`aurora`/`noir`/`minimal`, preferences) | Acceptance: profile CRUD; language preference has highest priority in locale negotiation; theme preference persisted (default `cinematic`). See [`../i18n-localization.md`](../i18n-localization.md) §3.
- [x] B3 — Parental controls: age rating limit, content restriction | Acceptance: restricted content hidden on child profile.

### Epic C: Watch state
- [x] C1 — Watch progress/resume and "watched" state | Acceptance: resume from where left off.
- [x] C2 — Per-user favorites/watchlist | Acceptance: CRUD + filter.

## Acceptance Criteria (Sprint Output)
- Multiple users log in; roles and parental restrictions applied; watch state maintained.
- [`04-security-tasks.md`](./04-security-tasks.md) requirements met.

## Risks
- Security critical → also validated in Sprint 10 security audit.

## Out of Scope
- External identity (OIDC/LDAP) — out of scope for this plan (future).
