package streaming

// LadderTier is one ABR rendition.
type LadderTier struct {
	Name         string
	Width        int
	Height       int
	VideoBitrate int64
	AudioBitrate int64
}

// BuildLadder generates 2–3 tiers scaled from source dimensions.
func BuildLadder(sourceWidth, sourceHeight int) []LadderTier {
	if sourceWidth <= 0 || sourceHeight <= 0 {
		return []LadderTier{
			{Name: "720p", Width: 1280, Height: 720, VideoBitrate: 2_500_000, AudioBitrate: 128_000},
			{Name: "480p", Width: 854, Height: 480, VideoBitrate: 1_000_000, AudioBitrate: 96_000},
		}
	}

	tiers := []LadderTier{
		{
			Name: "source", Width: sourceWidth, Height: sourceHeight,
			VideoBitrate: estimateVideoBitrate(sourceWidth, sourceHeight),
			AudioBitrate: 192_000,
		},
	}

	if sourceHeight >= 720 {
		tiers = append(tiers, LadderTier{
			Name: "720p", Width: 1280, Height: 720,
			VideoBitrate: 2_500_000, AudioBitrate: 128_000,
		})
	}
	if sourceHeight >= 480 {
		tiers = append(tiers, LadderTier{
			Name: "480p", Width: 854, Height: 480,
			VideoBitrate: 1_000_000, AudioBitrate: 96_000,
		})
	}

	if len(tiers) < 2 {
		tiers = append(tiers, LadderTier{
			Name: "low", Width: sourceWidth / 2, Height: sourceHeight / 2,
			VideoBitrate: 800_000, AudioBitrate: 96_000,
		})
	}
	return tiers
}

func estimateVideoBitrate(w, h int) int64 {
	pixels := int64(w) * int64(h)
	switch {
	case pixels >= 1920*1080:
		return 5_000_000
	case pixels >= 1280*720:
		return 2_500_000
	default:
		return 1_200_000
	}
}
