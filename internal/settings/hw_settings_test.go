package settings

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/platform/db"
	"github.com/somralab/somra-media/internal/streaming"
)

func TestToStreamingHWConfig(t *testing.T) {
	cfg := StreamingRuntimeConfig{
		MaxConcurrentTranscodes: 4,
		MaxHWTranscodes:         2,
		HWMode:                  HWModeAuto,
		HWAccelerator:           "auto",
		AvailableAccelerators:     []AcceleratorID{AcceleratorQSV, AcceleratorNVENC},
	}
	hw := cfg.ToStreamingHWConfig()
	assert.Equal(t, streaming.HWModeAuto, hw.Mode)
	assert.Equal(t, streaming.AccelNone, hw.Preferred)
	assert.Len(t, hw.Available, 2)
	assert.Equal(t, 2, hw.MaxHWSessions)
}

func TestToStreamingHWConfig_Preferred(t *testing.T) {
	cfg := StreamingRuntimeConfig{HWAccelerator: AcceleratorNVENC, AvailableAccelerators: []AcceleratorID{AcceleratorNVENC}}
	hw := cfg.ToStreamingHWConfig()
	assert.Equal(t, streaming.AccelNVENC, hw.Preferred)
}

func TestGetAll_IncludesHWPlayback(t *testing.T) {
	ctx := t.Context()
	d := openStreamingSyncTestDB(t)
	defer d.Close()
	svc := NewService(db.NewSettingsRepo(d.Querier()))

	snap, err := svc.GetAll(ctx)
	require.NoError(t, err)
	playback := snap[CategoryPlayback]
	assert.Equal(t, "auto", playback["hwMode"])
	assert.NotNil(t, playback["availableAccelerators"])
}

func TestRecommendAccelerator_DevicePresentFirst(t *testing.T) {
	accelerators := []AcceleratorInfo{
		{ID: AcceleratorNVENC, Available: true, DevicePresent: true},
		{ID: AcceleratorQSV, Available: true, DevicePresent: false},
	}
	assert.Equal(t, AcceleratorNVENC, RecommendAccelerator(accelerators))
}

func TestProbeAccelerators_EmptyFFmpeg(t *testing.T) {
	orig := ffmpegProbeFn
	ffmpegProbeFn = func(context.Context, string, string) string { return "" }
	t.Cleanup(func() { ffmpegProbeFn = orig })
	accelerators := ProbeAccelerators("ffmpeg")
	assert.Len(t, accelerators, 4)
	for _, a := range accelerators {
		assert.False(t, a.Available)
	}
}

func TestPatchPlayback_InvalidHWMode(t *testing.T) {
	ctx := t.Context()
	d := openStreamingSyncTestDB(t)
	defer d.Close()
	svc := NewService(db.NewSettingsRepo(d.Querier()))
	_, err := svc.PatchCategory(ctx, CategoryPlayback, map[string]any{"hwMode": "invalid"})
	assert.Error(t, err)
}

func TestPatchPlayback_InvalidAccelerator(t *testing.T) {
	ctx := t.Context()
	d := openStreamingSyncTestDB(t)
	defer d.Close()
	svc := NewService(db.NewSettingsRepo(d.Querier()))
	_, err := svc.PatchCategory(ctx, CategoryPlayback, map[string]any{"hwAccelerator": "bad"})
	assert.Error(t, err)
}

func TestApplySmartDefaults_HWFields(t *testing.T) {
	ctx := t.Context()
	d := openStreamingSyncTestDB(t)
	defer d.Close()
	svc := NewService(db.NewSettingsRepo(d.Querier()))

	err := svc.ApplySmartDefaults(ctx, SmartDefaults{
		MaxConcurrentTranscodes: 2,
		MaxHWTranscodes:         2,
		HWMode:                  HWModeAuto,
		RecommendedAccelerator:  AcceleratorQSV,
		ScanCron:                defaultScanCron,
	})
	require.NoError(t, err)

	playback, err := svc.getPlayback(ctx)
	require.NoError(t, err)
	assert.Equal(t, "auto", playback["hwMode"])
	assert.Equal(t, "qsv", playback["hwAccelerator"])
}
