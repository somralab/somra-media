package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
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
		SystemHandlers:   &SystemHandlers{DataDir: dir, CacheDir: dir},
		SettingsHandlers: &SettingsHandlers{Service: settingsSvc},
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/detect?paths="+dir, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestSystemDetectHandler_ReturnsAccelerators(t *testing.T) {
	dir := t.TempDir()
	h := testRouterWithAuth(New(Options{
		SystemHandlers: &SystemHandlers{DataDir: dir, CacheDir: dir, FFmpegBin: "ffmpeg"},
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/detect", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var profile settings.SystemProfile
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &profile))
	assert.NotEmpty(t, profile.Accelerators)
	assert.Greater(t, profile.CPUCores, 0)
}

func TestSettingsGetAllUnauthorized(t *testing.T) {
	d := openTestDB(t)
	svc := newTestAuthService(t, d)
	h := New(Options{
		AuthMiddleware:   &AuthMiddleware{Service: svc},
		SettingsHandlers: &SettingsHandlers{Service: settings.NewService(db.NewSettingsRepo(d.Querier()))},
	})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/settings", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestOnboardingCompleteUnauthenticated(t *testing.T) {
	d := openTestDB(t)
	svc := newTestAuthService(t, d)
	settingsRepo := db.NewSettingsRepo(d.Querier())
	settingsSvc := settings.NewService(settingsRepo)
	onb := settings.NewOnboarding(settingsRepo, settingsSvc, fakeSetupChecker{})

	h := New(Options{
		AuthMiddleware:     &AuthMiddleware{Service: svc},
		OnboardingHandlers: &OnboardingHandlers{Onboarding: onb},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/onboarding/complete", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

type fakeSetupChecker struct{}

func (fakeSetupChecker) SetupRequired(context.Context) (bool, error) { return false, nil }
