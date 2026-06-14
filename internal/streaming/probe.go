package streaming

// MediaProbe holds technical metadata used by the decision engine.
type MediaProbe struct {
	Container     string
	VideoCodec    string
	VideoWidth    int
	VideoHeight   int
	AudioCodec    string
	AudioChannels int
	DurationMs    int64
	Bitrate       int64
}

// EstimatedBitrate returns bits per second when explicit bitrate is unknown.
func (p MediaProbe) EstimatedBitrate() int64 {
	if p.Bitrate > 0 {
		return p.Bitrate
	}
	if p.DurationMs <= 0 {
		return 0
	}
	pixels := int64(p.VideoWidth) * int64(p.VideoHeight)
	if pixels <= 0 {
		return 2_000_000
	}
	// Rough heuristic for untagged files.
	return pixels * 3
}
