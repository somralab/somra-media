package settings_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/platform/db"
	"github.com/somralab/somra-media/internal/settings"
)

func TestOnboardingStatusSetupError(t *testing.T) {
	ctx := context.Background()
	d := openSettingsTestDB(t)
	repo := db.NewSettingsRepo(d.Querier())
	svc := settings.NewService(repo)
	onb := settings.NewOnboarding(repo, svc, errSetupChecker{})
	_, err := onb.Status(ctx)
	require.Error(t, err)
}

type errSetupChecker struct{}

func (errSetupChecker) SetupRequired(context.Context) (bool, error) {
	return false, errors.New("setup check failed")
}

func TestOnboardingStatusWithoutSetup(t *testing.T) {
	ctx := context.Background()
	d := openSettingsTestDB(t)
	repo := db.NewSettingsRepo(d.Querier())
	svc := settings.NewService(repo)
	onb := settings.NewOnboarding(repo, svc, fakeSetup{required: false})

	require.NoError(t, repo.Set(ctx, settings.KeyOnboardingPhase, settings.PhaseLanguage))
	state, err := onb.Status(ctx)
	require.NoError(t, err)
	assert.Equal(t, settings.PhaseLibrary, state.Phase)
}

func TestSettingsAutoDownloadAndAPIKey(t *testing.T) {
	ctx := context.Background()
	d := openSettingsTestDB(t)
	svc := settings.NewService(db.NewSettingsRepo(d.Querier()))
	_, err := svc.PatchCategory(ctx, settings.CategorySubtitles, map[string]any{"autoDownload": true})
	require.NoError(t, err)

	enabled, err := svc.AutoDownloadSubtitles(ctx)
	require.NoError(t, err)
	assert.True(t, enabled)

	_, err = svc.PatchCategory(ctx, settings.CategorySubtitles, map[string]any{"apiKey": ""})
	require.NoError(t, err)

	_, err = svc.PatchCategory(ctx, settings.CategorySubtitles, map[string]any{
		"preferredLanguages": []any{"a"},
	})
	require.Error(t, err)
}
