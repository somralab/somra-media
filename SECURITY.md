# Security Policy

Somra takes security seriously. This document describes supported versions,
how to report vulnerabilities, and what you can expect from us.

**License:** [AGPL-3.0-or-later](LICENSE). Security fixes ship under the same license.

## Supported Versions

| Version | Supported |
| ------- | --------- |
| **1.x** (latest tagged release) | ✅ |
| `main` (development) | ✅ best-effort |
| **0.x** / pre-1.0 tags | ❌ upgrade to 1.x |
| Custom / forked builds | ❌ maintain your own fork |

Container images at `ghcr.io/somralab/somra-media` follow the same policy.
Tags `latest`, `1.0.0`, and `1.x.y` receive security fixes.

Check your running version:

```bash
curl -s http://localhost:8080/api/v1/version | jq
```

## Reporting a Vulnerability

**Please do not open public GitHub issues for security vulnerabilities.**

Report privately via:

1. **Email:** security@somralab.com
2. **GitHub:** [Private security advisory](https://github.com/somralab/somra-media/security/advisories/new)

Include: affected version, reproduction steps, impact assessment, and any
proof-of-concept if available.

## Response Timeline

| Severity | Target fix |
| -------- | ---------- |
| **Critical** (RCE, auth bypass, data exfiltration) | 90 days |
| **High** (privilege escalation, SSRF to internal network) | 90 days |
| **Medium** | next minor release |
| **Low** | backlog / next convenient release |

We acknowledge reports within **5 business days** and provide status updates
at least every **14 days** until resolution.

## Severity Definitions

See [`docs/issue-severity.md`](docs/issue-severity.md) for project-wide
severity levels used in triage and release gates.

## Secure Defaults (1.0)

- JWT secret must be set via `SOMRA_JWT_SECRET` in production (compose warns on dev default).
- CORS defaults to local dev origins; restrict in production.
- Rate limiting on auth endpoints (Sprint 03).
- Outbound metadata/indexer calls use SSRF allowlists (`internal/metadata/ssrf.go`).
- Plugin adapters run isolated; core operates without plugins enabled.
- SQLite WAL mode; parameterized queries throughout data layer.

## Dependency Scanning

CI runs `govulncheck ./...` and `pnpm audit --audit-level=high` on every PR.
Critical findings block merge until remediated or accepted with documented risk.

## Plugin Security

Third-party acquisition plugins must not access the core database directly.
See [`docs/developer/plugin-development.md`](docs/developer/plugin-development.md)
and [`docs/plugin-packaging.md`](docs/plugin-packaging.md).

## Disclosure Policy

We coordinate disclosure with reporters. After a fix is released we publish:

- GitHub Security Advisory
- Release notes in the tagged release
- CVE assignment when applicable

Thank you for helping keep Somra and our users safe.
