# Sprint 01 (M1 foundation) — Definition of Done Checklist

Closes [`plan/sprint-01-foundation/06-qa-tasks.md`](../plan/sprint-01-foundation/06-qa-tasks.md)
Epik C. Items below mirror [`plan/definition-of-done.md`](../plan/definition-of-done.md)
§1 (task-level), §2 (sprint-level) and §6 (security defaults) plus the i18n acceptance
criteria from [`plan/i18n-localization.md`](../plan/i18n-localization.md) §6.

Tick the boxes that are satisfied today; leave open items for the sprints that own them.

---

## DoD §1 — Task-level (must be true for every Sprint 01 task)

- [x] **1.1** Acceptance criteria documented in the per-discipline task files.
- [x] **1.2** Code respects module boundaries (`plan/architecture.md` §3): `cmd/`,
      `internal/api`, `internal/platform/*`, `internal/jobs`, `internal/auth` (interface
      only), `web/`, `migrations/`, `api/`, `deploy/`.
- [x] **1.3** Unit tests written; integration tests added for cross-package boundaries;
      coverage thresholds (§4.1) met locally and in CI.
- [x] **1.4** Lint clean: `gofmt`, `golangci-lint` v2, ESLint, Prettier, `tsc --noEmit`.
- [x] **1.5** README + per-package docs updated (`README.md`, `web/README.md`,
      `docs/testing-strategy.md`, `docs/issue-severity.md`).
- [ ] **1.6** ≥ 1 code review approval per packet (handled at PR review time).
- [x] **1.7** CI pipeline runs all seven gates:
      `lint → i18n-check → unit-test → integration-test → coverage-gate → build → image-build`.
- [x] **1.8** i18n compliance — see DoD §i18n below.

## DoD §2 — Sprint-level

- [x] **2.1** All Sprint 01 "must" tasks (Paket 1–8) merged into `feat/sprint-01-m1`.
- [x] **2.2** Demo recipe is reproducible: `docker compose -f deploy/docker-compose.yml up`
      exposes `/api/v1/health`, `/api/v1/version`, the SSE stream, and serves the SPA on
      `http://localhost:8080`. See [`README.md` § Sprint 01 (M1) demo](../README.md).
- [x] **2.3** Regression: unit + integration + e2e harnesses all green.
- [x] **2.4** No outstanding P1/P2 issues against Sprint 01 scope (see
      [`issue-severity.md`](./issue-severity.md)).
- [x] **2.5** Outputs the next sprint depends on are published: Go module with
      `internal/api`, `internal/platform/{config,bootstrap,db,diagnostics,errors,i18n,log}`,
      `internal/jobs`, the OpenAPI spec, the Dockerfile, the SPA scaffold.

## i18n acceptance (`plan/i18n-localization.md` §6)

- [x] **i18n.1** No user-facing hardcoded strings — verified by `make i18n-check` and
      the `cmd/i18n-check` Go tool that compares the en-US ↔ tr-TR key sets across both
      backend (`internal/platform/i18n/locales/*.toml`) and frontend
      (`web/src/i18n/locales/*/*.json`).
- [x] **i18n.2** Every user-facing key exists in both en-US and tr-TR.
- [x] **i18n.3** Error envelopes localize via the `i18n.Middleware`; the integration
      test `TestIntegration_NotFoundLocalizedTurkish` asserts a Turkish 404 message.
- [x] **i18n.4** Frontend resolves all visible text via `react-i18next`; SPA tests cover
      both locales (`web/src/test/i18n.test.tsx`).
- [x] **i18n.5** Dates/numbers go through `Intl` (backend uses `golang.org/x/text`).

## Security defaults (`definition-of-done.md` §6)

- [x] **sec.1** Validate every input (chi handlers reject non-JSON bodies via
      `ContentTypeMiddleware`; SPA falls back through `errors.not_found` on traversal).
- [x] **sec.2** No secrets in tree — `.env.example` documents required env vars; no
      `.env` is committed; `.gitignore` excludes data/cache/dist artifacts.
- [x] **sec.3** All queries parameterized — `internal/platform/db` uses
      `database/sql` with `?` placeholders; no string-built SQL.
- [x] **sec.4** SPA respects path-traversal guard in `mountSPA` (`router.go`).
- [ ] **sec.5** RBAC + brute-force protection — **deferred to Sprint 03** (auth flow).
- [ ] **sec.6** SSRF/path-traversal hardening on outbound providers — **deferred to
      Sprint 02** (metadata providers introduced there).
- [ ] **sec.7** Plugin isolation for indexers — **deferred to Sprint 07**.

## Outstanding follow-ups (handed to later sprints)

- Sprint 02: metadata provider stack + SSRF guard.
- Sprint 03: auth/RBAC, refresh token revocation, brute-force protection (DoD §6
  items above).
- Sprint 03+: real business endpoints behind the gateway (currently the gateway only
  surfaces health/version/SSE).
- Sprint 04: streaming pipeline; performance test harness.
- Sprint 07: plugin sandbox for content acquisition.
- Sprint 10: Weblate self-host; UI for locale switching beyond the system default.

---

**Reviewer cross-check**: every checked item must be linked to a file, command, or test
elsewhere in this repo. Items left unchecked are intentional carry-overs documented in
the relevant sprint folder under [`plan/`](../plan/).
