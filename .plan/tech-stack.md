# Somra — Tech Stack

> Approved technology decisions and rationale. Adding new dependencies requires Tech Lead approval
> and updating this document (anti-drift).

Related: [`architecture.md`](./architecture.md) · [`project-brief.md`](./project-brief.md) · [`definition-of-done.md`](./definition-of-done.md)

---

## 1. Backend

| Topic | Choice | Rationale |
|---|---|---|
| Language | **Go (current stable)** | Single binary, low memory, high concurrency, easy deployment. |
| HTTP router | **`go-chi/chi`** | `net/http` compatible, lightweight, idiomatic middleware; minimal dependencies. |
| Database | **SQLite (WAL)** | Embedded, zero config, single file. |
| SQLite driver | **`modernc.org/sqlite` (pure Go, no CGO)** | No CGO → easy cross-compile and amd64/arm64 multi-arch builds. |
| Migrations | **`pressly/goose`** (embedded `embed.FS` migrations) | Versioned schema, Go + SQL migrations, auto-apply on startup. |
| Media analysis | **ffprobe** | Technical metadata extraction. |
| Transcode | **ffmpeg** | Industry standard; packaged in image. |
| Job scheduling | **Our lightweight scheduler + `robfig/cron/v3`** | Cron library for expressions; state/concurrency control in our layer. |
| Auth/session | **Short-lived JWT access token + revocable server-side refresh token (DB)** | Stateless API + token revocation balance. |
| API contract | **OpenAPI 3.1, design-first (hand-authored spec)** | Single source of truth; frontend TypeScript types generated from it. |
| i18n (backend) | **`nicksnyder/go-i18n/v2` + `golang.org/x/text`** | Locale-aware message catalog, plurals, language matching. See [`i18n-localization.md`](./i18n-localization.md). |
| Tests | Go `testing` + `testify` | Unit + integration. |

## 2. Frontend

| Topic | Choice | Rationale |
|---|---|---|
| Framework | **React + TypeScript** | Broad ecosystem, type safety. |
| Build | **Vite** | Fast dev/HMR, SPA output. |
| State management | **TanStack Query (server state) + Zustand (UI/global state)** | Clear server vs UI state separation, lightweight. |
| Video player | **hls.js** (+ native HLS on Safari) | Adaptive streaming in browser. |
| Packaging format | **CMAF (fMP4)** | HLS from single segment set (primary); DASH manifest from same segments optional. |
| Styling/design system | **Tailwind CSS + Radix UI primitives** | Build our design system on top; accessible, modern UI. |
| Theme system | **Dynamic, user-selectable multi-theme** (token-based) | Original theme set: **Cinematic (default)**, Aurora, Noir, Minimal (no brand mimicry). Theme remembered per user. Sprint 01/03/05. |
| i18n | **`i18next` + `react-i18next`** | Source en-US, translation tr-TR; date/number l10n via `Intl`. See [`i18n-localization.md`](./i18n-localization.md). |

## 3. Data & Deployment

| Topic | Choice |
|---|---|
| Primary data | SQLite |
| File system | User volumes (media), cache/transcode directory |
| Packaging | Single **Docker** image + `docker compose` example |
| Architectures | amd64 + arm64 (multi-arch build) |
| Hardware access | GPU passthrough (QSV/NVENC/VAAPI/AMF) — Sprint 07 |

## 4. CI/CD & Quality

| Topic | Choice |
|---|---|
| CI | Git-based pipeline (lint + test + build + image) |
| Lint | Go: `golangci-lint`; Frontend: ESLint + Prettier |
| Versioning | Semantic versioning (SemVer) + incremental milestone each sprint |
| Image publishing | **GitHub Container Registry (GHCR)** (primary) + optional Docker Hub mirror |
| Contribution agreement | **DCO (Developer Certificate of Origin)** — no CLA |

## 5. External Services / Providers

- **Metadata:** TMDB, TVDB, MusicBrainz, fanart.tv, OMDB (key/rate limit management Sprint 02).
- **Subtitles:** Open subtitle providers (Sprint 06).
- **Notifications:** Webhook, Discord, email (Sprint 08).

## 6. Dependency Policy

1. Minimum dependency principle: standard library first.
2. Every new dependency: license compatibility (must be AGPL-compatible), maintenance status, security review.
3. Dependencies incompatible with AGPL-3.0 **cannot be used** (see [`project-brief.md`](./project-brief.md) §5).

## 7. Closed Decisions (Decided)

> All technology decisions left open at plan start are now closed. Sprint 01 tasks are
> "implement/validate decision" not "decide". Changes only via Tech Lead approval + document update (anti-drift).

| Decision | Outcome |
|---|---|
| HTTP router | `go-chi/chi` |
| SQLite driver | `modernc.org/sqlite` (pure Go, no CGO) |
| Migration tool | `pressly/goose` (embedded migrations) |
| Scheduler | Our lightweight scheduler + `robfig/cron/v3` |
| Session strategy | JWT (short-lived) + revocable refresh token (DB) |
| API contract | OpenAPI 3.1 design-first → FE type generation |
| Frontend state | TanStack Query + Zustand |
| Video player / packaging | hls.js + CMAF (HLS primary, DASH optional) |
| Styling/design system | Tailwind CSS + Radix UI |
| Frontend i18n | `i18next` + `react-i18next` |
| Backend i18n | `nicksnyder/go-i18n/v2` + `x/text` |
| Translation platform | **Weblate** (self-host, OSS, git-integrated) |
| Image registry | GHCR (primary) |
| License / contribution | AGPL-3.0 + DCO |

Details: session/contract/plugins in [`architecture.md`](./architecture.md) §8; i18n in [`i18n-localization.md`](./i18n-localization.md); license in [`project-brief.md`](./project-brief.md) §5.
