# Somra — Planning Index

> **Start here.** Dashboard for humans, agents, and contributors.  
> Binding spec lives under [`.plan/`](./); working notes live under [`../notes/`](../notes/).

Related: [`project-brief.md`](./project-brief.md) · [`roadmap.md`](./roadmap.md) · [`definition-of-done.md`](./definition-of-done.md)

---

## Active sprint

| Field | Value |
|---|---|
| **Sprint** | **02 — Library & Metadata** |
| **Status** | In progress (Sprint 01 / M1 complete) |
| **Milestone target** | M2 — login + scanned library with metadata (with Sprint 03) |
| **Folder** | [`sprint-02-library-metadata/`](./sprint-02-library-metadata/) |
| **DoD checklist** | [`../docs/sprint-02-dod-checklist.md`](../docs/sprint-02-dod-checklist.md) |

When the active sprint changes, update this table only — details stay in [`roadmap.md`](./roadmap.md).

---

## Documentation hierarchy (binding order)

On conflict, higher rows win:

| Priority | Location | Role |
|---|---|---|
| 1 | [`.plan/project-brief.md`](./project-brief.md) | Vision, scope in/out, governance |
| 2 | [`.plan/architecture.md`](./architecture.md) · [`.plan/tech-stack.md`](./tech-stack.md) | Module boundaries, closed tech decisions |
| 3 | [`.plan/definition-of-done.md`](./definition-of-done.md) · [`.plan/i18n-localization.md`](./i18n-localization.md) | DoD, i18n rules |
| 4 | Sprint task files under [`.plan/sprint-XX-*/`](./) | Acceptance criteria per epic |
| 5 | [`docs/`](../docs/) | Published operational docs (testing, checklists, design) |
| 6 | [`api/openapi.yaml`](../api/openapi.yaml) | HTTP API contract (design-first) |
| 7 | [`notes/`](../notes/) | **Non-binding** — briefings, drafts; promote decisions to `.plan/` or `docs/` |

---

## Core planning documents

| Document | Purpose |
|---|---|
| [`project-brief.md`](./project-brief.md) | Single source of truth for product scope |
| [`roadmap.md`](./roadmap.md) | 10-sprint plan, dependencies, milestones |
| [`architecture.md`](./architecture.md) | Modules, data flow, boundaries |
| [`tech-stack.md`](./tech-stack.md) | Closed technology decisions |
| [`definition-of-done.md`](./definition-of-done.md) | Task / sprint / CI DoD |
| [`i18n-localization.md`](./i18n-localization.md) | Binding i18n rules |
| [`ideal-team.md`](./ideal-team.md) | Roles and sprint ownership |

Agent entry: [`../AGENTS.md`](../AGENTS.md)

---

## Milestones (summary)

Full definitions: [`roadmap.md`](./roadmap.md) §3.

| Milestone | Sprints | Outcome |
|---|---|---|
| **M1** | 01 | Runnable skeleton via `docker run` — **done** |
| **M2** | 02–03 | Login + scanned library with metadata |
| **M3** | 04–05 | Browser playback incl. transcode — first alpha |
| **M4** | 06–07 | Zero-config setup + HW acceleration — beta candidate |
| **M5** | 08–09 | Requests + automation + indexers |
| **M6** | 10 | OSS release — 1.0 |

---

## Sprint task folders

| Sprint | Folder | Main output |
|---|---|---|
| 01 | [`sprint-01-foundation/`](./sprint-01-foundation/) | Skeleton + CI/CD + Docker + API contract |
| 02 | [`sprint-02-library-metadata/`](./sprint-02-library-metadata/) | Scanning + metadata + file watching |
| 03 | [`sprint-03-auth-users/`](./sprint-03-auth-users/) | Multi-user, RBAC, profiles, watch state |
| 04 | [`sprint-04-streaming-transcode/`](./sprint-04-streaming-transcode/) | Direct play, transcode, HLS/DASH |
| 05 | [`sprint-05-web-ui/`](./sprint-05-web-ui/) | Library UI, player, search |
| 06 | [`sprint-06-onboarding-optimization/`](./sprint-06-onboarding-optimization/) | Setup wizard, smart defaults, subtitles |
| 07 | [`sprint-07-hardware-acceleration/`](./sprint-07-hardware-acceleration/) | QSV/NVENC/VAAPI/AMF |
| 08 | [`sprint-08-request-management/`](./sprint-08-request-management/) | Request/approval + notifications |
| 09 | [`sprint-09-automation-indexers/`](./sprint-09-automation-indexers/) | Plugin architecture + indexers + downloads |
| 10 | [`sprint-10-polish-oss-release/`](./sprint-10-polish-oss-release/) | Performance, docs, security audit, OSS release |

---

## Execution workflow

```
.plan/ task spec  →  GitHub Issue (status, assignee)  →  PR  →  CI gates
notes/ briefing   →  decision finalized  →  update .plan/ or docs/
```

- Open issues with the [**Sprint task**](../.github/ISSUE_TEMPLATE/sprint_task.md) template.
- Every issue must reference a `.plan/` task file and acceptance criteria.
- PR descriptions must reference the issue and DoD checklist items.

---

## Operational docs (`docs/`)

| Document | Purpose |
|---|---|
| [`docs/testing-strategy.md`](../docs/testing-strategy.md) | Test layers and expectations |
| [`docs/issue-severity.md`](../docs/issue-severity.md) | Bug severity definitions |
| [`docs/sprint-NN-dod-checklist.md`](../docs/) | Per-sprint DoD verification |
| [`docs/m3-alpha-checklist.md`](../docs/m3-alpha-checklist.md) | Alpha readiness |
| [`docs/m4-beta-checklist.md`](../docs/m4-beta-checklist.md) | Beta readiness |
