package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/platform/db"
)

func TestAuthHandlers_LoginRefreshLogoutFlow(t *testing.T) {
	d := openTestDB(t)
	svc := newTestAuthService(t, d)
	h := New(Options{
		AuthHandlers:    &AuthHandlers{Service: svc},
		AuthMiddleware:  &AuthMiddleware{Service: svc},
		ProfileHandlers: &ProfileHandlers{Profiles: dbNewProfileRepo(d)},
	})

	setupBody, _ := json.Marshal(map[string]string{"username": "admin", "password": "AdminPass1"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/setup/admin", bytes.NewReader(setupBody))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var setupResp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &setupResp))
	require.NotEmpty(t, setupResp["accessToken"])
	cookie := rec.Result().Cookies()[0]

	loginBody, _ := json.Marshal(map[string]string{"username": "admin", "password": "AdminPass1"})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	req.AddCookie(cookie)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	var refreshResp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &refreshResp))
	require.NotEmpty(t, refreshResp["accessToken"])
	require.NotNil(t, refreshResp["user"])
	access := refreshResp["accessToken"].(string)
	cookie = rec.Result().Cookies()[0]

	req = httptest.NewRequest(http.MethodGet, "/api/v1/profile", nil)
	req.Header.Set("Authorization", "Bearer "+access)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.AddCookie(cookie)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNoContent, rec.Code)
}

func TestAuthHandlers_UnauthorizedLibrary(t *testing.T) {
	d := openTestDB(t)
	svc := newTestAuthService(t, d)
	h := New(Options{
		AuthMiddleware:  &AuthMiddleware{Service: svc},
		LibraryHandlers: &LibraryHandlers{},
	})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/libraries", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func dbNewProfileRepo(d *db.DB) *db.ProfileRepo {
	return db.NewProfileRepo(d.Querier())
}
