# Issue Severity Model

Operational support for [`.plan/definition-of-done.md`](../.plan/definition-of-done.md) §2,
which requires sprints to ship with no outstanding **critical** or **high** issues.

The intent is to keep triage predictable: severity drives whether an issue blocks a sprint,
whether the on-call rotation should be paged, and how aggressively a fix is rolled out.

## Levels

| Severity          | Definition                                                                                                                                                   | Example                                                                                                                    | Response SLA (acknowledge / patch)    |
| ----------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------- | ------------------------------------- |
| **Critical (P1)** | Production outage, data loss, security breach, or a regression that prevents any user from completing a core flow (login, browse, playback). No workaround.  | `/api/v1/health` returns 500 in production; auth token signing key leak; SQLite migration drops a column.                  | **≤ 1 h / ≤ 24 h** (hotfix)           |
| **High (P2)**     | A core flow is degraded or one of the four critical modules misbehaves, but a workaround exists. Blocks the current sprint from being marked Done.           | Playback works only on Chrome; `/api/v1/version` omits the commit; CI consistently flakes on the integration gate.         | **≤ 1 business day / ≤ 1 week**       |
| **Medium (P3)**   | Non-core flow defect, UX nit on a critical screen, accessibility regression below WCAG AA. Tracked in the backlog and addressed in the next planning window. | A localized string truncates in Turkish; a non-critical setting cannot be reset; an SSE keepalive logs at the wrong level. | **≤ 1 sprint / next planned release** |
| **Low (P4)**      | Cosmetic, documentation, micro-optimization, code health. Pulled into a sprint when capacity allows; otherwise stays in the backlog.                         | A help icon is slightly misaligned in dark mode; an internal log entry uses `Error` where `Warn` would be more accurate.   | **Best-effort**                       |

The response SLAs are upper bounds: "acknowledge" means an owner is assigned and the
status is updated; "patch" means a fix is merged (and, for critical/high, rolled out).

## Picking the right level

1. **Does it break a critical user flow with no workaround?** → P1.
2. **Does it affect a critical module, or block the sprint's acceptance criteria?** → P2.
3. **Does it visibly hurt a user, but not core flows?** → P3.
4. **Otherwise** → P4.

Default to the higher severity when unsure. Lowering severity later is cheaper than the
fallout from missing a regression.

## Sprint closure rule

A sprint **cannot** be marked Done with open P1/P2 issues that fall inside its scope.
Open P3/P4 issues roll into the backlog with a label referencing the sprint that
introduced them. See DoD §2.4 for the binding rule.

## Security-specific notes

- Suspected credential leak or unauthorized data access → always **P1**, regardless of
  reproducibility.
- Vulnerabilities in third-party deps (CVE) → severity follows CVSS: ≥ 9.0 = P1, 7.0–8.9
  = P2, 4.0–6.9 = P3, < 4.0 = P4.
- Path traversal / SSRF / authz bypass → minimum **P2**, even when behind a feature
  flag.
