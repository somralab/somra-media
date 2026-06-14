# Sprint 06 Definition of Done Checklist

- [x] OpenAPI 3.1: system detect, settings, onboarding, subtitles endpoints
- [x] `make openapi-types` generates frontend types
- [x] Backend: `internal/settings` (detect, defaults, settings API, onboarding SM)
- [x] Backend: `internal/subtitles` (provider, OpenSubtitles, auto-download job)
- [x] SSRF allowlist includes OpenSubtitles hosts
- [x] Frontend: onboarding wizard (`/setup/wizard`)
- [x] Frontend: extended settings page with advanced toggle
- [x] Frontend: subtitle management on media detail page
- [x] i18n: `onboarding`, `settings`, `subtitles` namespaces (en-US + tr-TR)
- [x] Unit tests: settings, subtitles, API handlers
- [x] E2E: onboarding, settings, subtitles specs
- [ ] `make lint test i18n-check coverage build` all green (verify in CI)

## Manual demo steps

1. Fresh DB → open `/setup/wizard`
2. Select language → create admin → add library → apply smart defaults
3. Trigger scan → complete wizard
4. Configure OpenSubtitles API key in Settings → advanced → subtitles
5. Open media detail → search/download/upload subtitles
