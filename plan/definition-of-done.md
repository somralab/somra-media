# Somra — Definition of Done & Working Rules

> This document contains **binding rules**. A task/PR/sprint cannot be marked "complete" until
> it meets the criteria here. All task files reference this document.

Related: [`project-brief.md`](./project-brief.md) · [`tech-stack.md`](./tech-stack.md) · [`architecture.md`](./architecture.md) · [`i18n-localization.md`](./i18n-localization.md)

---

## 1. Task-Level DoD

A task is "Done" when:

1. Acceptance criteria (in the task file) are met and verified.
2. Code conforms to [`architecture.md`](./architecture.md) module boundaries.
3. Unit tests written and passing; integration tests where relevant; coverage thresholds (§4.1) met.
4. Lint/format clean (Go: `golangci-lint`; FE: ESLint + Prettier).
5. Required documentation/README updates done.
6. At least 1 code review approval (Tech Lead approval if architectural impact).
7. CI green on all stages.
8. **i18n compliance:** All user-facing text resolved via keys (no hardcoded strings); en-US **and** tr-TR keys added. See [`i18n-localization.md`](./i18n-localization.md) §5.

## 2. Sprint-Level DoD

1. All "must" tasks in the sprint goal are Done.
2. Working **incremental release** can be demoed.
3. Regression suite passes (QA).
4. No known critical/high bugs remain.
5. Outputs required for the dependent next sprint are published.

## 3. Code Standards (Summary)

- **Go:** `gofmt`/`goimports`, meaningful package boundaries, error wrapping (`%w`), context usage,
  avoid global state. Exported APIs documented.
- **TypeScript/React:** strict mode, typed API layer, component/hook separation, accessibility.
- **Comments:** No unnecessary "what it does" comments; only intent/constraint/trade-off.
- **Commit/PR:** small, focused; PR description references related task and acceptance criteria.

## 4. Test Policy

| Layer | Expectation |
|---|---|
| Unit | Mandatory for business rules and pure functions. |
| Integration | For boundaries: DB, file scanning, transcode pipeline. |
| E2E | Critical user flows (login, browsing, playback, setup wizard). |
| Performance | Basic measurements for streaming/transcode and scanning (Sprint 04+). |

### 4.1 Coverage Standard — Binding

| Area | Minimum line coverage |
|---|---|
| Core business logic (Go) | **≥ 80%** |
| Critical modules (Go): auth/RBAC, scanning, transcode decision engine, automation import | **≥ 90%** |
| Frontend component tests | **≥ 70%** |
| Frontend critical flows (login, browsing, playback, setup wizard) | **e2e required** (flow coverage, not percentage) |

- Thresholds are **measured and enforced in CI**; PRs below threshold **cannot merge** (see §5).
- Critical module list updated with Tech Lead approval when architecture changes.
- Coverage is a goal, not inflated with meaningless tests; review enforces this.

## 5. CI Gates

Required green stages for PR merge: `lint` → `i18n-check` → `unit-test` → `integration-test` → `coverage-gate` → `build` → `image-build`.
If any gate is red, merge **must not** happen.

> `i18n-check`: missing/unused translation keys and en-US/tr-TR completeness. See [`i18n-localization.md`](./i18n-localization.md) §6.
>
> `coverage-gate`: measures §4.1 thresholds (core ≥80%, critical modules ≥90%, frontend components ≥70%) and produces coverage report; merge blocked below threshold.

## 6. Secure Defaults

- Least privilege; all input validated.
- No secrets in code; environment variable/secret management.
- External provider keys entered by user, stored securely.

## 7. Anti-Drift (Scope Protection) Rules

> Operational counterpart to [`project-brief.md`](./project-brief.md) §9 governance rules.

1. A task cannot exceed its sprint goal; if it does, open a new task/sprint.
2. Work touching [`project-brief.md`](./project-brief.md) §7 out-of-scope items cannot start without brief update.
3. "Nice to have" extras go to backlog, not silently added mid-sprint.
4. Every task file includes: **Goal, Owner role(s), Tasks, Dependencies, Acceptance criteria, Risks, Out of scope** sections.

## 8. Task File Template (Used in all sprints)

```md
# Sprint NN — <Discipline> Tasks

> Sprint goal: ...
> Related: project-brief.md, architecture.md, definition-of-done.md, (prior sprints)

## Owner Role(s)
...

## Dependencies
...

## Epics and Tasks
### Epic A: ...
- [ ] Task A1 — <description> | Acceptance: <criterion>
...

## Acceptance Criteria (Sprint Output)
...

## Risks
...

## Out of Scope
...
```
