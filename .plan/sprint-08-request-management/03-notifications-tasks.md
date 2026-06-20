# Sprint 08 — Notification Tasks

> **Sprint goal:** Event-driven notification infrastructure (webhook, Discord, email) and
> integration with request/system events.
>
> **Related:** [`../tech-stack.md`](../tech-stack.md) §5 · [`01-backend-tasks.md`](./01-backend-tasks.md)

## Responsible Role(s)
- Backend (primary)

## Dependencies
- [`01-backend-tasks.md`](./01-backend-tasks.md) (request events), scheduler/event infrastructure (Sprint 01).

## Epics and Tasks

### Epic A: Notification infrastructure
- [x] A1 — Event → notification abstraction (extensible channels) | Acceptance: new channel easily added.
- [x] A2 — Templates and language support (i18n): select tr-TR/en-US template per recipient language preference, en-US fallback if missing | Acceptance: notification sent in recipient's language. See [`../i18n-localization.md`](../i18n-localization.md) §2.

### Epic B: Channel integrations
- [x] B1 — Webhook (generic) | Acceptance: configurable webhook triggered.
- [x] B2 — Discord + email (SMTP) | Acceptance: test send successful.

### Epic C: Event binding
- [x] C1 — Request events (created/approved/completed) and system events (errors) | Acceptance: notification on correct event.
- [x] C2 — User/admin notification preferences | Acceptance: subscription management.

## Acceptance Criteria (Sprint Output)
- Events notified via configurable channels; preferences managed.

## Risks
- Spam/excessive notifications → preferences + debounce.

## Out of Scope
- Push/native mobile notifications — out of scope for this plan.
