package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/platform/db"
	"github.com/somralab/somra-media/internal/settings"
)

func TestSettingsAndOnboardingHandlers(t *testing.T) {
	d := openTestDB(t)
	svc := newTestAuthService(t, d)
	settingsRepo := db.NewSettingsRepo(d.Querier())
	settingsSvc := settings.NewService(settingsRepo)
	onb := settings.NewOnboarding(settingsRepo, settingsSvc, svc)

	h := testRouterWithAuth(New(Options{
		AuthHandlers:       &AuthHandlers{Service: svc, Onboarding: onb},
		SystemHandlers:     &SystemHandlers{DataDir: t.TempDir(), CacheDir: t.TempDir()},
		OnboardingHandlers: &OnboardingHandlers{Onboarding: onb},
		SettingsHandlers:   &SettingsHandlers{Service: settingsSvc},
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/setup/status", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	var status map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &status))
	assert.Equal(t, true, status["setupRequired"])
	assert.Equal(t, "language", status["phase"])

	req = httptest.NewRequest(http.MethodGet, "/api/v1/system/detect", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	stepBody, _ := json.Marshal(map[string]string{"phase": "language", "locale": "tr-TR"})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/onboarding/step", bytes.NewReader(stepBody))
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	setupBody, _ := json.Marshal(map[string]string{"username": "admin", "password": "AdminPass1"})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/setup/admin", bytes.NewReader(setupBody))
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var tokenResp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &tokenResp))
	token := tokenResp["accessToken"].(string)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/settings", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestSettingsServiceIntegration(t *testing.T) {
	d := openTestDB(t)
	repo := db.NewSettingsRepo(d.Querier())
	svc := settings.NewService(repo)
	_, err := svc.PatchCategory(t.Context(), settings.CategoryGeneral, map[string]any{"defaultLocale": "tr-TR"})
	require.NoError(t, err)
	snap, err := svc.GetAll(t.Context())
	require.NoError(t, err)
	assert.Equal(t, "tr-TR", snap["general"]["defaultLocale"])
	_ = filepath.Join(t.TempDir(), "unused")
}
