# Sprint 03 — Definition of Done Checklist

## Backend
- [x] User/role/session/profile/watch schema migrated
- [x] JWT access + hashed refresh tokens
- [x] First admin setup (`POST /setup/admin`)
- [x] Login / refresh / logout / session list
- [x] RBAC middleware on library/media routes
- [x] Profile CRUD with locale + theme
- [x] Parental content rating filter for child profiles
- [x] Watch state, favorites, watchlist APIs
- [x] Brute-force lockout on login
- [x] `internal/auth` coverage ≥ 90%

## Frontend
- [x] Login / setup page
- [x] Protected library routes
- [x] Profile page (locale, theme, parental rating)
- [x] Admin users page
- [x] Auth i18n (`auth.*` en-US + tr-TR)
- [x] Access token in memory + refresh cookie

## QA
- [x] RBAC matrix unit tests
- [x] Auth handler integration tests
- [x] Lockout tests
- [ ] Playwright login → library e2e (run locally: `make e2e`)

## CI
- [x] `make test` (Go)
- [x] `make i18n-check`
- [x] `make build`
- [ ] `make lint` (requires golangci-lint in CI)
- [ ] `make coverage-gate` (full suite)
