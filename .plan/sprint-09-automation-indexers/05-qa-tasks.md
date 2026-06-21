# Sprint 09 — QA Tasks

> **Sprint goal:** Validate plugin architecture, indexer integration, and end-to-end automation.
> M5 (entry to full parity phase) check.
>
> **Related:** [`../definition-of-done.md`](../definition-of-done.md) · [`../roadmap.md`](../roadmap.md) (M5) · [`03-download-client-tasks.md`](./03-download-client-tasks.md)

## Responsible Role(s)
- QA (primary)

## Dependencies
- This sprint's full outputs.

## Epics and Tasks

### Epic A: Plugin & isolation
- [x] A1 — Core full functionality test without plugins | Acceptance: system works without plugins.
- [x] A2 — Plugin add/configure/disable test | Acceptance: lifecycle correct.

### Epic B: End-to-end automation
- [x] B1 — Request → grab → download → import → library e2e | Acceptance: flow completes.
- [x] B2 — Quality profile selection accuracy | Acceptance: correct release selected.
- [x] B3 — Error/interruption recovery test | Acceptance: system remains consistent.

### Epic C: M5 acceptance
- [x] C1 — M5 checklist | Acceptance: full parity phase criteria met.

## Acceptance Criteria (Sprint Output)
- Automation and isolation covered by tests; end-to-end flow reliable; M5 criteria met.

## Risks
- External client/indexer dependency → mock + controlled real test environment.

## Out of Scope
- Production security audit — Sprint 10.
