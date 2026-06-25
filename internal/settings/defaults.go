package settings

// SmartDefaults holds CPU/GPU-based recommendations for a fresh install.
type SmartDefaults struct {
	MaxConcurrentTranscodes int           `json:"maxConcurrentTranscodes"`
	MaxHWTranscodes         int           `json:"maxHWTranscodes,omitempty"`
	HWMode                  HWMode        `json:"hwMode,omitempty"`
	RecommendedAccelerator  AcceleratorID `json:"recommendedAccelerator,omitempty"`
	ScanCron                string        `json:"scanCron"`
	DefaultLocale           string        `json:"defaultLocale,omitempty"`
}

const defaultScanCron = "0 3 * * *"

// RecommendDefaults derives conservative transcode concurrency from CPU cores and GPU detection.
func RecommendDefaults(cpuCores int, locale string) SmartDefaults {
	return RecommendDefaultsWithProfile(cpuCores, locale, DetectSystem(nil))
}

// RecommendDefaultsWithProfile uses a full system profile for HW recommendations.
func RecommendDefaultsWithProfile(cpuCores int, locale string, profile SystemProfile) SmartDefaults {
	concurrency := 2
	if cpuCores <= 2 {
		concurrency = 1
	}
	if locale == "" {
		locale = "en-US"
	}
	def := SmartDefaults{
		MaxConcurrentTranscodes: concurrency,
		ScanCron:                defaultScanCron,
		DefaultLocale:           locale,
		HWMode:                  HWModeOff,
	}
	if profile.RecommendedAccelerator != "" {
		def.HWMode = HWModeAuto
		def.RecommendedAccelerator = profile.RecommendedAccelerator
		def.MaxHWTranscodes = 1
		if cpuCores >= 6 {
			def.MaxHWTranscodes = 2
		}
		if cpuCores >= 8 {
			def.MaxHWTranscodes = 3
		}
	}
	return def
}
