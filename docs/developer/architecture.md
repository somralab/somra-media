# Architecture (developer summary)

Authoritative module boundaries: [`.plan/architecture.md`](../../.plan/architecture.md).

Somra is a **monolithic Go binary** serving a React SPA. Modules:

| Module | Path | Role |
|--------|------|------|
| API gateway | `internal/api` | chi router, auth, handlers |
| Auth | `internal/auth` | JWT + refresh tokens, RBAC |
| Library | `internal/library` | scan, watch, path validation |
| Metadata | `internal/metadata` | provider registry, matching |
| Streaming | `internal/streaming` | transcode, HLS sessions |
| Requests | `internal/requests` | discover, approve workflow |
| Automation | `internal/automation` | indexers, downloads, monitors |
| Plugins | `internal/plugin` | isolated acquisition adapters |
| Platform | `internal/platform` | db, config, diagnostics, i18n |

Data flow: **scan → metadata match → browse → playback transcode decision → HLS delivery**.

Definition of Done: [`.plan/definition-of-done.md`](../../.plan/definition-of-done.md).
