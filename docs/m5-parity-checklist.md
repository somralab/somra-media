# M5 Parity Checklist

Sprint 08 (requests) + Sprint 09 (automation & indexers) complete the M5 milestone.

## Sprint 08 — Requests (done)

- [x] User content request discover + create flow
- [x] Admin approve/reject queue and policies
- [x] Request SSE/polling status updates
- [x] Frontend: discover, my requests, admin queue

## Sprint 09 — Automation & indexers (done)

- [x] Plugin isolation — core operates without enabled plugins (`pluginless_integration_test`)
- [x] Plugin lifecycle — create, configure, disable, enable, delete (`plugin_lifecycle_integration_test`)
- [x] Indexer search via stub plugins (`automation_integration_test`)
- [x] Download client monitor + handoff pipeline
- [x] Quality profiles list/create/get/patch (API + UI)
- [x] Import → library scan path (`TestIntegration_AutomationImportToLibrary`)
- [x] Series monitors CRUD + scanner job (API + UI)
- [x] Frontend automation hub, indexers, download clients, downloads, monitors

## CI gates

- [x] lint → i18n-check → unit-test → integration-test → coverage-gate → build
- [x] Integration: `go test -tags=integration ./internal/platform/bootstrap/...`
- [x] E2E (advisory): Playwright automation specs under `web/e2e/automation-*.spec.ts`

## Known open items (deferred)

- [ ] `web/e2e/playback.spec.ts` — still skipped; playback e2e tracked for Sprint 10+
- [ ] Download SSE real-time — MVP uses polling (`refetchInterval` 5s)

## E2E coverage

| Spec | Scope |
|------|--------|
| `automation-plugins.spec.ts` | Admin stub indexer + download client, test connection |
| `automation-downloads.spec.ts` | Downloads page loads for admin |
| `automation-monitors.spec.ts` | Monitor CRUD smoke via UI |
