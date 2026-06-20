# Testing Strategy

Operational complement to [`.plan/definition-of-done.md`](../.plan/definition-of-done.md) §4.
This document describes _how_ Somra tests its code; the binding policy (thresholds, CI
gates, coverage targets) lives in the DoD.

## Pyramid

Three layers, each with a clear responsibility. Higher layers are slower and more brittle —
prefer the lowest layer that actually exercises the behaviour under test.

| Layer              | Stack                                                                               | Scope                                                                                                               | Where it lives                                                         |
| ------------------ | ----------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------- |
| Unit               | Go `testing` + `testify`, Vitest + RTL                                              | Pure functions, single packages, single components. Fast (< 1 s).                                                   | `internal/**/*_test.go`, `web/src/**/*.test.tsx`                       |
| Integration        | Go `testing` with `//go:build integration` tag, in-process SQLite via `t.TempDir()` | Cross-package boundaries: DB ↔ repo, chi router ↔ bootstrap, i18n middleware ↔ error envelope. No external network. | `internal/api/integration_*_test.go`, `internal/platform/db/*_test.go` |
| End-to-end (smoke) | Playwright (`@playwright/test`), Chromium headless                                  | Status page renders against a real backend; SSE event arrives in the panel. One spec per critical flow.             | `web/e2e/*.spec.ts`                                                    |

Performance and load testing are deliberately out of scope for Sprint 01. Sprint 02 adds
parser/provider unit tests and a naming fixture set under `testdata/library/`; a basic
scan performance report can be captured manually via `go test -bench` when needed.

## Coverage policy (binding)

[DoD §4.1](../.plan/definition-of-done.md) is the source of truth. Operationally:

- Go core packages: **≥ 80 %** statement coverage. Measured per package by
  [`scripts/coverage-gate.sh`](../scripts/coverage-gate.sh).
- **Critical** Go packages: **≥ 90 %**. The list lives in `scripts/coverage-gate.sh` —
  changes require Tech Lead approval and a DoD update.
- Frontend statements: **≥ 70 %**. Measured by Vitest v8 coverage `coverage-summary.json`.
- Critical user flows: **e2e mandatory** (login, library browse, playback, setup wizard);
  the percentage gate does not apply, the flow either passes or it doesn't.

### Critical modules — current list (Sprint 01)

| Package                         | Reason                                   |
| ------------------------------- | ---------------------------------------- |
| `internal/auth`                 | RBAC + session security.                 |
| `internal/jobs`                 | Scheduler/queue concurrency.             |
| `internal/platform/db`          | SQLite WAL + migration correctness.      |
| `internal/platform/errors`      | Wire envelope contract.                  |
| `internal/platform/i18n`        | Locale negotiation; user-facing strings. |
| `internal/platform/diagnostics` | `/api/v1/health` correctness.            |

The `cmd/somra` package is **excluded** from the gate (entrypoint wiring, exercised by
integration tests instead).

## Running tests locally

```bash
make test           # unit (Go + Vitest)
make coverage       # unit + writes coverage profile + summary
make coverage-gate  # enforce DoD §4.1 thresholds locally
make i18n-check     # locale parity (en-US ↔ tr-TR)

go test -tags=integration ./...   # integration suite (in-process)

pnpm --dir web exec playwright install --with-deps chromium   # one-time
pnpm --dir web exec playwright test                            # e2e smoke
```

Go unit tests run without `-race` (CGO-free build via `modernc.org/sqlite`; CI matches).

### Database integrity

`internal/platform/db.IntegrityCheck` runs SQLite `PRAGMA integrity_check` and returns
`ok` for a healthy file. Covered by `TestIntegrityCheck_Ok` in
`internal/platform/db/migrate_test.go`; surfaced in `/api/v1/health` diagnostics.

`make lint test i18n-check coverage` must all pass before opening a PR — `make lint`
covers `gofmt`, `golangci-lint`, ESLint + Prettier + `tsc`.

## CI gates

PR merges block on every gate (DoD §5):

```
lint → i18n-check → unit-test → integration-test → coverage-gate → build → image-build
```

Sprint 01 adds an `e2e` job between `build` and `image-build` covering the Playwright
status-page smoke. The job builds the SPA, runs the Go binary with
`SOMRA_WEB_DIR=web/dist` and points Playwright at `http://127.0.0.1:8080`.

## Adding tests for new code

- New business logic in `internal/**` → unit test alongside the change.
- New HTTP endpoint or middleware → unit test for the handler **and** an entry in
  `integration_test.go` that hits it through the chi router.
- New user-facing UI flow → component test (Vitest + RTL) **and**, if the flow is
  declared critical in the sprint brief, an e2e spec.
- New i18n key → both `en-US` and `tr-TR` translations land in the same PR. `make
i18n-check` enforces parity.
