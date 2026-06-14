package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/platform/db"
	"github.com/somralab/somra-media/internal/settings"
)

func TestSystemDetectHandler(t *testing.T) {
	d := openTestDB(t)
	settingsRepo := db.NewSettingsRepo(d.Querier())
	settingsSvc := settings.NewService(settingsRepo)
	dir := t.TempDir()

	h := testRouterWithAuth(New(Options{
		SystemHandlers: &SystemHandlers{DataDir: dir, CacheDir: dir},
		SettingsHandlers: &SettingsHandlers{Service: settingsSvc},
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/detect?paths="+dir, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestOnboardingCompleteUnauthenticated(t *testing.T) {
	d := openTestDB(t)
	settingsRepo := db.NewSettingsRepo(d.Querier())
	settingsSvc := settings.NewService(settingsRepo)
	onb := settings.NewOnboarding(settingsRepo, settingsSvc, fakeSetupChecker{})

	h := testRouterWithAuth(New(Options{
		OnboardingHandlers: &OnboardingHandlers{Onboarding: onb},
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/onboarding/complete", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNoContent, rec.Code)
}

type fakeSetupChecker struct{}

func (fakeSetupChecker) SetupRequired(context.Context) (bool, error) { return false, nil }
