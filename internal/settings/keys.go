package settings

// Setting keys stored in the settings table.
const (
	KeyDefaultLocale           = "system.default_locale"
	KeyOnboardingCompleted     = "onboarding.completed"
	KeyOnboardingPhase         = "onboarding.phase"
	KeyStreamingMaxConcurrent  = "streaming.max_concurrent_transcodes"
	KeyStreamingMaxHW          = "streaming.max_concurrent_hw_transcodes"
	KeyStreamingHWMode         = "streaming.hw_mode"
	KeyStreamingHWAccelerator  = "streaming.hw_accelerator"
	KeyLibraryScanCron         = "library.scan_cron"
	KeySubtitlesAutoDownload   = "subtitles.auto_download"
	KeySubtitlesPreferredLangs = "subtitles.preferred_languages"
	KeySubtitlesAPIKey         = "subtitles.opensubtitles_api_key"
)

// Category names for the settings API.
const (
	CategoryGeneral   = "general"
	CategoryLibrary   = "library"
	CategoryPlayback  = "playback"
	CategorySubtitles = "subtitles"
	CategoryStreaming = "streaming"
)

// Phase values for the onboarding wizard.
const (
	PhaseLanguage = "language"
	PhaseAdmin    = "admin"
	PhaseLibrary  = "library"
	PhaseDefaults = "defaults"
	PhaseScan     = "scan"
	PhaseComplete = "complete"
)
