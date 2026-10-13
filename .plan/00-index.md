# Somra — Planning Index

> **Start here.** Dashboard for humans, agents, and contributors.  
> Binding spec lives under `[.plan/](./)`; working notes live under `[../notes/](../notes/)`.

Related: `[project-brief.md](./project-brief.md)` · `[roadmap.md](./roadmap.md)` · `[definition-of-done.md](./definition-of-done.md)`

---

## Active sprint

> **Read [`current-state.md`](./current-state.md) first.** It records what actually works
> (audit 2026-10-13). Sprint checkboxes below record what was planned; 26 tasks were reopened.

| Field | Value |
|---|---|
| **Sprint** | **11 — Playback & Data Foundation** |
| **Status** | **Active** |
| **Milestone target** | **M5.5** — core experience & performance (Sprints 11–13) |
| **Folder** | [`sprint-11-playback-foundation/`](./sprint-11-playback-foundation/) |
| **Dependencies** | Reopened tasks in Sprints 02, 04, 05, 07 · [`definition-of-done.md`](./definition-of-done.md) §1 |

### Short-term plan (M5.5)

| Order | Sprint | Folder | Focus | Estimate |
|---|---|---|---|---|
| 1 | 11 — Playback & Data Foundation | [`sprint-11-playback-foundation/`](./sprint-11-playback-foundation/01-tasks.md) | Playback works in all browsers, transcode lifetime, rescans keep data, TV model, CI green | ~2 weeks |
| 2 | 12 — Streaming-App Experience | [`sprint-12-streaming-experience/`](./sprint-12-streaming-experience/01-tasks.md) | Netflix-like UI: design system, consumer/admin split, hero + rows, detail, full-screen player, image pipeline | ~3 weeks |
| 3 | 13 — Performance at Scale | [`sprint-13-performance-scale/`](./sprint-13-performance-scale/01-tasks.md) | Budgets on 4c/8GB: ABR, seek-ahead, keyset pagination, scan speed, image/bundle budgets | ~2 weeks |
| 4 | 10 — Polish & Open Source (resumed) | [`sprint-10-polish-oss-release/`](./sprint-10-polish-oss-release/) | Security audit, release checklist, first `v*` tag | after M5.5 |

**On hold:** Sprint 08 (requests) and Sprint 09 (automation / indexers / download clients).
Scope is unchanged ([`project-brief.md`](./project-brief.md) §6). They resume after M5.5; whether
they ship before or after the 1.0 release (Sprint 10) is decided at that point.
While on hold, only fix what breaks the build or tests.

When the active sprint changes, update this section only — details stay in [`roadmap.md`](./roadmap.md).

---

## Sprint status

| Sprint | Name | Milestone | Status |
|---|---|---|---|
| 01 | Foundation & Architecture | M1 | **Done** |
| 02 | Library & Metadata | M2 (partial) | **Partial** — 8 tasks reopened (TV metadata, images, watcher, parser) |
| 03 | Auth & Users | M2 | **Done** (refresh race tracked in Sprint 11) |
| 04 | Streaming & Transcode | M3 (partial) | **Partial** — 7 tasks reopened (transcode lifetime, ABR, tracks, subtitles, seek) |
| 05 | Web UI | M3 | **Partial** — 7 tasks reopened (grid, episodes, icons, player controls, responsive, a11y) |
| 06 | Onboarding & Optimization | M4 (partial) | **Done** |
| 07 | Hardware Acceleration | M4 | **Partial** — 4 tasks reopened (HW ffmpeg args invalid) |
| 08 | Request Management | M5 (partial) | **On hold** (skeleton implemented) |
| 09 | Automation & Indexers | M5 | **On hold** (skeleton; default build ships stub plugins) |
| 10 | Polish & Open Source | M6 | **Paused** — docs/tooling landed; no release tagged; resumes after M5.5 |
| 11 | Playback & Data Foundation | M5.5 | **Active** |
| 12 | Streaming-App Experience | M5.5 | Planned |
| 13 | Performance at Scale | M5.5 | Planned |

---

## Documentation hierarchy (binding order)

On conflict, higher rows win:


| Priority | Location                                                                                                          | Role                                                                          |
| -------- | ----------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------- |
| 1        | `[.plan/project-brief.md](./project-brief.md)`                                                                    | Vision, scope in/out, governance                                              |
| 2        | `[.plan/architecture.md](./architecture.md)` · `[.plan/tech-stack.md](./tech-stack.md)`                           | Module boundaries, closed tech decisions                                      |
| 3        | `[.plan/definition-of-done.md](./definition-of-done.md)` · `[.plan/i18n-localization.md](./i18n-localization.md)` | DoD, i18n rules                                                               |
| 4        | Sprint task files under `[.plan/sprint-XX-*/](./)`                                                                | Acceptance criteria per epic                                                  |
| 5        | `[docs/](../docs/)`                                                                                               | Published operational docs (testing, checklists, design)                      |
| 6        | `[api/openapi.yaml](../api/openapi.yaml)`                                                                         | HTTP API contract (design-first)                                              |
| 7        | `[notes/](../notes/)`                                                                                             | **Non-binding** — briefings, drafts; promote decisions to `.plan/` or `docs/` |


---

## Core planning documents


| Document                                           | Purpose                                  |
| -------------------------------------------------- | ---------------------------------------- |
| `[project-brief.md](./project-brief.md)`           | Single source of truth for product scope |
| `[roadmap.md](./roadmap.md)`                       | 10-sprint plan, dependencies, milestones |
| `[architecture.md](./architecture.md)`             | Modules, data flow, boundaries           |
| `[tech-stack.md](./tech-stack.md)`                 | Closed technology decisions              |
| `[definition-of-done.md](./definition-of-done.md)` | Task / sprint / CI DoD                   |
| `[i18n-localization.md](./i18n-localization.md)`   | Binding i18n rules                       |
| `[ideal-team.md](./ideal-team.md)`                 | Roles and sprint ownership               |


Agent entry: `[../AGENTS.md](../AGENTS.md)`

---

## Milestones (summary)

Full definitions: `[roadmap.md](./roadmap.md)` §3.


| Milestone | Sprints | Outcome | Status |
| --------- | ------- | ------- | ------ |
| **M1** | 01 | Runnable skeleton via `docker run` | **Done** |
| **M2** | 02–03 | Login + scanned library with metadata | **Done** |
| **M3** | 04–05 | Browser playback incl. transcode — first alpha | **Partial** (see current-state) |
| **M4** | 06–07 | Zero-config setup + HW acceleration — beta candidate | **Partial** (HW broken) |
| **M5** | 08–09 | Requests + automation + indexers | **On hold** |
| **M5.5** | 11–13 | Core experience (Netflix-like UI) + performance | **Active** |
| **M6** | 10 | OSS release — 1.0 | Paused (after M5.5) |


---

## Sprint task folders


| Sprint | Folder                                                                       | Main output                                    |
| ------ | ---------------------------------------------------------------------------- | ---------------------------------------------- |
| 01     | `[sprint-01-foundation/](./sprint-01-foundation/)`                           | Skeleton + Docker + API contract + make gates  |
| 02     | `[sprint-02-library-metadata/](./sprint-02-library-metadata/)`               | Scanning + metadata + file watching            |
| 03     | `[sprint-03-auth-users/](./sprint-03-auth-users/)`                           | Multi-user, RBAC, profiles, watch state        |
| 04     | `[sprint-04-streaming-transcode/](./sprint-04-streaming-transcode/)`         | Direct play, transcode, HLS/DASH               |
| 05     | `[sprint-05-web-ui/](./sprint-05-web-ui/)`                                   | Library UI, player, search                     |
| 06     | `[sprint-06-onboarding-optimization/](./sprint-06-onboarding-optimization/)` | Setup wizard, smart defaults, subtitles        |
| 07     | `[sprint-07-hardware-acceleration/](./sprint-07-hardware-acceleration/)`     | QSV/NVENC/VAAPI/AMF                            |
| 08     | `[sprint-08-request-management/](./sprint-08-request-management/)`           | Request/approval + notifications               |
| 09     | `[sprint-09-automation-indexers/](./sprint-09-automation-indexers/)`         | Plugin architecture + indexers + downloads     |
| 10     | `[sprint-10-polish-oss-release/](./sprint-10-polish-oss-release/)`           | Performance, docs, security audit, OSS release |
| 11     | `[sprint-11-playback-foundation/](./sprint-11-playback-foundation/)`         | Playback in all browsers + data correctness    |
| 12     | `[sprint-12-streaming-experience/](./sprint-12-streaming-experience/)`       | Netflix-like UI + image pipeline               |
| 13     | `[sprint-13-performance-scale/](./sprint-13-performance-scale/)`             | Performance budgets at 10k-item scale          |


---

## Execution workflow

```
.plan/ task spec  →  GitHub Issue (status, assignee)  →  PR  →  make check (local, no CI)
notes/ briefing   →  decision finalized  →  update .plan/ or docs/
```

- Open issues with the **[Sprint task](../.github/ISSUE_TEMPLATE/sprint_task.md)** template.
- Every issue must reference a `.plan/` task file and acceptance criteria.
- PR descriptions must reference the issue and DoD checklist items.

---

## Operational docs (`docs/`)


| Document                                                      | Purpose                      |
| ------------------------------------------------------------- | ---------------------------- |
| `[docs/testing-strategy.md](../docs/testing-strategy.md)`     | Test layers and expectations |
| `[docs/issue-severity.md](../docs/issue-severity.md)`         | Bug severity definitions     |
| `[docs/sprint-NN-dod-checklist.md](../docs/)`                 | Per-sprint DoD verification  |
| `[docs/m3-alpha-checklist.md](../docs/m3-alpha-checklist.md)` | Alpha readiness              |
| `[docs/m4-beta-checklist.md](../docs/m4-beta-checklist.md)`   | Beta readiness               |


