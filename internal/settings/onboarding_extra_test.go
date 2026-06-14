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

func TestOnboardingCompletedAndSetupBranches(t *testing.T) {
	ctx := context.Background()

	t.Run("completed status", func(t *testing.T) {
		d := openSettingsTestDB(t)
		repo := db.NewSettingsRepo(d.Querier())
		svc := settings.NewService(repo)
		require.NoError(t, repo.Set(ctx, settings.KeyOnboardingCompleted, "true"))
		onb := settings.NewOnboarding(repo, svc, fakeSetup{required: false})
		state, err := onb.Status(ctx)
		require.NoError(t, err)
		assert.True(t, state.Completed)
		assert.Equal(t, settings.PhaseComplete, state.Phase)
	})

	t.Run("setup required redirects to admin", func(t *testing.T) {
		d := openSettingsTestDB(t)
		repo := db.NewSettingsRepo(d.Querier())
		svc := settings.NewService(repo)
		require.NoError(t, repo.Set(ctx, settings.KeyOnboardingPhase, settings.PhaseLibrary))
		onb := settings.NewOnboarding(repo, svc, fakeSetup{required: true})
		state, err := onb.Status(ctx)
		require.NoError(t, err)
		assert.Equal(t, settings.PhaseAdmin, state.Phase)
	})

	t.Run("defaults without apply", func(t *testing.T) {
		d := openSettingsTestDB(t)
		repo := db.NewSettingsRepo(d.Querier())
		svc := settings.NewService(repo)
		require.NoError(t, repo.Set(ctx, settings.KeyOnboardingPhase, settings.PhaseDefaults))
		onb := settings.NewOnboarding(repo, svc, fakeSetup{required: false})
		state, err := onb.AdvanceStep(ctx, settings.StepRequest{Phase: settings.PhaseDefaults, ApplyDefaults: false})
		require.NoError(t, err)
		assert.Equal(t, settings.PhaseScan, state.Phase)
	})

	t.Run("library step advances to defaults", func(t *testing.T) {
		d := openSettingsTestDB(t)
		repo := db.NewSettingsRepo(d.Querier())
		svc := settings.NewService(repo)
		require.NoError(t, repo.Set(ctx, settings.KeyOnboardingPhase, settings.PhaseLibrary))
		onb := settings.NewOnboarding(repo, svc, fakeSetup{required: false})
		state, err := onb.AdvanceStep(ctx, settings.StepRequest{Phase: settings.PhaseLibrary, LibraryID: 1})
		require.NoError(t, err)
		assert.Equal(t, settings.PhaseDefaults, state.Phase)
	})

	t.Run("library step is idempotent when already past defaults", func(t *testing.T) {
		d := openSettingsTestDB(t)
		repo := db.NewSettingsRepo(d.Querier())
		svc := settings.NewService(repo)
		require.NoError(t, repo.Set(ctx, settings.KeyOnboardingPhase, settings.PhaseScan))
		onb := settings.NewOnboarding(repo, svc, fakeSetup{required: false})
		state, err := onb.AdvanceStep(ctx, settings.StepRequest{Phase: settings.PhaseLibrary, LibraryID: 1})
		require.NoError(t, err)
		assert.Equal(t, settings.PhaseScan, state.Phase)
	})
}

func TestValidatePathsMissingDir(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist", "nested")
	paths := settings.ValidatePaths([]string{missing})
	require.Len(t, paths, 1)
	assert.False(t, paths[0].Readable)
}

func TestSettingsServiceSubtitlesCommaList(t *testing.T) {
	ctx := context.Background()
	d := openSettingsTestDB(t)
	svc := settings.NewService(db.NewSettingsRepo(d.Querier()))

	subs, err := svc.PatchCategory(ctx, settings.CategorySubtitles, map[string]any{
		"preferredLanguages": "en, tr",
	})
	require.NoError(t, err)
	langs := subs["preferredLanguages"].([]string)
	assert.Equal(t, []string{"en", "tr"}, langs)
}

func TestSettingsServiceInvalidSubtitlesPatch(t *testing.T) {
	ctx := context.Background()
	d := openSettingsTestDB(t)
	svc := settings.NewService(db.NewSettingsRepo(d.Querier()))

	_, err := svc.PatchCategory(ctx, settings.CategorySubtitles, map[string]any{
		"preferredLanguages": 123,
	})
	require.Error(t, err)
}
