# Somra — Project Brief

> This file is the project's **single source of truth**. All sprint and task files reference
> this document for scope, goals, and decisions. If a task conflicts with the scope here,
> the task is wrong — update this document instead (with a change log entry).

Related documents: [`ideal-team.md`](./ideal-team.md) · [`architecture.md`](./architecture.md) · [`tech-stack.md`](./tech-stack.md) · [`roadmap.md`](./roadmap.md) · [`definition-of-done.md`](./definition-of-done.md) · [`i18n-localization.md`](./i18n-localization.md)

---

## 1. Vision

**Somra** is an integrated platform that delivers everything a home media server needs under
**one roof**, with **a single Docker install**, following a **minimum configuration / maximum
optimization** philosophy.

Today, a user must install, integrate, and configure dozens of separate tools — Jellyfin/Emby/Plex
(media server + transcode), Sonarr/Radarr/Lidarr (content management), Prowlarr (indexers),
Bazarr (subtitles), Overseerr/Jellyseerr (request management), and download clients. Somra
unifies this fragmented experience into **one product**.

## 2. Problem

- Installing, updating, and integrating many separate tools requires expertise.
- Each tool has its own settings, identity, and database layer — causing inconsistency and maintenance burden.
- Optimization (transcode profiles, hardware acceleration, quality profiles) requires experience.
- The barrier to entry for new users is high.

## 3. Solution

A single-binary/service core written in Go plus a unified React interface. The user installs via
Docker, goes through a setup wizard, and the system starts with **smart defaults**. Power users
can fine-tune everything from the UI.

## 4. Consolidated Decisions

| Topic | Decision | Detail |
|---|---|---|
| Core strategy | **Build our own engine from scratch** | Scanning, metadata, transcode, and streaming are our code. See [`architecture.md`](./architecture.md). |
| Backend | **Go** | Single binary, low resource usage, high concurrency. |
| Frontend | **React SPA (Vite)** | Web-first interface. |
| Database | **Embedded SQLite** | Zero config, single file. |
| Transcode | **Software first (ffmpeg CPU)** | Hardware acceleration in Sprint 07. |
| Platform | **Web-first** | Mobile/TV on future roadmap; out of scope for this plan. |
| Identity | **Multi-user + RBAC + parental controls** | See Sprint 03. |
| i18n | **Cross-cutting requirement** | Source locale **en-US**, translation **tr-TR**. Full coverage (UI, backend messages, notifications, metadata language, l10n, docs). See [`i18n-localization.md`](./i18n-localization.md). |
| Content acquisition (*arr/indexer/torrent/usenet) | **Excluded in first phase; full in later sprint** | Full *arr/Prowlarr automation via plugin architecture in Sprint 09. |
| Team | **Ideal team assumption** | See [`ideal-team.md`](./ideal-team.md). |
| Timeline | **No hard deadline** | Default sprint cadence is 2 weeks; sprints sized by work packages. |
| Brand | **somra** | Visual identity in Sprint 05 design tasks. |
| License | **AGPL-3.0 + DCO (decided)** | See below. |

## 5. License: AGPL-3.0 + DCO (Decided)

**Decision: GNU Affero General Public License v3.0 (AGPL-3.0)** with **DCO (Developer Certificate
of Origin)** as the contribution agreement. No CLA.

Rationale:
- Somra is **server/network service** software. Permissive licenses like MIT/Apache allow third
  parties to take the code, run it as a **closed SaaS**, and not share changes back.
- AGPL-3.0 **requires source sharing** even for network-delivered modifications. This protects
  community contributions and keeps the project open (a stronger version of Jellyfin's GPL philosophy).
- **DCO** secures contribution provenance without CLA bureaucracy (every commit has `Signed-off-by`).
  Low friction for the open source community.

Alternatives considered and rejected:
- **Apache-2.0 / MIT**: Rejected because they allow SaaS exploitation.
- **GPL-3.0**: Rejected in favor of the network clause (AGPL).
- **CLA**: Rejected in favor of DCO due to contribution friction.

> **Implementation:** `LICENSE` (AGPL-3.0) and `DCO`/`CONTRIBUTING` added in Sprint 01;
> formal approval recorded in Sprint 01 architecture tasks (implementation, not re-debate).

## 6. Scope (In This Plan)

This 10-sprint plan covers the following capabilities:

1. Library scanning, technical + enriched metadata, file watching.
2. Multi-user, RBAC, profiles, parental controls, watch state/continue watching.
3. Direct play + software transcode, HLS/DASH, adaptive bitrate, subtitles.
4. Web UI: library browsing, detail pages, web player, search; **user-selectable dynamic themes**
   (original theme set: Cinematic default, Aurora, Noir, Minimal — remembered per user, no brand mimicry).
5. Setup wizard + smart defaults (minimum config / maximum optimization).
6. Hardware acceleration (QSV/NVENC/VAAPI/AMF).
7. Request management (Overseerr-like) + notifications.
8. Download automation + indexer (torrent + usenet) + plugin architecture (*arr/Prowlarr-like).
9. Open source release readiness.
10. **i18n/l10n:** en-US (source) + tr-TR; development is i18n-compliant from day one. See [`i18n-localization.md`](./i18n-localization.md).

## 7. Out of Scope (Outside This Plan — Drift Prevention)

The following are **deliberately outside this plan**. If a task touches any of these, this
document must be updated first and a new sprint/effort planned:

- Native mobile and TV apps (iOS/Android/Android TV/tvOS/webOS/Tizen).
- Live TV / DVR / EPG.
- Multi-server federation / cloud sync / external CDN.
- Advanced music features (full Lidarr parity) — beyond basic music library.
- Full DLNA/Chromecast compatibility (web-first only).

## 8. Success Criteria (High Level)

- **Setup:** A user with zero prior knowledge reaches a running server in under 10 minutes with
  a single `docker run`/`docker compose up`.
- **Optimization:** Automatic transcode profile based on hardware/media detection; smooth playback
  without manual tuning.
- **Integration:** Core functions of the separate tools above are usable from one UI, one identity,
  and one data layer.
- **Performance:** Runs reasonably on home server hardware (e.g. 4 cores / 8GB RAM).
- **Open source readiness:** License, contribution guide, security policy, CI/CD, and documentation ready.

## 9. Governance Rules (Anti-Drift / Framing)

> These rules are **binding** for all sprint and task files.

1. **Scope lock:** No sprint may enter out-of-scope items from [Section 7](#7-out-of-scope-outside-this-plan--drift-prevention) without updating this document.
2. **DoD required:** No task is "done" until it meets [`definition-of-done.md`](./definition-of-done.md) criteria.
3. **Dependency discipline:** Each sprint builds on completed outputs from prior sprints; dependencies are stated explicitly in task files.
4. **Single source of truth:** Decisions are consolidated here; this document wins on conflict.
5. **Version tagging:** Each sprint ends with a working incremental release (milestone). See [`roadmap.md`](./roadmap.md).
6. **i18n requirement:** No user-facing text may be embedded in code; every feature ships with en-US + tr-TR. See [`i18n-localization.md`](./i18n-localization.md) and [`definition-of-done.md`](./definition-of-done.md).

## 10. Change Log

| Date | Change | Owner |
|---|---|---|
| (creation) | Initial brief created | — |
