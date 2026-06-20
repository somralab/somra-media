# Sprint 03 — Database Tasks (User Schema)

> **Sprint goal:** User, role, session, profile, parental control, and watch state schema.
>
> **Related:** [`../architecture.md`](../architecture.md) §4 · [`01-backend-tasks.md`](./01-backend-tasks.md)

## Responsible Role(s)
- Backend (primary), Tech Lead (review)

## Dependencies
- Sprint 01 migration infrastructure, Sprint 02 media schema.

## Epics and Tasks

### Epic A: Identity schema
- [x] A1 — `user`, `role`, `permission`, `user_role` tables | Acceptance: migration + integrity.
- [x] A2 — `session`/token table (multi-device) | Acceptance: expiry/revocation fields.

### Epic B: Profile and controls
- [x] B1 — `user_profile` (preferences, **language/locale**, **UI theme**, avatar), parental control fields (rating limit) | Acceptance: child profile restrictions queryable; language and theme preferences stored (theme default `cinematic`). See [`../i18n-localization.md`](../i18n-localization.md).

### Epic C: Watch state
- [x] C1 — `watch_state` (progress, watched), `favorite`, `watchlist` tables | Acceptance: user+item indexes.

## Acceptance Criteria (Sprint Output)
- Schema set up via migrations; backend flows work consistently.

## Risks
- Permission model expands in later sprints → flexible design.

## Out of Scope
- Request management tables — Sprint 08.
