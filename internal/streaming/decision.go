package streaming

import "strings"

// DecisionEngine selects direct play, direct stream, or transcode.
type DecisionEngine struct{}

// NewDecisionEngine returns a decision engine with default rules.
func NewDecisionEngine() *DecisionEngine {
	return &DecisionEngine{}
}

// Decide picks the lowest-cost delivery mode that satisfies client capabilities.
func (e *DecisionEngine) Decide(caps ClientCapabilities, probe MediaProbe) Decision {
	video := normalizeCodec(probe.VideoCodec)
	audio := normalizeCodec(probe.AudioCodec)

	videoSupported := video == "h264" || (video == "hevc" && (caps.SupportsHEVC || containsCodec(caps.VideoCodecs, "hevc")))
	audioSupported := audio == "aac" || containsCodec(caps.AudioChannelsSupported(probe.AudioChannels), audio)

	if !videoSupported || !audioSupported {
		return Decision{Mode: ModeTranscode, Reason: "unsupported_codec"}
	}

	if probe.AudioChannels > caps.MaxAudioChannels && caps.MaxAudioChannels > 0 {
		return Decision{Mode: ModeTranscode, Reason: "audio_channels_downmix"}
	}

	bitrate := probe.EstimatedBitrate()
	if caps.MaxBitrate > 0 && bitrate > caps.MaxBitrate {
		return Decision{Mode: ModeTranscode, Reason: "bitrate_exceeds_client"}
	}

	containerOK := isDirectPlayContainer(probe.Container)
	if containerOK && containsContainer(caps.Containers, probe.Container) {
		return Decision{Mode: ModeDirectPlay, Reason: "browser_safe_container"}
	}

	return Decision{Mode: ModeDirectStream, Reason: "remux_required"}
}

func isDirectPlayContainer(container string) bool {
	for _, part := range strings.Split(container, ",") {
		switch strings.ToLower(strings.TrimSpace(part)) {
		case "mp4", "m4v", "mov", "ism", "cmaf":
			return true
		}
	}
	return false
}

// AudioChannelsSupported returns codecs the client can play for the channel count.
func (c ClientCapabilities) AudioChannelsSupported(channels int) []string {
	if channels <= c.MaxAudioChannels || c.MaxAudioChannels <= 0 {
		return c.AudioCodecs
	}
	return c.AudioCodecs
}
