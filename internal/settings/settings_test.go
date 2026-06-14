package settings_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/platform/db"
	"github.com/somralab/somra-media/internal/settings"
)

func TestRecommendDefaults(t *testing.T) {
	def := settings.RecommendDefaults(2, "tr-TR")
	assert.Equal(t, 1, def.MaxConcurrentTranscodes)
	assert.Equal(t, "0 3 * * *", def.ScanCron)
	assert.Equal(t, "tr-TR", def.DefaultLocale)

	def = settings.RecommendDefaults(8, "")
	assert.Equal(t, 2, def.MaxConcurrentTranscodes)
}

func TestDetectSystem(t *testing.T) {
	dir := t.TempDir()
	profile := settings.DetectSystem([]string{dir})
	assert.Greater(t, profile.CPUCores, 0)
	require.Len(t, profile.Paths, 1)
	assert.True(t, profile.Paths[0].Readable)
	assert.True(t, profile.Paths[0].Writable)
}

func TestSettingsServicePatchAndOnboarding(t *testing.T) {
	ctx := context.Background()
	dataDir := filepath.Join(t.TempDir(), "data")
	cfg := db.Default()
	cfg.DataDir = dataDir
	d, err := db.Initialize(ctx, cfg, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = d.Close() })

	repo := db.NewSettingsRepo(d.Querier())
	svc := settings.NewService(repo)
	onb := settings.NewOnboarding(repo, svc, fakeSetup{required: true})

	state, err := onb.Status(ctx)
	require.NoError(t, err)
	assert.Equal(t, settings.PhaseLanguage, state.Phase)
	assert.True(t, state.SetupRequired)
	assert.False(t, state.Completed)

	state, err = onb.AdvanceStep(ctx, settings.StepRequest{Phase: settings.PhaseLanguage, Locale: "en-US"})
	require.NoError(t, err)
	assert.Equal(t, settings.PhaseAdmin, state.Phase)

	locale, err := svc.GetString(ctx, settings.KeyDefaultLocale, "")
	require.NoError(t, err)
	assert.Equal(t, "en-US", locale)

	def := settings.RecommendDefaults(4, "en-US")
	require.NoError(t, svc.ApplySmartDefaults(ctx, def))
	snap, err := svc.GetAll(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, snap["playback"]["maxConcurrentTranscodes"])
}

type fakeSetup struct {
	required bool
}

func (f fakeSetup) SetupRequired(context.Context) (bool, error) { return f.required, nil }
