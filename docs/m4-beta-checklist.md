# M4 Beta Checklist

Sprint 06 + Sprint 07 complete the M4 beta candidate core.

## Sprint 06 complete

- [x] Zero-config wizard path exists (<10 min target)
- [x] System detection (CPU, RAM, GPU presence, path validation)
- [x] Smart defaults (transcode concurrency, scan cron)
- [x] Central settings API with categories
- [x] Subtitle provider interface + OpenSubtitles integration
- [x] Auto-download job scheduled

## Sprint 07 complete (hardware acceleration)

- [x] GPU / accelerator detection (QSV, NVENC, VAAPI, AMF)
- [x] Hardware transcode pipeline with HW→SW fallback
- [x] Auto accelerator selection (Intel QSV priority)
- [x] HW session limits separate from CPU transcode limit
- [x] Settings + onboarding HW management UI
- [x] Docker GPU passthrough documented (VAAPI + NVIDIA overlays)
- [x] HW telemetry and structured fallback logging

## M4 beta candidate criteria

- [x] Onboarding wizard with smart HW defaults
- [x] Settings API for playback/HW configuration
- [x] At least one HW path implemented with safe fallback
- [x] CI gates: lint, i18n, test, coverage, build
