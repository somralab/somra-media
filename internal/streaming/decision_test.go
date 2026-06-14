package streaming

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecisionEngine_Matrix(t *testing.T) {
	t.Parallel()
	eng := NewDecisionEngine()
	caps := DefaultBrowserCapabilities()

	tests := []struct {
		name   string
		caps   ClientCapabilities
		probe  MediaProbe
		want   Mode
		reason string
	}{
		{
			name: "direct_play_h264_aac_mp4",
			caps: caps,
			probe: MediaProbe{
				Container: "mp4", VideoCodec: "h264", AudioCodec: "aac",
				VideoWidth: 1920, VideoHeight: 1080, AudioChannels: 2,
			},
			want: ModeDirectPlay, reason: "browser_safe_container",
		},
		{
			name: "direct_stream_mkv_h264",
			caps: caps,
			probe: MediaProbe{
				Container: "matroska,webm", VideoCodec: "h264", AudioCodec: "aac",
				VideoWidth: 1280, VideoHeight: 720, AudioChannels: 2,
			},
			want: ModeDirectStream, reason: "remux_required",
		},
		{
			name: "transcode_hevc",
			caps: caps,
			probe: MediaProbe{
				Container: "mp4", VideoCodec: "hevc", AudioCodec: "aac",
				VideoWidth: 1920, VideoHeight: 1080, AudioChannels: 2,
			},
			want: ModeTranscode, reason: "unsupported_codec",
		},
		{
			name: "transcode_ac3",
			caps: caps,
			probe: MediaProbe{
				Container: "mp4", VideoCodec: "h264", AudioCodec: "ac3",
				VideoWidth: 1280, VideoHeight: 720, AudioChannels: 6,
			},
			want: ModeTranscode, reason: "unsupported_codec",
		},
		{
			name: "transcode_high_bitrate",
			caps: ClientCapabilities{
				VideoCodecs: []string{"h264"}, AudioCodecs: []string{"aac"},
				Containers: []string{"mp4"}, MaxBitrate: 1_000_000, MaxAudioChannels: 2,
			},
			probe: MediaProbe{
				Container: "mp4", VideoCodec: "h264", AudioCodec: "aac",
				VideoWidth: 3840, VideoHeight: 2160, AudioChannels: 2, Bitrate: 15_000_000,
			},
			want: ModeTranscode, reason: "bitrate_exceeds_client",
		},
		{
			name: "direct_play_no_unnecessary_transcode",
			caps: caps,
			probe: MediaProbe{
				Container: "mov", VideoCodec: "avc1", AudioCodec: "mp4a",
				VideoWidth: 1280, VideoHeight: 720, AudioChannels: 2,
			},
			want: ModeDirectPlay, reason: "browser_safe_container",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := eng.Decide(tc.caps, tc.probe)
			assert.Equal(t, tc.want, got.Mode)
			assert.Equal(t, tc.reason, got.Reason)
		})
	}
}

func TestBuildLadder(t *testing.T) {
	t.Parallel()
	tiers := BuildLadder(1920, 1080)
	assert.GreaterOrEqual(t, len(tiers), 2)
	assert.Equal(t, 1920, tiers[0].Width)
}

func TestWriteMasterPlaylist(t *testing.T) {
	t.Parallel()
	out := WriteMasterPlaylist([]VariantInfo{{
		Name: "720p", Bandwidth: 2_500_000, Width: 1280, Height: 720, MediaPlaylist: "v0.m3u8",
	}})
	assert.Contains(t, out, "#EXTM3U")
	assert.Contains(t, out, "v0.m3u8")
}

func TestValidateCacheSegmentPath(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	session := "abc-123"
	dir := SessionCacheDir(root, session)
	assert.NoError(t, EnsureOutputDir(dir))

	ok, err := ValidateCacheSegmentPath(root, session, InitSegmentName)
	assert.NoError(t, err)
	assert.Contains(t, ok, InitSegmentName)

	_, err = ValidateCacheSegmentPath(root, session, "../../../etc/passwd")
	assert.Error(t, err)
}

func TestBuildFFmpegArgs(t *testing.T) {
	t.Parallel()
	args := BuildFFmpegArgs(PackagerOptions{
		SourcePath: "/media/movie.mkv", OutputDir: "/cache/s1",
		Mode: ModeTranscode, Tiers: BuildLadder(1920, 1080),
		TranscodePath: TranscodePath{UseHW: false, VideoEncoder: "libx264"},
	})
	assert.Contains(t, args, "-c:v")
	assert.Contains(t, args, "libx264")
}

func TestNormalizeCodec(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "h264", normalizeCodec("avc1"))
	assert.Equal(t, "hevc", normalizeCodec("h265"))
	assert.True(t, containsCodec([]string{"h264"}, "avc1"))
}
