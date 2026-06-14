package streaming

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelectTranscodePath_AutoQSVPriority(t *testing.T) {
	cfg := HWRuntimeConfig{
		Mode:          HWModeAuto,
		Available:     []Accelerator{AccelNVENC, AccelQSV, AccelVAAPI},
		MaxHWSessions: 2,
	}
	path := SelectTranscodePath(cfg, MediaProbe{VideoCodec: "hevc"}, 0)
	assert.True(t, path.UseHW)
	assert.Equal(t, AccelQSV, path.Accelerator)
	assert.Equal(t, "h264_qsv", path.VideoEncoder)
}

func TestSelectTranscodePath_HWOff(t *testing.T) {
	cfg := HWRuntimeConfig{Mode: HWModeOff, Available: []Accelerator{AccelQSV}}
	path := SelectTranscodePath(cfg, MediaProbe{}, 0)
	assert.False(t, path.UseHW)
	assert.Equal(t, "libx264", path.VideoEncoder)
}

func TestSelectTranscodePath_SessionLimit(t *testing.T) {
	cfg := HWRuntimeConfig{
		Mode:          HWModeAuto,
		Available:     []Accelerator{AccelQSV},
		MaxHWSessions: 2,
	}
	path := SelectTranscodePath(cfg, MediaProbe{}, 2)
	assert.False(t, path.UseHW)
	assert.Equal(t, "hw_session_limit", path.SelectionNote)
}

func TestSelectTranscodePath_PreferredNVENC(t *testing.T) {
	cfg := HWRuntimeConfig{
		Mode:          HWModeAuto,
		Preferred:     AccelNVENC,
		Available:     []Accelerator{AccelQSV, AccelNVENC},
		MaxHWSessions: 2,
	}
	path := SelectTranscodePath(cfg, MediaProbe{}, 0)
	assert.Equal(t, AccelNVENC, path.Accelerator)
}

func TestAppendHWVideoArgs_QSV(t *testing.T) {
	args := AppendHWVideoArgs([]string{"-i", "in.mkv"}, TranscodePath{
		UseHW: true, Accelerator: AccelQSV, VideoEncoder: "h264_qsv",
	}, LadderTier{Width: 1280, Height: 720, VideoBitrate: 2_500_000}, "")
	assert.Contains(t, args, "h264_qsv")
	assert.Contains(t, args, "-hwaccel")
}

func TestAppendHWVideoArgs_SWFallback(t *testing.T) {
	args := AppendHWVideoArgs([]string{}, TranscodePath{UseHW: false}, LadderTier{VideoBitrate: 1_000_000}, "")
	assert.Contains(t, args, "libx264")
}

func TestAcceleratorsFromSettings(t *testing.T) {
	out := AcceleratorsFromSettings([]string{"qsv", "nvenc", "invalid"})
	assert.Equal(t, []Accelerator{AccelQSV, AccelNVENC}, out)
}
