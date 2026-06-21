# Plugin Development

Extends [`docs/plugin-packaging.md`](../plugin-packaging.md) for third-party authors.

## Contract

Plugins implement typed interfaces in `internal/plugin/`:

- **Indexer** — search releases
- **Download client** — add torrents / NZBs, poll status

Build with `-tags acquisition` to include bundled adapters. Core image builds **without** acquisition tags by default.

## Isolation rules

- No direct SQLite access from plugins.
- Secrets stored via plugin secret API (encrypted at rest).
- Outbound HTTP must use pinned/validated clients (`internal/platform/outbound`) to mitigate SSRF.
- Failures in plugins must not crash the core process.
- Plugin and automation API routes require `plugins:manage` (admin by default).

Integration tests: `pluginless_integration_test.go`, `plugin_isolation_security_test.go`.
See [deployment-security.md](../deployment-security.md) and [SECURITY.md](../../SECURITY.md).

## Testing

```bash
go test -tags acquisition ./internal/plugin/...
go test -tags integration ./internal/platform/bootstrap/...
```

Use `implementation: stub` instances for offline CI.

## Packaging

See [`docs/plugin-packaging.md`](../plugin-packaging.md) for build tags, Docker stages, and GHCR layout.
