# Sprint 07 Definition of Done Checklist

## DevOps (Wave 1)

- [x] VAAPI `/dev/dri` compose overlay (`deploy/docker-compose.vaapi.yml`)
- [x] NVIDIA compose overlay (`deploy/docker-compose.nvidia.yml`)
- [x] ffmpeg HW libraries in runtime image
- [x] GPU setup guide (`docs/gpu-setup.md`)

## Accelerator detection (Wave 2)

- [x] `SystemProfile.accelerators` + capability matrix via ffmpeg probe
- [x] OpenAPI `AcceleratorInfo` schema
- [x] Unit tests with mock ffmpeg output

## HW transcode pipeline (Wave 3)

- [x] Platform ffmpeg arg builder (`internal/streaming/hwaccel.go`)
- [x] Intel QSV priority in auto selection
- [x] HW→SW fallback on failure
- [x] Integration with `BuildFFmpegArgs` / `StartPackaging`

## Selection engine & limits (Wave 4)

- [x] Path selection based on hardware + settings
- [x] Separate HW session limit
- [x] Settings DB ↔ streaming runtime sync

## Backend settings & telemetry (Wave 5)

- [x] HW mode / accelerator settings API
- [x] Onboarding GPU recommendation in smart defaults
- [x] HW session metrics (`Metrics.HWStarts`, `HWFallbacks`, …)
- [x] Structured HW fallback logging

## Frontend (Wave 6)

- [x] SettingsPage HW acceleration section
- [x] OnboardingWizardPage accelerator recommendation
- [x] i18n en-US + tr-TR
- [x] Vitest component tests

## QA & M4 (Wave 7)

- [x] Unit/integration tests for selection, fallback, settings
- [x] E2E settings HW toggle smoke
- [x] `docs/m4-beta-checklist.md` updated
- [x] Plan sprint task files marked complete

## CI gates

- [ ] `make lint`
- [ ] `make test`
- [ ] `make i18n-check`
- [ ] `make coverage-gate`
- [ ] `make build`
