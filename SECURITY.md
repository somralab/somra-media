# Security Policy

Somra takes security seriously. This document describes how to report
vulnerabilities, what we support, and what you can expect from us.

**License:** [AGPL-3.0-or-later](LICENSE). Security fixes are published under the
same license.

## Supported Versions

Somra is in **active development** — Sprint 09 in progress; **1.0** is planned in
Sprint 10. Until then, treat production deployments as **pre-release**.

We provide security fixes only for versions we actively maintain:

| Version / channel | Supported |
| ----------------- | --------- |
| `main` (latest commit) | ✅ |
| Latest tagged release on GitHub | ✅ |
| Older tagged releases | ❌ (upgrade to the latest tag or `main`) |
| Custom / forked builds | ❌ (maintain your own fork) |

After **1.0**, this table will list supported SemVer ranges (e.g. `1.x` supported,
`0.x` not). Tagged releases and `ghcr.io/somralab/somra-media` image tags will
follow the same policy.

Check your running version:

```bash
curl -s http://localhost:8080/api/v1/version | jq
