# AGENTS.md

Guidance for AI coding agents working on **Somra** — an all-in-one, self-hosted home media
server (library + metadata + transcoding/streaming + requests + automation) delivered as a
single Docker install, built from scratch.

> This file is the agent-facing companion to the human docs. The **authoritative product and
> scope source of truth lives in [`.plan/`](./.plan/)**. When in doubt about *what*
> to build or whether something is in scope, read `.plan/project-brief.md` first. This file
> tells you *how* to work in the repo.
>
> Language note: this file, planning docs under `.plan/`, and all source code / code comments /
> commit messages are in **English** (the project's source locale is `en-US`). User-facing
> strings are localized (`en-US` + `tr-TR`); see the i18n rules below.

---

## Project overview

- **Goal:** Replace the fragmented stack (Jellyfin/Plex + Sonarr/Radarr + Prowlarr + Bazarr +
  Overseerr + download clients) with one unified product. Philosophy: **minimum
  configuration, maximum optimization**.
- **Status:** Sprint 01 (M1) complete. **Active sprint:** 02 — Library & Metadata.
  See [`.plan/00-index.md`](./.plan/00-index.md) for the current dashboard.
- **License:** **AGPL-3.0** with **DCO** sign-off (no CLA). Any dependency must be
  AGPL-3.0-compatible.

Key references in [`.plan/`](./.plan/):

- [`.plan/00-index.md`](./.plan/00-index.md) — **start here** — active sprint, doc hierarchy, links.
- [`.plan/project-brief.md`](./.plan/project-brief.md) — vision, scope (in/out), decisions, governance.
- [`.plan/architecture.md`](./.plan/architecture.md) — modules, data flow, decided architecture.
- [`.plan/tech-stack.md`](./.plan/tech-stack.md) — closed technology decisions (§7).
- [`.plan/definition-of-done.md`](./.plan/definition-of-done.md) — DoD, coding standards, test/coverage, CI gates.
- [`.plan/i18n-localization.md`](./.plan/i18n-localization.md) — binding i18n rules.
- [`.plan/roadmap.md`](./.plan/roadmap.md) — sprint/milestone plan.

## Documentation hierarchy

On conflict, higher rows win:

| Priority | Location | Role |
|---|---|---|
| 1 | `.plan/project-brief.md` | Scope and governance |
| 2 | `.plan/architecture.md`, `.plan/tech-stack.md` | Module boundaries, closed tech |
| 3 | `.plan/definition-of-done.md`, `.plan/i18n-localization.md` | DoD, i18n |
| 4 | `.plan/sprint-XX-*/` task files | Acceptance criteria |
| 5 | `docs/` | Operational docs (testing, checklists) |
| 6 | `api/openapi.yaml` | HTTP API contract |
| 7 | `notes/` | **Non-binding** — briefings, draft ADRs (context only) |

**Execution status** (todo / in progress / done) lives in **GitHub Issues**, not in `notes/`.
Open issues with [`.github/ISSUE_TEMPLATE/sprint_task.md`](./.github/ISSUE_TEMPLATE/sprint_task.md);
every issue must reference a `.plan/` task file.

Do **not** use external knowledge bases (e.g. NotebookLM) as a source of truth — read the repo.

## Tech stack (decided — do not swap without Tech Lead approval + doc update)

**Backend (Go)**
- HTTP router: `go-chi/chi`
- DB: **SQLite (WAL)** via `modernc.org/sqlite` (pure Go, **no CGO** — keep builds CGO-free)
- Migrations: `pressly/goose` (embedded via `embed.FS`, applied on startup)
- Scheduler: in-house lightweight scheduler + `robfig/cron/v3`
- Auth: short-lived **JWT access token** + revocable server-side **refresh token** (DB)
- API: **OpenAPI 3.1, design-first** (hand-authored spec is the source of truth; FE types generated from it)
- i18n: `nicksnyder/go-i18n/v2` + `golang.org/x/text`
- Media: `ffprobe` (analysis) + `ffmpeg` (transcode), packaged in the image
- Tests: Go `testing` + `testify`

**Frontend (React + TypeScript)**
- Build: **Vite** (SPA), TypeScript **strict**
- State: **TanStack Query** (server state) + **Zustand** (UI state)
- Styling: **Tailwind CSS** + **Radix UI** primitives (own design system on top)
- Video: **hls.js** (+ native HLS on Safari); delivery via **CMAF** (HLS primary, DASH optional)
- i18n: `i18next` + `react-i18next` (+ `Intl` for date/number)
- Themes: dynamic, user-selectable — `cinematic` (default), `aurora`, `noir`, `minimal`

**Platform**
- Single **Docker** image + `docker compose`; multi-arch **amd64 + arm64**
- Image registry: **GHCR** (primary)

## Repository layout (target)

```
.
├── AGENTS.md            # this file
├── .plan/               # authoritative planning docs (start: 00-index.md)
├── notes/               # Obsidian vault — non-binding working notes
├── docs/                # operational docs (testing, checklists, design)
├── cmd/                 # Go entrypoint(s)
├── internal/            # Go modules (api, auth, library, metadata, streaming, settings, jobs, ...)
├── web/                 # React + Vite SPA (frontend)
├── migrations/          # goose migrations (embedded)
├── api/                 # OpenAPI 3.1 spec (source of truth for the API contract)
├── deploy/              # Dockerfile, docker-compose examples
└── Makefile             # standard dev tasks
```

Module boundaries must match [`.plan/architecture.md`](./.plan/architecture.md) §3. In this
monorepo you may add **nested `AGENTS.md`** files (e.g. `web/AGENTS.md`) for subproject-specific
instructions — the closest file to an edited file wins.

## Dev environment & commands

Prefer the `Makefile` targets (created in Sprint 01). Intended commands:

```bash
make dev          # run backend (hot reload) + frontend dev server
make build        # build Go binary + frontend static assets
make test         # run all tests (unit + integration)
make lint         # golangci-lint + ESLint/Prettier
make i18n-check   # missing/unused keys + en-US/tr-TR completeness
make coverage     # run tests with coverage + enforce thresholds
make migrate      # apply DB migrations
make docker       # build the single multi-arch image
```

Backend: Go (current stable). Frontend: Node LTS + `pnpm` (under `web/`). Keep builds CGO-free.

## Code style

- **Go:** `gofmt`/`goimports`, `golangci-lint` clean. Meaningful package boundaries, wrap
  errors with `%w`, pass `context.Context`, avoid global state, document exported APIs.
- **TypeScript/React:** strict mode, typed API layer (generated from OpenAPI), component/hook
  separation, accessibility (WCAG), ESLint + Prettier clean.
- **Comments:** explain intent/trade-offs/constraints only — no comments that merely narrate
  what the code does. Do not describe the change you are making in a comment.
- **No hardcoded user-facing strings** (see i18n).

## Testing instructions

Per [`.plan/definition-of-done.md`](./.plan/definition-of-done.md) §4 / §4.1:

- Unit tests are **mandatory** for business logic and pure functions.
- Integration tests for boundaries: DB, file scanning, transcode pipeline.
- E2E tests for critical user flows: login, browsing, playback, setup wizard.
- **Coverage thresholds (enforced by CI, blocking):**
  - Core Go business logic: **≥ 80%** line coverage
  - Critical modules (auth/RBAC, scanning, transcode-decision engine, automation import): **≥ 90%**
  - Frontend component tests: **≥ 70%**; critical flows require **e2e**
- Run `make test` and `make coverage`; fix all failures and type errors before finishing.
- Add/update tests for code you change, even if not explicitly asked.

## CI gates (must be green to merge)

```
lint → i18n-check → unit-test → integration-test → coverage-gate → build → image-build
```

If any gate is red, **do not merge**. `i18n-check` fails on missing/incomplete `en-US`/`tr-TR`
keys; `coverage-gate` fails below the thresholds above.

## Internationalization (binding)

See [`.plan/i18n-localization.md`](./.plan/i18n-localization.md).

- Source/fallback locale: **`en-US`**; first translation: **`tr-TR`**.
- **No user-facing hardcoded text** — always resolve via keys (`domain.context.key`).
- Every feature ships its `en-US` **and** `tr-TR` keys together; otherwise it is not "Done".
- Use `Intl`/locale-aware formatting for dates/numbers (no manual string concatenation).
- Locale negotiation order: user profile → system default → `Accept-Language` → `en-US`.

## Security considerations

See [`.plan/definition-of-done.md`](./.plan/definition-of-done.md) §6 and
[`.plan/sprint-03-auth-users/04-security-tasks.md`](./.plan/sprint-03-auth-users/04-security-tasks.md).

- Least privilege; validate **all** input; parameterized queries (no string-built SQL).
- Never hardcode secrets; provider API keys are user-supplied and stored securely.
- Enforce authorization on every protected endpoint (RBAC); strong password hashing;
  brute-force protection (rate limit/lockout); revocable tokens.
- Guard against path traversal (media file access) and SSRF (outbound provider/indexer calls).
- Content acquisition (indexers/torrent/usenet) is an **isolated plugin** concern — the core
  must work fully without plugins. Do not couple acquisition into the core.

## Commit & PR guidelines

- Small, focused commits/PRs; reference the related task and acceptance criteria.
- **Sign off every commit (DCO):** `git commit -s` (adds `Signed-off-by`).
- PRs must pass all CI gates (above) and meet the DoD before merge; architecture-affecting
  changes require Tech Lead approval.
- Do not commit secrets or `.env`-style credential files.

## Scope & governance (anti-drift)

- **Do not expand scope** beyond [`.plan/project-brief.md`](./.plan/project-brief.md) §6; items in
  §7 (e.g. native mobile/TV apps, live TV/DVR, server federation) are **out of scope** unless
  the brief is updated first.
- Technology/architecture decisions are **closed** (see `tech-stack.md` §7, `architecture.md`
  §8). Changing one requires Tech Lead approval **and** updating the relevant doc.
- "Nice to have" extras go to the backlog, not silently into the current sprint.

## Good first pointers for an agent

1. Read [`.plan/00-index.md`](./.plan/00-index.md) then the active sprint folder under [`.plan/`](./.plan/).
2. Confirm the task is in scope ([`.plan/project-brief.md`](./.plan/project-brief.md)) and identify the owning module ([`.plan/architecture.md`](./.plan/architecture.md) §3).
3. If a GitHub Issue exists, use it for status; otherwise reference the `.plan/` task file in the PR.
4. Implement with tests + i18n keys; run `make lint test i18n-check coverage`.
5. Open a focused, DCO-signed PR referencing the issue / task and acceptance criteria.

## Cursor Cloud specific instructions

Notes for cloud agents (toolchain is pre-installed in the VM snapshot):

- **Sprint 01 is complete** — `go.mod`, `web/package.json`, `Makefile`, CI, and the M1 skeleton exist.
  Run `make lint test i18n-check coverage build` before finishing substantive changes.
- **Toolchain:** Go (current stable, `/usr/local/go`), Node 22 + `pnpm` 10, `ffmpeg`/`ffprobe` 6.x,
  `golangci-lint` 2.x, GNU `make`.
- **Docker is NOT installed** in the default cloud snapshot. It is only needed for `make docker` /
  image builds — not for `make dev`. Install on demand for image/CI work.
- **Keep builds CGO-free** (`CGO_ENABLED=0`); the SQLite driver is `modernc.org/sqlite` (pure Go).
- The startup update script runs `go mod download` and `pnpm install` when manifests exist.
