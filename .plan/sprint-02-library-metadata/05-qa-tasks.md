# Sprint 02 — QA Tasks

> **Sprint goal:** Verify accuracy and resilience of the scanning and metadata pipeline.
>
> **Related:** [`../definition-of-done.md`](../definition-of-done.md) §4 · [`01-backend-tasks.md`](./01-backend-tasks.md) · [`03-metadata-providers-tasks.md`](./03-metadata-providers-tasks.md)

## Responsible Role(s)
- QA (primary), Backend (test data)

## Dependencies
- Sprint 01 test harnesses.

## Epics and Tasks

### Epic A: Scanning tests
- [x] A1 — Test set of various naming/folder structures | Acceptance: parsing accuracy measured.
- [x] A2 — Corrupt/unsupported file resilience test | Acceptance: scan continues without crashing.
- [x] A3 — Large library performance test (basic) | Acceptance: time/resource reported.

### Epic B: Metadata tests
- [x] B1 — Provider matching accuracy (mock + real sample) | Acceptance: acceptable match rate.
- [x] B2 — Rate limit/cache behavior test | Acceptance: limit not exceeded.

### Epic C: Regression
- [x] C1 — Sprint 02 regression package | Acceptance: runs in CI.

## Acceptance Criteria (Sprint Output)
- Scan/metadata flows under test coverage; no critical bugs.

## Risks
- External provider tests are fragile → balance of mock + limited real tests.

## Out of Scope
- Playback/streaming tests — Sprint 04.
