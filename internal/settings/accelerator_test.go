package settings

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const mockEncoderOutput = `
 V..... h264_qsv             H.264 / AVC (Intel Quick Sync Video)
 V..... hevc_qsv             HEVC (Intel Quick Sync Video)
 V..... h264_nvenc           NVIDIA NVENC H.264 encoder
 V..... h264_vaapi           H.264/AVC (VAAPI)
 V..... h264_amf             AMD AMF H.264 Encoder
`

const mockDecoderOutput = `
 V..... h264_qsv             H.264 / AVC (Intel Quick Sync Video)
 V..... h264_cuvid           Nvidia CUVID H.264 decoder
 V..... h264_vaapi           H.264/AVC (VAAPI)
`

const mockHWAccelOutput = `
Hardware acceleration methods:
qsv
cuda
vaapi
`

func withMockFFmpegProbe(t *testing.T, fn func(ctx context.Context, bin string, flag string) string) {
	t.Helper()
	orig := ffmpegProbeFn
	ffmpegProbeFn = fn
	t.Cleanup(func() { ffmpegProbeFn = orig })
}

func TestProbeAccelerators_MockFFmpeg(t *testing.T) {
	withMockFFmpegProbe(t, func(_ context.Context, _ string, flag string) string {
		switch flag {
		case "-encoders":
			return mockEncoderOutput
		case "-decoders":
			return mockDecoderOutput
		case "-hwaccels":
			return mockHWAccelOutput
		default:
			return ""
		}
	})

	accelerators := ProbeAccelerators("ffmpeg")
	require.Len(t, accelerators, 4)

	byID := make(map[AcceleratorID]AcceleratorInfo, len(accelerators))
	for _, a := range accelerators {
		byID[a.ID] = a
	}
	assert.Contains(t, byID[AcceleratorQSV].EncodeCodecs, "h264_qsv")
	assert.Contains(t, byID[AcceleratorNVENC].EncodeCodecs, "h264_nvenc")
}

func TestParseCodecSet(t *testing.T) {
	found := parseCodecSet(mockEncoderOutput)
	assert.True(t, found["h264_qsv"])
	assert.True(t, found["h264_nvenc"])
	assert.False(t, found["libx264"])
}

func TestRecommendAccelerator_Priority(t *testing.T) {
	accelerators := []AcceleratorInfo{
		{ID: AcceleratorNVENC, Available: true},
		{ID: AcceleratorQSV, Available: true},
		{ID: AcceleratorVAAPI, Available: true},
	}
	assert.Equal(t, AcceleratorQSV, RecommendAccelerator(accelerators))
}

func TestRecommendAccelerator_None(t *testing.T) {
	assert.Empty(t, RecommendAccelerator(nil))
}

func TestDetectSystemWithFFmpeg_IncludesAccelerators(t *testing.T) {
	withMockFFmpegProbe(t, func(_ context.Context, _ string, flag string) string {
		switch flag {
		case "-encoders":
			return mockEncoderOutput
		case "-decoders":
			return mockDecoderOutput
		case "-hwaccels":
			return mockHWAccelOutput
		default:
			return ""
		}
	})

	profile := DetectSystemWithFFmpeg(nil, "ffmpeg")
	require.Len(t, profile.Accelerators, 4)
	assert.Equal(t, AcceleratorQSV, profile.RecommendedAccelerator)
}

func TestParseHWAccels(t *testing.T) {
	found := parseHWAccels(mockHWAccelOutput)
	assert.True(t, found["qsv"])
	assert.True(t, found["cuda"])
}
