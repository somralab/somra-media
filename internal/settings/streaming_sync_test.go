package settings

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/platform/db"
)

func TestGetStreamingRuntimeConfig_Defaults(t *testing.T) {
	ctx := t.Context()
	d := openStreamingSyncTestDB(t)
	defer d.Close()
	svc := NewService(db.NewSettingsRepo(d.Querier()))

	cfg, err := svc.GetStreamingRuntimeConfig(ctx, "ffmpeg")
	require.NoError(t, err)
	assert.Equal(t, 2, cfg.MaxConcurrentTranscodes)
	assert.Equal(t, HWModeAuto, cfg.HWMode)
}

func TestPatchPlayback_HWSettings(t *testing.T) {
	ctx := t.Context()
	d := openStreamingSyncTestDB(t)
	defer d.Close()
	svc := NewService(db.NewSettingsRepo(d.Querier()))

	out, err := svc.PatchCategory(ctx, CategoryPlayback, map[string]any{
		"hwMode":          "force",
		"hwAccelerator":   "qsv",
		"maxHWTranscodes": 3,
	})
	require.NoError(t, err)
	assert.Equal(t, "force", out["hwMode"])
	assert.Equal(t, "qsv", out["hwAccelerator"])
	assert.Equal(t, 3, out["maxHWTranscodes"])
}

func TestRecommendDefaultsWithProfile_GPU(t *testing.T) {
	profile := SystemProfile{RecommendedAccelerator: AcceleratorQSV}
	def := RecommendDefaultsWithProfile(8, "en-US", profile)
	assert.Equal(t, HWModeAuto, def.HWMode)
	assert.Equal(t, AcceleratorQSV, def.RecommendedAccelerator)
	assert.Equal(t, 3, def.MaxHWTranscodes)
}

func openStreamingSyncTestDB(t *testing.T) *db.DB {
	t.Helper()
	dataDir := filepath.Join(t.TempDir(), "data")
	cfg := db.Default()
	cfg.DataDir = dataDir
	d, err := db.Initialize(t.Context(), cfg, nil)
	require.NoError(t, err)
	return d
}
