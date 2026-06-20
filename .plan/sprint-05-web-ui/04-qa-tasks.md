# Sprint 05 — QA Tasks

> **Sprint goal:** Verify the end-to-end user flow (login → browsing → search → playback) and M3
> alpha quality.
>
> **Related:** [`../definition-of-done.md`](../definition-of-done.md) · [`../roadmap.md`](../roadmap.md) (M3)

## Responsible Role(s)
- QA (primary)

## Dependencies
- This sprint's frontend/backend + Sprint 04 player.

## Epics and Tasks

### Epic A: E2E flows
- [x] A1 — Login → library → detail → playback e2e | Acceptance: critical path passes.
- [x] A2 — Search/filter/shelf accuracy | Acceptance: results are consistent.

### Epic B: Compatibility & accessibility
- [x] B1 — Responsive/browser test | Acceptance: desktop/mobile browsers.
- [x] B2 — Accessibility check (keyboard, contrast) | Acceptance: basic WCAG (in all four themes).
- [x] B3 — Theme test: consistency across four themes + theme persistence | Acceptance: theme changes, is remembered, contrast is preserved.

### Epic C: Alpha acceptance
- [x] C1 — M3 alpha acceptance checklist | Acceptance: ready for alpha demo.

## Acceptance Criteria (Sprint Output)
- End-to-end flow is under test coverage; M3 alpha criteria are met.

## Risks
- Performance regression on large library → measured testing.

## Out of Scope
- Onboarding wizard testing — Sprint 06.
