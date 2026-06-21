# Somra

> All-in-one, self-hosted home media server — library + metadata +
> transcoding/streaming + requests + automation, delivered as a single
> Docker install. Philosophy: **minimum configuration, maximum
> optimization**. See
> [`.plan/project-brief.md`](./.plan/project-brief.md) for the full vision
> and scope.

**Status:** Sprints **01–09 complete**. **Active sprint:** **10 — Polish & Open Source**
(M6 / 1.0 release prep). See [`.plan/00-index.md`](./.plan/00-index.md).

**License:** [AGPL-3.0-or-later](./LICENSE) · [Code of Conduct](./CODE_OF_CONDUCT.md) ·
[Contributing](./CONTRIBUTING.md) (DCO sign-off on every commit).

---

## Run Somra (end users)

Pull the published image and start Somra — no build toolchain required:

```bash
mkdir -p ~/somra/{config,data,cache,media}
docker run -d \
  --name somra \
  -p 8080:8080 \
  -e SOMRA_JWT_SECRET="$(openssl rand -base64 32)" \
  -v ~/somra/config:/config \
  -v ~/somra/data:/data \
  -v ~/somra/cache:/cache \
  -v ~/somra/media:/media:ro \
  --restart unless-stopped \
  ghcr.io/somralab/somra-media:1.0.0
```

Open **http://localhost:8080** and complete the setup wizard. Put your media files in
`~/somra/media` (or bind-mount an existing library).

**Documentation**

| English | Turkish |
| ------- | ------- |
| [Installation](./docs/user/en-US/installation.md) | [Kurulum](./docs/user/tr-TR/installation.md) |
| [Getting started](./docs/user/en-US/getting-started.md) | [Başlangıç](./docs/user/tr-TR/getting-started.md) |
| [Library](./docs/user/en-US/library.md) | [Kütüphane](./docs/user/tr-TR/library.md) |
| [Playback](./docs/user/en-US/playback.md) | [Oynatma](./docs/user/tr-TR/playback.md) |
| [Requests & automation](./docs/user/en-US/requests-automation.md) | [İstekler & otomasyon](./docs/user/tr-TR/requests-automation.md) |
| [FAQ](./docs/user/en-US/faq-troubleshooting.md) | [SSS](./docs/user/tr-TR/faq-troubleshooting.md) |

GPU passthrough: [`docs/gpu-setup.md`](./docs/gpu-setup.md). API reference:
[`docs/api/index.html`](./docs/api/index.html) (from [`api/openapi.yaml`](./api/openapi.yaml)).

---

## Milestones

Summary — full plan: [`.plan/roadmap.md`](./.plan/roadmap.md).

| Milestone | Sprints | Outcome | Status |
| --------- | ------- | ------- | ------ |
| **M1** | 01 | Runnable skeleton via `docker run` | Done |
| **M2** | 02–03 | Login + scanned library with metadata | Done |
| **M3** | 04–05 | Browser playback incl. transcode — first alpha | Done |
| **M4** | 06–07 | Zero-config setup + HW acceleration — beta | Done |
| **M5** | 08–09 | Requests + automation + indexers | Done |
| **M6** | 10 | Open source release — **1.0** | In progress |

Track execution via [GitHub Issues](https://github.com/somralab/somra-media/issues).

---

## Develop locally

For contributors building from source:

```bash
make help            # list all targets
make dev             # backend + frontend dev servers concurrently
make build           # backend (CGO=0) + frontend bundle
make test            # Go unit tests + Vitest
make lint            # gofmt + golangci-lint + ESLint + Prettier + tsc
make i18n-check      # verify en-US / tr-TR key parity (Go + frontend)
make coverage        # run all tests with coverage outputs
make docs-api        # generate Redoc HTML at docs/api/index.html
make docker          # build the container image for the local arch
```

Docker Compose (build from repo):

```bash
mkdir -p deploy/config deploy/data deploy/cache deploy/media
docker compose -f deploy/docker-compose.yml up --build
# → http://localhost:8080 (API under /api/v1)
```

Developer docs: [`docs/developer/architecture.md`](./docs/developer/architecture.md),
[`docs/developer/plugin-development.md`](./docs/developer/plugin-development.md),
[`docs/developer/translating.md`](./docs/developer/translating.md).

Agent workflow: [`AGENTS.md`](./AGENTS.md).

---

## Tech stack (snapshot)

Decided technologies — see [`.plan/tech-stack.md`](./.plan/tech-stack.md).

- **Backend (Go):** chi, SQLite (WAL, pure Go), goose, JWT + refresh tokens, go-i18n.
- **Frontend:** Vite, React, TanStack Query, Zustand, Tailwind + Radix, hls.js.
- **Media:** ffprobe + ffmpeg in the image.
- **API:** OpenAPI 3.1 design-first — [`api/openapi.yaml`](./api/openapi.yaml).
- **Platform:** single Docker image, multi-arch amd64 + arm64, GHCR.

---

## CI gates

Every PR: `lint → i18n-check → sca → unit-test → integration-test → coverage-gate → build → e2e → image-build`.

Coverage thresholds (DoD §4.1): core Go ≥ 80 %, critical modules ≥ 90 %, frontend ≥ 70 %.
Details: [`.plan/definition-of-done.md`](./.plan/definition-of-done.md).

---

## Repository layout

```
.
├── .plan/              # authoritative planning (start: 00-index.md)
├── docs/user/          # en-US + tr-TR user guides
├── docs/developer/     # architecture, plugins, translating, API HTML
├── api/openapi.yaml    # HTTP API contract (source of truth)
├── cmd/ internal/ web/ migrations/ deploy/
└── Makefile
```

Planning index: [`.plan/00-index.md`](./.plan/00-index.md) · Working notes: [`notes/`](./notes/) (non-binding).
