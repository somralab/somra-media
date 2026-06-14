package settings_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/platform/db"
	"github.com/somralab/somra-media/internal/settings"
)

func TestOnboardingAfterHooksAndErrors(t *testing.T) {
	ctx := context.Background()
	d := openSettingsTestDB(t)
	repo := db.NewSettingsRepo(d.Querier())
	svc := settings.NewService(repo)
	onb := settings.NewOnboarding(repo, svc, fakeSetup{required: false})

	require.NoError(t, onb.AfterAdminCreated(ctx))
	require.NoError(t, onb.AfterLibraryCreated(ctx))

	_, err := onb.AdvanceStep(ctx, settings.StepRequest{Phase: "bogus"})
	require.Error(t, err)

	_, err = onb.AdvanceStep(ctx, settings.StepRequest{Phase: settings.PhaseLanguage, Locale: "de-DE"})
	require.Error(t, err)
}

func TestSettingsServiceGetStringAndPlayback(t *testing.T) {
	ctx := context.Background()
	d := openSettingsTestDB(t)
	repo := db.NewSettingsRepo(d.Querier())
	svc := settings.NewService(repo)

	require.NoError(t, repo.Set(ctx, settings.KeyStreamingMaxConcurrent, "not-a-number"))
	snap, err := svc.GetAll(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, snap["playback"]["maxConcurrentTranscodes"])

	val, err := svc.GetString(ctx, "missing.key", "fallback")
	require.NoError(t, err)
	assert.Equal(t, "fallback", val)
}

func TestSettingsServicePatchGeneralNoop(t *testing.T) {
	ctx := context.Background()
	d := openSettingsTestDB(t)
	svc := settings.NewService(db.NewSettingsRepo(d.Querier()))
	out, err := svc.PatchCategory(ctx, settings.CategoryGeneral, map[string]any{})
	require.NoError(t, err)
	assert.Equal(t, "en-US", out["defaultLocale"])
}

func TestSettingsServicePlaybackIntTypes(t *testing.T) {
	ctx := context.Background()
	d := openSettingsTestDB(t)
	svc := settings.NewService(db.NewSettingsRepo(d.Querier()))
	out, err := svc.PatchCategory(ctx, settings.CategoryPlayback, map[string]any{"maxConcurrentTranscodes": float64(3)})
	require.NoError(t, err)
	assert.Equal(t, 3, out["maxConcurrentTranscodes"])

	langsOut, err := svc.PatchCategory(ctx, settings.CategorySubtitles, map[string]any{
		"preferredLanguages": []string{"en", "de"},
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"en", "de"}, langsOut["preferredLanguages"])
}

func TestOnboardingAfterAdminWhenCompleted(t *testing.T) {
	ctx := context.Background()
	d := openSettingsTestDB(t)
	repo := db.NewSettingsRepo(d.Querier())
	svc := settings.NewService(repo)
	require.NoError(t, repo.Set(ctx, settings.KeyOnboardingCompleted, "true"))
	onb := settings.NewOnboarding(repo, svc, fakeSetup{required: false})
	require.NoError(t, onb.AfterAdminCreated(ctx))
	require.NoError(t, onb.AfterLibraryCreated(ctx))
}

func TestApplySmartDefaultsMinimal(t *testing.T) {
	ctx := context.Background()
	d := openSettingsTestDB(t)
	svc := settings.NewService(db.NewSettingsRepo(d.Querier()))
	require.NoError(t, svc.ApplySmartDefaults(ctx, settings.SmartDefaults{
		MaxConcurrentTranscodes: 1,
		DefaultLocale:           "tr-TR",
	}))
}
