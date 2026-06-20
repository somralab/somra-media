# Sprint 07 — QA Tasks

> **Sprint goal:** Validate hardware acceleration correctness, performance gains, and fallback
> resilience. M4 beta candidate check.
>
> **Related:** [`../definition-of-done.md`](../definition-of-done.md) · [`01-media-streaming-tasks.md`](./01-media-streaming-tasks.md) · [`../roadmap.md`](../roadmap.md) (M4)

## Responsible Role(s)
- QA (primary), Media Specialist (hardware environment)

## Dependencies
- This sprint's media/devops/backend outputs.

## Epics and Tasks

### Epic A: Functional
- [x] A1 — HW transcode correctness (video/audio quality) | Acceptance: output is acceptable.
- [x] A2 — Automatic selection and HW→SW fallback test | Acceptance: seamless transition.

### Epic B: Performance
- [x] B1 — CPU vs HW comparison (resources/concurrent sessions) | Acceptance: significant gain reported.
- [x] B2 — Hardware session limit stress test | Acceptance: limit is maintained.

### Epic C: Beta acceptance
- [x] C1 — M4 beta candidate checklist | Acceptance: beta criteria met.

## Acceptance Criteria (Sprint Output)
- HW path covered by tests; performance gain measured; fallback safe; M4 criteria satisfied.

## Risks
- Hardware diversity limits test environment → deep testing on priority hardware.

## Out of Scope
- Certification on all GPU models — best-effort.
