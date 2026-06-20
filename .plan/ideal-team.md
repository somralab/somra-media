# Somra — Ideal Team Structure

> This document defines the **realistic ideal team** for the "build our own engine from scratch +
> full feature parity" goal. Per [`project-brief.md`](./project-brief.md), the plan assumes an
> ideal team; the team will be scaled accordingly.

Related: [`roadmap.md`](./roadmap.md) (people/work distribution) · [`definition-of-done.md`](./definition-of-done.md)

---

## 1. Summary

| Scale | Total | Note |
|---|---|---|
| **Ideal team** | **~8–9 people** | Full-speed parallel development. This plan's assumption. |
| **Minimum core** | **3–4 people** | Possible but timeline ~2–3x longer; less parallel work. |

Disciplines are the same at both scales; the difference is parallel capacity and timeline.

## 2. Roles and Responsibilities

### 2.1 Tech Lead / Software Architect (Go) — 1 person
- System architecture and module boundaries (see [`architecture.md`](./architecture.md)).
- API contracts, code standards, owner of [`definition-of-done.md`](./definition-of-done.md).
- Technical risk management, library/technology decisions, final code review authority.
- Sprint planning and technical backlog prioritization (with PM).
- **Output owner:** Sprint 01 architecture tasks.

### 2.2 Backend Engineer (Go) — 3 people
- Library scanning, metadata pipeline, user/RBAC, request management, automation, indexer.
- Data access layer, business rules, background jobs (job scheduler).
- Unit and integration tests.
- **Output owner:** Most of Sprint 02, 03, 06, 08, 09 backend tasks.

### 2.3 Media / Streaming Specialist (ffmpeg, codecs) — 1 person
- Transcode pipeline, HLS/DASH packaging, adaptive bitrate, subtitle/audio track handling.
- Hardware acceleration (QSV/NVENC/VAAPI/AMF) detection and selection.
- Playback compatibility matrix and quality profiles.
- **Output owner:** Sprint 04, 07 media tasks.

### 2.4 Frontend Engineer (React) — 2 people
- React SPA (Vite), design system implementation, state management, API integration.
- Web video player (hls.js/dash.js), library browsing, search, setup wizard.
- Accessibility and performance (lazy load, virtual lists).
- **Output owner:** Sprint 05 and frontend tasks in other sprints.

### 2.5 DevOps / Platform Engineer — 1 person
- Single Docker image + `docker compose` deployment, multi-arch (amd64/arm64) build.
- CI/CD pipeline, release automation, image publishing (registry).
- Hardware device access (GPU passthrough), observability (log/metrics).
- **Output owner:** Sprint 01, 07, 10 devops tasks.

### 2.6 QA / Test Automation Engineer — 1 person
- Test strategy, e2e automation, regression suite, acceptance testing.
- DoD verification and bug tracking each sprint.
- **Output owner:** QA tasks in every sprint.

### 2.7 UX/UI Designer — 0.5 person (part-time)
- Design system, flows, setup wizard UX, visual identity (somra brand).
- **Output owner:** Sprint 05 design tasks.

### 2.8 Product Manager / PM — 0.5 person (part-time)
- Backlog, sprint management, scope protection (anti-drift), release planning.
- Enforcement of [`project-brief.md`](./project-brief.md) governance rules.

## 3. Headcount Table

| Role | Ideal | Minimum core |
|---|---|---|
| Tech Lead / Architect (Go) | 1 | 1 |
| Backend (Go) | 3 | 1 |
| Media/Streaming | 1 | (shared with Tech Lead/Backend) |
| Frontend (React) | 2 | 1 |
| DevOps/Platform | 1 | (shared with Tech Lead) |
| QA | 1 | (shared among developers) |
| UX/UI | 0.5 | (external) |
| PM | 0.5 | (shared with Tech Lead) |
| **Total** | **~9 people** | **3–4 people** |

## 4. Skill Matrix

| Skill | Primary role | Secondary |
|---|---|---|
| Go service architecture | Tech Lead | Backend |
| SQLite / data modeling | Backend | Tech Lead |
| ffmpeg / codec / transcode | Media Specialist | Tech Lead |
| Hardware acceleration (GPU) | Media Specialist | DevOps |
| React / TypeScript | Frontend | — |
| Video player (HLS/DASH) | Frontend | Media Specialist |
| Docker / CI-CD | DevOps | Tech Lead |
| Test automation | QA | all developers |
| Security / RBAC | Backend | Tech Lead |
| UX / design | UX/UI | Frontend |

## 5. Recommended Hiring/Onboarding Order

1. **Tech Lead (Go)** — architecture and standards foundation (before Sprint 01).
2. **Backend #1 + DevOps** — skeleton, CI/CD, data layer.
3. **Frontend #1** — design system and API integration foundation.
4. **Media/Streaming Specialist** — before Sprint 04.
5. **Backend #2–#3, Frontend #2, QA** — as parallel capacity grows.
6. **UX/UI and PM** — part-time from Sprint 01.

## 6. Working Cadence

- **Sprint cadence:** 2 weeks (default). No hard deadline; scope protection is essential.
- **Ceremonies:** Sprint planning, daily standup, sprint demo (working incremental release), retrospective.
- **Code review:** Every PR needs at least 1 approval; Tech Lead approval required for architectural changes.
- **Single source of truth:** [`project-brief.md`](./project-brief.md) governs scope and decision disputes.
