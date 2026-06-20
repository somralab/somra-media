# Somra

> All-in-one, self-hosted home media server — library + metadata +
> transcoding/streaming + requests + automation, delivered as a single
> Docker install. Philosophy: **minimum configuration, maximum
> optimization**. See
> [`.plan/project-brief.md`](./.plan/project-brief.md) for the full vision
> and scope.

**Status:** Sprints **01–08 complete**. **Active sprint:** **09 — Automation &
Indexers** (plugin architecture, indexers, download automation). See
[`.plan/00-index.md`](./.plan/00-index.md) for the planning dashboard and
Sprint 09 task order.

**License:** [AGPL-3.0-or-later](./LICENSE) with
[Developer Certificate of Origin](https://developercertificate.org)
sign-off on every commit (no CLA). See
[`CONTRIBUTING.md`](./CONTRIBUTING.md).

---

## Milestones

Summary only — full sprint plan and dependencies:
[`.plan/roadmap.md`](./.plan/roadmap.md).

| Milestone | Sprints | Outcome | Status |
| --------- | ------- | ------- | ------ |
| **M1** | 01 | Runnable skeleton via `docker run` | Done |
| **M2** | 02–03 | Login + scanned library with metadata | Done |
| **M3** | 04–05 | Browser playback incl. transcode — first alpha | Done |
| **M4** | 06–07 | Zero-config setup + HW acceleration — beta | Done |
| **M5** | 08–09 | Requests + automation + indexers | In progress |
| **M6** | 10 | Open source release — 1.0 | Planned |

Track execution via [GitHub Issues](https://github.com/somralab/somra-media/issues)
(sprint task template). Working notes live in [`notes/`](./notes/) (Obsidian vault).

---

## Tech stack (snapshot)

Decided technologies — do not swap without Tech Lead approval and a doc
update (see [`.plan/tech-stack.md`](./.plan/tech-stack.md)).

- **Backend (Go):** `go-chi/chi`, SQLite (WAL) via `modernc.org/sqlite`
  (pure Go, **no CGO**), `pressly/goose` migrations,
  `robfig/cron/v3`, JWT access + revocable refresh tokens,
  `nicksnyder/go-i18n/v2`.
- **Frontend (React + TypeScript):** Vite, strict TS, TanStack Query +
  Zustand, Tailwind CSS + Radix UI, `i18next` / `react-i18next`,
  `hls.js` for playback (HLS / CMAF).
- **Media:** `ffprobe` (analysis) + `ffmpeg` (transcode), packaged in
  the image.
- **API:** OpenAPI 3.1, **design-first** — [`api/openapi.yaml`](./api/openapi.yaml)
  is the source of truth, frontend types are generated from it.
- **Platform:** single Docker image + `docker compose`, multi-arch
  (amd64 + arm64), published to GHCR.

---

## Quick start

```bash
make help            # list all targets
make dev             # backend + frontend dev servers concurrently
make build           # backend (CGO=0) + frontend bundle
make test            # Go unit tests + Vitest
make lint            # gofmt + golangci-lint + ESLint + Prettier + tsc
make i18n-check      # verify en-US / tr-TR key parity (Go + frontend)
make coverage        # run all tests with coverage outputs
make coverage-gate   # enforce DoD §4.1 thresholds (blocks PR merge in CI)
make docker          # build the container image for the local arch
make docker-multiarch # buildx + push amd64 + arm64 to GHCR
```

Running with Docker (builds locally — no GHCR pull required):

```bash
mkdir -p deploy/config deploy/data deploy/cache deploy/media
docker compose -f deploy/docker-compose.yml up --build
# → API + SPA on http://localhost:8080 (API under /api/v1)
```

The `ghcr.io/somralab/somra-media` image tag in compose is for local naming only.
CI and tagged releases publish to GHCR; for a standalone image build use `make docker`.

### Sprint 01 (M1) demo

The M1 milestone ships a fully-functional skeleton that you can reproduce
locally in three steps. Both recipes (docker and `go run`) hit the same
HTTP gateway exposed by the single `somra` binary.

**Container (preferred — matches CI image recipe):**

```bash
mkdir -p deploy/config deploy/data deploy/cache deploy/media
docker compose -f deploy/docker-compose.yml up --build
# alternative: make docker && docker compose -f deploy/docker-compose.yml up
```

Then, in another shell:

```bash
curl -s http://localhost:8080/api/v1/health   | jq    # → {"status":"ok",...,"checks":{"database":{...}}}
curl -s http://localhost:8080/api/v1/version  | jq    # → version / commit / builtAt
curl -sN http://localhost:8080/api/v1/events/stream | head -3   # → event: hello + heartbeat
open http://localhost:8080                            # SPA status dashboard
```

**Without Docker (Go + Node toolchains on host):**

```bash
make build                                            # builds bin/somra + web/dist
SOMRA_WEB_DIR=web/dist ./bin/somra                    # serves API + SPA on :8080
```

Verification:

- `/api/v1/health` is `ok` with a non-empty `checks.database` entry.
- `/api/v1/version` returns the linker-injected version.
- The status page at `/` renders the version card, the localized health
  label (en-US / tr-TR via the dropdown) and at least one SSE event in
  the live panel.
- `Accept-Language: tr-TR` localizes error envelopes:
  `curl -H 'Accept-Language: tr-TR' http://localhost:8080/api/v1/nope`
  returns `"message":"İstenen kaynak bulunamadı."`.

### Prerequisites

- Go (current stable) — backend build & test (project uses Go 1.26).
- Node.js LTS (≥20) + `pnpm` 10 — frontend build & test (under `web/`).
- `ffmpeg` / `ffprobe` available at runtime (the production image ships
  them; for local backend work install them via your OS package
  manager).
- Optional: `golangci-lint` v2, Docker + buildx (for `make docker`).

---

## CI gates

Every PR runs the following jobs in order — each must be green to merge
(see [`.plan/definition-of-done.md`](./.plan/definition-of-done.md) §5):

`lint → i18n-check → unit-test → integration-test → coverage-gate → build → e2e → image-build`

The `e2e` job runs the Playwright status-page smoke (see
[`web/e2e/status.spec.ts`](./web/e2e/status.spec.ts)) against the Go
binary serving the built SPA. Run it locally with `make e2e`.

Coverage thresholds enforced by `scripts/coverage-gate.sh` (DoD §4.1):

| Layer                                                           | Minimum |
| --------------------------------------------------------------- | ------- |
| Core Go business logic                                          | ≥ 80%   |
| Critical Go modules (auth, jobs, db, errors, i18n, diagnostics) | ≥ 90%   |
| Frontend statements                                             | ≥ 70%   |

---

## Regenerating frontend types from the OpenAPI spec

The OpenAPI document at [`api/openapi.yaml`](./api/openapi.yaml) is the
single source of truth for the HTTP API. Frontend TypeScript types are
generated, not hand-authored:

```bash
make openapi-types
# equivalent to: bash scripts/gen-openapi-types.sh
```

Output: `web/src/api/generated/openapi.d.ts` (the parent directory is
created on demand, so the script is safe to run before `web/` exists).
The generator uses
[`openapi-typescript`](https://www.npmjs.com/package/openapi-typescript)
via `npx -y`; no global install required.

Re-run this command whenever you change `api/openapi.yaml`.

---

## Internationalization

Binding rules live in
[`.plan/i18n-localization.md`](./.plan/i18n-localization.md). Short
version:

- Source / fallback locale: **`en-US`**. First translation: **`tr-TR`**.
- **Never** hardcode user-facing strings — always resolve via keys in
  the form `domain.context.key`.
- Every feature ships its `en-US` and `tr-TR` keys together. CI fails
  on missing or unused keys (`make i18n-check`).
- Use locale-aware formatting (`Intl`, `go-i18n`) for dates, numbers
  and plurals — no manual string concatenation.

Locale negotiation order: user profile → system default →
`Accept-Language` → `en-US`.

---

## Repository layout (target)

```
.
├── AGENTS.md           # agent-facing how-to (companion to .plan/)
├── .plan/              # authoritative planning docs (start: 00-index.md)
├── notes/              # Obsidian vault — briefings & draft decisions (non-binding)
├── docs/               # operational docs (testing, DoD checklists, design)
├── cmd/                # Go entrypoint(s)
├── internal/           # Go modules
├── web/                # React + Vite SPA
├── migrations/         # goose migrations (embedded)
├── api/                # OpenAPI 3.1 spec (source of truth for HTTP API)
├── deploy/             # Dockerfile, docker compose
├── scripts/            # helper shell scripts
└── Makefile            # standard dev tasks
```

---

## Planning & governance

**Start:** [`.plan/00-index.md`](./.plan/00-index.md) — active sprint, doc hierarchy,
links to all sprint folders.

| Layer | Location | Binding? |
| ----- | -------- | -------- |
| Spec & scope | [`.plan/`](./.plan/) | Yes |
| Execution | GitHub Issues ([templates](./.github/ISSUE_TEMPLATE/)) | Yes (status) |
| Operations | [`docs/`](./docs/) | Yes |
| Working notes | [`notes/`](./notes/) (Obsidian) | No — context only |

Core documents: [`project-brief.md`](./.plan/project-brief.md),
[`roadmap.md`](./.plan/roadmap.md),
[`architecture.md`](./.plan/architecture.md),
[`definition-of-done.md`](./.plan/definition-of-done.md).

Agent workflow: [`AGENTS.md`](./AGENTS.md).

---

## Contributing

See [`CONTRIBUTING.md`](./CONTRIBUTING.md). TL;DR: small focused commits,
Conventional Commits, **DCO sign-off (`git commit -s`)** on every commit
(no CLA), all CI gates green.
