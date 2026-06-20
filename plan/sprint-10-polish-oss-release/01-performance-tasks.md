# Sprint 10 — Performance & Hardening Tasks

> **Sprint goal:** Harden the entire system for performance, stability, and resource efficiency.
> Quality bar for M6 (1.0).
>
> **Related:** [`../project-brief.md`](../project-brief.md) (success criteria: efficiency on home hardware) · [`../roadmap.md`](../roadmap.md) (M6)

## Responsible Role(s)
- Tech Lead (primary), Backend, Media Specialist, Frontend

## Dependencies
- Outputs from all previous sprints.

## Epics and Tasks

### Epic A: Backend performance
- [ ] A1 — Profiling (CPU/memory) + hot path optimization | Acceptance: measured improvement.
- [ ] A2 — DB query/index optimization (large library) | Acceptance: slow queries resolved.
- [ ] A3 — Resource efficiency of scan/metadata/automation jobs | Acceptance: stable on home hardware.

### Epic B: Streaming efficiency
- [ ] B1 — Transcode/HW session efficiency and latency | Acceptance: target session count met.

### Epic C: Frontend performance
- [ ] C1 — Bundle size, lazy load, render optimization | Acceptance: good load/interaction metrics.

### Epic D: Stability
- [ ] D1 — Long-running (soak) test + memory leak check | Acceptance: no leaks.

## Acceptance Criteria (Sprint Output)
- System runs stably and efficiently at target load on home server hardware (e.g. 4 cores/8GB).

## Risks
- Late-discovered bottlenecks → early profiling habit.

## Out of Scope
- New feature development — this sprint is polish-focused.
