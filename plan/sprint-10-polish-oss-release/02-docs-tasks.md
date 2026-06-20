# Sprint 10 — Documentation & Community Tasks

> **Sprint goal:** Complete user/developer documentation and community infrastructure for
> open source release.
>
> **Related:** [`../project-brief.md`](../project-brief.md) §5 (license) · [`03-devops-release-tasks.md`](./03-devops-release-tasks.md)

## Responsible Role(s)
- PM (primary), Tech Lead, entire team (contributions)

## Dependencies
- Stable feature set (Sprint 01–09).

## Epics and Tasks

### Epic A: User documentation (TR + EN)
- [ ] A1 — Installation guide (Docker/compose, GPU passthrough) | Acceptance: installable with zero prior knowledge; **tr-TR + en-US**.
- [ ] A2 — Feature/usage guides (library, playback, requests, automation) | Acceptance: main flows documented; **tr-TR + en-US**.
- [ ] A3 — FAQ + troubleshooting | Acceptance: common issues covered; **tr-TR + en-US**.
- [ ] A4 — Documentation translation completeness (TR↔EN parity) | Acceptance: both languages complete. See [`../i18n-localization.md`](../i18n-localization.md) §2.

### Epic B: Developer documentation
- [ ] B1 — Architecture/contribution guide (CONTRIBUTING) + code standards | Acceptance: aligned with [`../definition-of-done.md`](../definition-of-done.md).
- [ ] B2 — API documentation (OpenAPI publication) | Acceptance: current contract.
- [ ] B3 — Plugin development guide | Acceptance: third-party plugins can be written.

### Epic C: Community & license
- [ ] C1 — LICENSE (AGPL-3.0 approved), Code of Conduct, issue/PR templates | Acceptance: repo meets OSS standards.
- [ ] C2 — README + project introduction (Somra brand) | Acceptance: clear value proposition.
- [ ] C3 — Translation contribution guide + **Weblate** setup (self-host, git integrated) | Acceptance: community can contribute translations via Weblate; repo sync works. See [`../i18n-localization.md`](../i18n-localization.md) §8.

## Acceptance Criteria (Sprint Output)
- User and developer documentation complete; license and community infrastructure ready.

## Risks
- Missing docs reduce adoption → must be completed before release.

## Out of Scope
- Translation to languages other than TR/EN — future (infrastructure will be ready to add new languages; see [`../i18n-localization.md`](../i18n-localization.md) §1).
