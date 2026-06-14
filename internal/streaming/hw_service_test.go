package streaming

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyRuntimeSettings(t *testing.T) {
	svc := NewService(ServiceConfig{CacheDir: t.TempDir(), MaxConcurrent: 2}, nil, nil, nil)
	svc.ApplyRuntimeSettings(HWRuntimeConfig{
		Mode: HWModeForce, Preferred: AccelQSV,
		Available:     []Accelerator{AccelQSV, AccelNVENC},
		MaxHWSessions: 3, MaxTotalSessions: 4, VAAPIDevice: "/dev/dri/renderD129",
	})
	cfg := svc.currentHWConfig()
	assert.Equal(t, HWModeForce, cfg.Mode)
	assert.Equal(t, AccelQSV, cfg.Preferred)
	assert.Equal(t, 3, cfg.MaxHWSessions)
	assert.Equal(t, 4, svc.cfg.MaxConcurrent)
}

func TestHWActiveTracking(t *testing.T) {
	svc := NewService(ServiceConfig{CacheDir: t.TempDir()}, nil, nil, nil)
	svc.incHWActive()
	svc.incHWActive()
	assert.Equal(t, 2, svc.activeHWCount())
	assert.Equal(t, int64(2), svc.Metrics().HWActiveSessions())
	svc.decHWActive()
	assert.Equal(t, 1, svc.activeHWCount())
}

func TestMetricsHWCounters(t *testing.T) {
	m := NewMetrics()
	m.incHWStarts()
	m.incHWErrors()
	m.incHWFallbacks()
	assert.Equal(t, int64(1), m.HWStarts())
	assert.Equal(t, int64(1), m.HWErrors())
	assert.Equal(t, int64(1), m.HWFallbacks())
}

func TestStartTranscodeWithFallback_HWFailsToSW(t *testing.T) {
	svc := NewService(ServiceConfig{CacheDir: t.TempDir()}, nil, nil, nil)
	opts := PackagerOptions{
		SourcePath: filepath.Join(t.TempDir(), "missing.mkv"),
		OutputDir:  filepath.Join(t.TempDir(), "out"),
		Mode:       ModeTranscode,
		Tiers:      []LadderTier{{VideoBitrate: 1_000_000, AudioBitrate: 128_000}},
		TranscodePath: TranscodePath{
			UseHW: true, Accelerator: AccelQSV, VideoEncoder: "h264_qsv",
		},
	}
	pm := NewProcessManager(ProcessManagerConfig{MaxConcurrent: 1, FFmpegBin: "/bin/false"})
	svc.procMgr = pm
	err := svc.startTranscodeWithFallback(context.Background(), "sess-hw", opts, opts.TranscodePath)
	assert.Error(t, err)
	assert.Equal(t, int64(1), svc.Metrics().HWFallbacks())
}

func TestStartTranscodeWithFallback_SWOnly(t *testing.T) {
	svc := NewService(ServiceConfig{CacheDir: t.TempDir()}, nil, nil, nil)
	opts := PackagerOptions{
		SourcePath:    filepath.Join(t.TempDir(), "missing.mkv"),
		OutputDir:     filepath.Join(t.TempDir(), "out"),
		Mode:          ModeTranscode,
		Tiers:         []LadderTier{{VideoBitrate: 1_000_000}},
		TranscodePath: TranscodePath{UseHW: false, VideoEncoder: "libx264"},
	}
	svc.procMgr = NewProcessManager(ProcessManagerConfig{MaxConcurrent: 1, FFmpegBin: "/bin/false"})
	err := svc.startTranscodeWithFallback(context.Background(), "sess-sw", opts, opts.TranscodePath)
	assert.Error(t, err)
	assert.Equal(t, int64(0), svc.Metrics().HWFallbacks())
}

func TestAppendHWVideoArgs_AllAccelerators(t *testing.T) {
	tier := LadderTier{Width: 640, Height: 360, VideoBitrate: 900_000}
	for _, acc := range []Accelerator{AccelNVENC, AccelVAAPI, AccelAMF} {
		args := AppendHWVideoArgs(nil, TranscodePath{
			UseHW: true, Accelerator: acc, VideoEncoder: hwEncoderFor(acc),
		}, tier, "/dev/dri/renderD128")
		assert.Contains(t, args, hwEncoderFor(acc))
	}
}

func TestProcessManager_SetMaxConcurrent(t *testing.T) {
	pm := NewProcessManager(ProcessManagerConfig{MaxConcurrent: 1})
	pm.SetMaxConcurrent(3)
	require.NoError(t, pm.Acquire(context.Background()))
	require.NoError(t, pm.Acquire(context.Background()))
	pm.Release()
	pm.Release()
}

func TestSelectTranscodePath_ForceUnavailable(t *testing.T) {
	path := SelectTranscodePath(HWRuntimeConfig{Mode: HWModeForce, Available: nil}, MediaProbe{}, 0)
	assert.False(t, path.UseHW)
	assert.True(t, path.FallbackToSW)
}
