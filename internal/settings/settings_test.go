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

func TestSettingsServiceAllCategories(t *testing.T) {
	ctx := context.Background()
	d := openSettingsTestDB(t)
	repo := db.NewSettingsRepo(d.Querier())
	svc := settings.NewService(repo)

	snap, err := svc.GetAll(ctx)
	require.NoError(t, err)
	assert.Equal(t, "en-US", snap["general"]["defaultLocale"])
	assert.Equal(t, 2, snap["playback"]["maxConcurrentTranscodes"])

	lib, err := svc.PatchCategory(ctx, settings.CategoryLibrary, map[string]any{"scanCron": "0 4 * * *"})
	require.NoError(t, err)
	assert.Equal(t, "0 4 * * *", lib["scanCron"])

	playback, err := svc.PatchCategory(ctx, settings.CategoryPlayback, map[string]any{"maxConcurrentTranscodes": 3})
	require.NoError(t, err)
	assert.Equal(t, 3, playback["maxConcurrentTranscodes"])

	streaming, err := svc.PatchCategory(ctx, settings.CategoryStreaming, map[string]any{"maxConcurrentTranscodes": 4})
	require.NoError(t, err)
	assert.Equal(t, 4, streaming["maxConcurrentTranscodes"])

	subs, err := svc.PatchCategory(ctx, settings.CategorySubtitles, map[string]any{
		"autoDownload":       true,
		"preferredLanguages": []any{"en", "tr"},
		"apiKey":             "test-key",
	})
	require.NoError(t, err)
	assert.Equal(t, true, subs["autoDownload"])
	assert.Equal(t, true, subs["apiKeySet"])

	auto, err := svc.AutoDownloadSubtitles(ctx)
	require.NoError(t, err)
	assert.True(t, auto)

	langs, err := svc.PreferredLanguages(ctx)
	require.NoError(t, err)
	assert.Equal(t, []string{"en", "tr"}, langs)

	key, err := svc.OpenSubtitlesAPIKey(ctx)
	require.NoError(t, err)
	assert.Equal(t, "test-key", key)

	_, err = svc.PatchCategory(ctx, "unknown", map[string]any{})
	require.Error(t, err)

	_, err = svc.PatchCategory(ctx, settings.CategoryGeneral, map[string]any{"defaultLocale": "fr-FR"})
	require.Error(t, err)

	_, err = svc.PatchCategory(ctx, settings.CategoryLibrary, map[string]any{"scanCron": ""})
	require.Error(t, err)

	_, err = svc.PatchCategory(ctx, settings.CategoryPlayback, map[string]any{"maxConcurrentTranscodes": 0})
	require.Error(t, err)

	err = svc.ApplySmartDefaults(ctx, settings.SmartDefaults{MaxConcurrentTranscodes: 0})
	require.Error(t, err)
}

func TestOnboardingAdvanceAndComplete(t *testing.T) {
	ctx := context.Background()
	d := openSettingsTestDB(t)
	repo := db.NewSettingsRepo(d.Querier())
	svc := settings.NewService(repo)
	onb := settings.NewOnboarding(repo, svc, fakeSetup{required: false})

	require.NoError(t, onb.AfterAdminCreated(ctx))
	state, err := onb.Status(ctx)
	require.NoError(t, err)
	assert.Equal(t, settings.PhaseLibrary, state.Phase)

	require.NoError(t, onb.AfterLibraryCreated(ctx))
	state, err = onb.Status(ctx)
	require.NoError(t, err)
	assert.Equal(t, settings.PhaseDefaults, state.Phase)
	require.NotNil(t, state.SmartDefaults)

	state, err = onb.AdvanceStep(ctx, settings.StepRequest{Phase: settings.PhaseDefaults, ApplyDefaults: true})
	require.NoError(t, err)
	assert.Equal(t, settings.PhaseScan, state.Phase)

	state, err = onb.AdvanceStep(ctx, settings.StepRequest{Phase: settings.PhaseScan})
	require.NoError(t, err)
	assert.Equal(t, settings.PhaseComplete, state.Phase)

	require.NoError(t, onb.Complete(ctx))
	state, err = onb.Status(ctx)
	require.NoError(t, err)
	assert.True(t, state.Completed)
	assert.Equal(t, settings.PhaseComplete, state.Phase)

	_, err = onb.AdvanceStep(ctx, settings.StepRequest{Phase: settings.PhaseLanguage, Locale: "bad"})
	require.Error(t, err)
}

func TestValidatePathsAndDetect(t *testing.T) {
	dir := t.TempDir()
	paths := settings.ValidatePaths([]string{dir, ""})
	require.Len(t, paths, 1)
	assert.True(t, paths[0].Readable)
	assert.True(t, paths[0].Writable)

	profile := settings.DetectSystem([]string{dir})
	assert.Greater(t, profile.CPUCores, 0)
}

func openSettingsTestDB(t *testing.T) *db.DB {
	t.Helper()
	cfg := db.Default()
	cfg.DataDir = filepath.Join(t.TempDir(), "data")
	d, err := db.Initialize(context.Background(), cfg, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = d.Close() })
	return d
}
