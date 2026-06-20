# Sprint 10 — Security Audit Tasks

> **Sprint goal:** Comprehensive security audit before open source release and validation of
> secure defaults.
>
> **Related:** [`../definition-of-done.md`](../definition-of-done.md) §6 · Sprint 03 [`../sprint-03-auth-users/04-security-tasks.md`](../sprint-03-auth-users/04-security-tasks.md)

## Responsible Role(s)
- Tech Lead (primary), Backend, QA

## Dependencies
- Security-related outputs from all sprints.

## Epics and Tasks

### Epic A: Security audit
- [ ] A1 — Identity/permission/session surface audit (RBAC scope validation) | Acceptance: no permission violations found.
- [ ] A2 — Input validation/injection/SSRF/path traversal audit | Acceptance: known classes closed.
- [ ] A3 — Dependency security scan (SCA) + license compliance | Acceptance: no critical vulnerabilities/incompatible licenses.

### Epic B: Secure defaults
- [ ] B1 — Default configuration hardening (secrets, CORS, rate limit, HTTPS guide) | Acceptance: secure defaults.
- [ ] B2 — Plugin isolation security review | Acceptance: plugins cannot compromise core.

### Epic C: Process
- [ ] C1 — Security policy (SECURITY.md) + responsible disclosure process | Acceptance: available at release.

## Acceptance Criteria (Sprint Output)
- No critical/high security findings remain; secure defaults and security policy ready.

## Risks
- Late-discovered vulnerabilities delay release → audit should start at sprint beginning.

## Out of Scope
- Official third-party penetration test certification — optional/future.
