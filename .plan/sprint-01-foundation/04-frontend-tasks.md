# Sprint 01 — Frontend Tasks

> **Sprint goal:** React + Vite SPA skeleton, API client layer, and design system foundation.
>
> **Related:** [`../tech-stack.md`](../tech-stack.md) · [`../architecture.md`](../architecture.md) §5 · [`../definition-of-done.md`](../definition-of-done.md)

## Responsible Role(s)
- Frontend (primary), UX/UI (design system oversight)

## Dependencies
- [`01-architecture-tasks.md`](./01-architecture-tasks.md) Epic C (API endpoints) — `health`/`version` consumable.

## Epics and Tasks

### Epic A: SPA skeleton
- [x] A1 — Vite + React + TypeScript (strict) + **Tailwind CSS + Radix UI** project setup | Acceptance: dev/build work, lint clean. See [`../tech-stack.md`](../tech-stack.md) §2.
- [x] A2 — Routing (router) and page layout skeleton | Acceptance: basic layout + 2 sample routes.
- [x] A3 — Environment/config management (API base URL) | Acceptance: build-time/run-time config strategy.

### Epic B: API client layer
- [x] B1 — Typed HTTP client + error handling | Acceptance: `health`/`version` calls return typed responses.
- [x] B2 — State management integration: **TanStack Query** (server state) + **Zustand** (UI state) | Acceptance: sample query/cache + global UI state work.
- [x] B3 — Real-time event (WS/SSE) client skeleton | Acceptance: connection established and events received.

### Epic C: Design system foundation
- [x] C1 — Token-based design system (color, typography, spacing) | Acceptance: all styles via tokens.
- [x] C1b — **Dynamic theme infrastructure (theme provider):** runtime-switchable, token-set-based multi-theme; original theme sets **Cinematic (default)**, Aurora, Noir, Minimal (keys: `cinematic`/`aurora`/`noir`/`minimal`); persistence of selection (unauthenticated: `localStorage`; authenticated: user profile — Sprint 03) | Acceptance: theme changes instantly and persists on reload. New themes must be addable **only by adding a token set** (no code changes required).
- [x] C2 — Base components (button, input, card, modal, toast) | Acceptance: accessible, documented.
- [x] C3 — i18n infrastructure: library setup, namespace/key structure, en-US (source) + tr-TR translation files, language switcher, date/number formatting with `Intl`, browser language detection | Acceptance: language switching works, no hardcoded text, missing keys warn at development time. See [`../i18n-localization.md`](../i18n-localization.md).

## Acceptance Criteria (Sprint Output)
- SPA compiles and displays backend `health`/`version` information.
- Design system base components are ready for use.

## Risks
- Design system decisions affect all UI → early consistency is important.

## Out of Scope
- Real feature screens (library, player) — Sprint 05.
