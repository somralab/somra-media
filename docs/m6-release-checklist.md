# M6 Release Checklist

Sprint 10 / 1.0 open-source release gate. Sign off when all items are verified.

## Sprint 10 epics

### Performance (01)
- [x] Profiling script (`scripts/profile.sh`) + baseline doc
- [x] Browse DB indexes migration
- [ ] Measured numbers recorded on 4c/8GB hardware
- [x] Soak script (`scripts/soak.sh`)

### Documentation (02)
- [x] User docs en-US + tr-TR (`docs/user/`)
- [x] Developer docs (`docs/developer/`)
- [x] CODE_OF_CONDUCT.md + PR template
- [x] README 1.0 refresh
- [x] Weblate MVP compose

### DevOps (03)
- [ ] Local release procedure (tag + git-cliff changelog + `make docker-multiarch`); `release.yml` removed 2026-10-13, no CI/CD
- [ ] GHCR `latest` + semver tags pushed by hand on release
- [x] Upgrade integration test
- [x] Backup/restore doc
- [x] Production compose example
- [x] Extended health diagnostics

### Security (04)
- [x] RBAC matrix integration test
- [x] SSRF tests
- [x] SCA gate (`make sca`: govulncheck + pnpm audit)
- [x] SECURITY.md complete

### QA (05)
- [x] Playback e2e enabled
- [x] M6 e2e matrix doc
- [x] i18n overflow gate
- [ ] Full sprint DoD checklist regression (manual)
- [ ] No open critical/high bugs

## Gates / release

- [ ] `make check` green on `main`
- [ ] `make e2e` green
- [ ] `v1.0.0` tag created
- [ ] GHCR image `ghcr.io/somralab/somra-media:1.0.0` published
- [ ] GitHub Release notes published

## Acceptance (project brief §8)

- [ ] Install < 10 min on clean VM (document timing)
- [ ] Runs on 4c/8GB home hardware
- [ ] Secure defaults documented
- [ ] en-US + tr-TR UI and user docs complete
- [ ] Upgrade from prior DB snapshot preserves data

## Sign-off

| Role | Name | Date |
|------|------|------|
| Tech Lead | | |
| PM | | |
| QA | | |
