# M6 E2E Scenario Matrix

Playwright specs under `web/e2e/`. Run locally: `bash scripts/gen-e2e-media.sh && make e2e`.

| # | Scenario | Spec | Priority |
|---|----------|------|----------|
| 1 | Health / status page | `status.spec.ts` | P0 |
| 2 | Onboarding wizard | `00-onboarding.spec.ts` | P0 |
| 3 | Auth login/logout | `auth.spec.ts` | P0 |
| 4 | Browse libraries | `browse.spec.ts` | P0 |
| 5 | Search | `search.spec.ts` | P1 |
| 6 | Media detail | `browse.spec.ts` | P1 |
| 7 | **Playback** (play + seek) | `playback.spec.ts` | P0 |
| 8 | Subtitles section | `subtitles.spec.ts` | P1 |
| 9 | Settings | `settings.spec.ts` | P1 |
| 10 | Themes / locale | `themes.spec.ts` | P2 |
| 11 | Requests flow | `requests.spec.ts` | P1 |
| 12 | Notifications | `notifications.spec.ts` | P2 |
| 13 | Automation plugins | `automation-plugins.spec.ts` | P1 |
| 14 | Automation downloads | `automation-downloads.spec.ts` | P1 |
| 15 | Series monitors | `automation-monitors.spec.ts` | P1 |

## End-to-end demo script (M6)

1. Pull `ghcr.io/somralab/somra-media:1.0.0` → running server
2. Complete onboarding → add library → scan
3. Play title (direct or transcode)
4. Submit request → admin approve → automation stub flow
5. Switch locale tr-TR → verify UI strings
6. Show `/api/v1/health` diagnostics

## Browser matrix (manual B2)

| Browser | Version | Status |
|---------|---------|--------|
| Chrome | latest | |
| Firefox | latest | |
| Safari | latest macOS | |

## Fixtures

| Fixture | Path | Generator |
|---------|------|-----------|
| Playback MP4 | `testdata/media/Sample Movie (2010).mp4` | `scripts/gen-e2e-media.sh` |
| E2E admin user | created by `web/e2e/helpers.ts` | automatic |

## Out of scope (1.0)

- Download SSE real-time (polling MVP)
- Mobile/TV clients
