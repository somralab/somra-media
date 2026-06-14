package streaming

import "strings"

// ClientCapabilities describes what the playback client can decode.
type ClientCapabilities struct {
	VideoCodecs      []string `json:"videoCodecs"`
	AudioCodecs      []string `json:"audioCodecs"`
	Containers       []string `json:"containers"`
	MaxBitrate       int64    `json:"maxBitrate"`
	SupportsHDR      bool     `json:"supportsHdr"`
	SupportsHEVC     bool     `json:"supportsHevc"`
	MaxAudioChannels int      `json:"maxAudioChannels"`
}

// DefaultBrowserCapabilities returns a conservative Chrome-like profile.
func DefaultBrowserCapabilities() ClientCapabilities {
	return ClientCapabilities{
		VideoCodecs:      []string{"h264", "avc1"},
		AudioCodecs:      []string{"aac", "mp4a"},
		Containers:       []string{"mp4", "m4v", "mov", "webm"},
		MaxBitrate:       20_000_000,
		SupportsHDR:      false,
		SupportsHEVC:     false,
		MaxAudioChannels: 2,
	}
}

func normalizeCodec(c string) string {
	c = strings.ToLower(strings.TrimSpace(c))
	switch {
	case strings.Contains(c, "h264"), strings.Contains(c, "avc"):
		return "h264"
	case strings.Contains(c, "hevc"), strings.Contains(c, "h265"):
		return "hevc"
	case strings.Contains(c, "aac"), strings.Contains(c, "mp4a"):
		return "aac"
	case strings.Contains(c, "ac3"), strings.Contains(c, "eac3"):
		return "ac3"
	case strings.Contains(c, "dts"):
		return "dts"
	default:
		return c
	}
}

func containsCodec(list []string, want string) bool {
	want = normalizeCodec(want)
	for _, c := range list {
		if normalizeCodec(c) == want {
			return true
		}
	}
	return false
}

func containsContainer(list []string, want string) bool {
	want = strings.ToLower(strings.TrimSpace(want))
	for _, part := range strings.Split(want, ",") {
		part = strings.TrimSpace(part)
		for _, c := range list {
			if strings.EqualFold(c, part) {
				return true
			}
		}
	}
	return false
}
