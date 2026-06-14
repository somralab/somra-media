package settings

// SmartDefaults holds CPU-based recommendations for a fresh install.
type SmartDefaults struct {
	MaxConcurrentTranscodes int    `json:"maxConcurrentTranscodes"`
	ScanCron                string `json:"scanCron"`
	DefaultLocale           string `json:"defaultLocale,omitempty"`
}

const defaultScanCron = "0 3 * * *"

// RecommendDefaults derives conservative transcode concurrency from CPU cores.
func RecommendDefaults(cpuCores int, locale string) SmartDefaults {
	concurrency := 2
	switch {
	case cpuCores <= 2:
		concurrency = 1
	case cpuCores <= 4:
		concurrency = 2
	default:
		concurrency = 2
	}
	if locale == "" {
		locale = "en-US"
	}
	return SmartDefaults{
		MaxConcurrentTranscodes: concurrency,
		ScanCron:                defaultScanCron,
		DefaultLocale:           locale,
	}
}
