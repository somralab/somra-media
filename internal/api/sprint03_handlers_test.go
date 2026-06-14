package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/auth"
	"github.com/somralab/somra-media/internal/platform/db"
)

func newSprint03Router(t *testing.T) (http.Handler, *auth.Service, *db.DB, string) {
	t.Helper()
	d := openTestDB(t)
	svc := newTestAuthService(t, d)
	h := New(Options{
		AuthHandlers:    &AuthHandlers{Service: svc},
		AuthMiddleware:  &AuthMiddleware{Service: svc},
		UserHandlers:    &UserHandlers{Service: svc, Users: db.NewUserRepo(d.Querier())},
		ProfileHandlers: &ProfileHandlers{Profiles: db.NewProfileRepo(d.Querier())},
		WatchHandlers:   &WatchHandlers{Watch: db.NewWatchRepo(d.Querier())},
		BrowseHandlers: &BrowseHandlers{
			Browse: db.NewBrowseRepo(d.Querier()),
			Locale: func(*http.Request) string { return "en-US" },
		},
	})

	setupBody, _ := json.Marshal(map[string]string{"username": "admin", "password": "AdminPass1"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/setup/admin", bytes.NewReader(setupBody))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var tok map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &tok))
	access := tok["accessToken"].(string)
	return h, svc, d, access
}

func authRequest(method, path, token string, body []byte) *http.Request {
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req
}

func TestAuthHandlers_SessionsAndLoginErrors(t *testing.T) {
	h, _, _, access := newSprint03Router(t)

	req := authRequest(http.MethodGet, "/api/v1/auth/sessions", access, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var sessions []map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &sessions))
	require.NotEmpty(t, sessions)
	sessionID := sessions[0]["id"].(string)

	req = authRequest(http.MethodDelete, "/api/v1/auth/sessions/"+sessionID, access, nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNoContent, rec.Code)

	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader([]byte(`{`)))
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)

	badLogin, _ := json.Marshal(map[string]string{"username": "admin", "password": "wrong"})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(badLogin))
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusUnauthorized, rec.Code)

	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusUnauthorized, rec.Code)

	req = httptest.NewRequest(http.MethodPost, "/api/v1/setup/admin", bytes.NewReader([]byte(`{"username":"x","password":"AdminPass1"}`)))
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusConflict, rec.Code)
}

func TestAuthHandlers_RefreshFromBody(t *testing.T) {
	d := openTestDB(t)
	svc := newTestAuthService(t, d)
	h := New(Options{
		AuthHandlers:   &AuthHandlers{Service: svc},
		AuthMiddleware: &AuthMiddleware{Service: svc},
	})

	setupBody, _ := json.Marshal(map[string]string{"username": "admin", "password": "AdminPass1"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/setup/admin", bytes.NewReader(setupBody))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)
	cookie := rec.Result().Cookies()[0]

	body, _ := json.Marshal(map[string]string{"refreshToken": cookie.Value})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(body))
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestUserHandlers_CRUD(t *testing.T) {
	h, _, _, access := newSprint03Router(t)

	req := authRequest(http.MethodGet, "/api/v1/users", access, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	createBody, _ := json.Marshal(map[string]any{
		"username": "member",
		"password": "MemberPass1",
		"roles":    []string{"user"},
	})
	req = authRequest(http.MethodPost, "/api/v1/users", access, createBody)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var created map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))
	userID := created["id"].(string)

	req = authRequest(http.MethodGet, "/api/v1/users/"+userID, access, nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	updateBody, _ := json.Marshal(map[string]any{"disabled": true, "password": "NewPass1"})
	req = authRequest(http.MethodPut, "/api/v1/users/"+userID, access, updateBody)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	dupBody, _ := json.Marshal(map[string]any{"username": "member", "password": "MemberPass1"})
	req = authRequest(http.MethodPost, "/api/v1/users", access, dupBody)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusConflict, rec.Code)

	weakBody, _ := json.Marshal(map[string]any{"username": "weak", "password": "short"})
	req = authRequest(http.MethodPost, "/api/v1/users", access, weakBody)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)

	req = authRequest(http.MethodGet, "/api/v1/users/missing-id", access, nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestProfileHandlers_GetAndUpdate(t *testing.T) {
	h, _, _, access := newSprint03Router(t)

	req := authRequest(http.MethodGet, "/api/v1/profile", access, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	updateBody, _ := json.Marshal(map[string]any{
		"locale":           "tr-TR",
		"theme":            "aurora",
		"maxContentRating": "PG-13",
		"isChild":          false,
	})
	req = authRequest(http.MethodPut, "/api/v1/profile", access, updateBody)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var profile map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &profile))
	assert.Equal(t, "tr-TR", profile["locale"])
	assert.Equal(t, "aurora", profile["theme"])
}

func TestWatchHandlers_FullFlow(t *testing.T) {
	h, _, d, access := newSprint03Router(t)
	ctx := context.Background()

	users := db.NewUserRepo(d.Querier())
	all, err := users.List(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, all)
	userID := all[0].ID

	lib := db.NewLibraryRepo(d.Querier())
	media := db.NewMediaRepo(d.Querier())
	libRec, err := lib.Create(ctx, "Films", db.LibraryKindMovie, []string{t.TempDir()}, false)
	require.NoError(t, err)
	itemID, err := media.CreateItem(ctx, libRec.ID, db.LibraryKindMovie, "Title", nil)
	require.NoError(t, err)
	_ = userID

	itemPath := fmt.Sprintf("/api/v1/watch-state/%d", itemID)
	watchBody, _ := json.Marshal(map[string]any{"positionMs": 5000, "completed": false})
	req := authRequest(http.MethodPut, itemPath, access, watchBody)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	req = authRequest(http.MethodGet, itemPath, access, nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	req = authRequest(http.MethodGet, "/api/v1/watch-state", access, nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	favPath := fmt.Sprintf("/api/v1/favorites/%d", itemID)
	req = authRequest(http.MethodPost, favPath, access, nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNoContent, rec.Code)

	req = authRequest(http.MethodGet, "/api/v1/favorites", access, nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	req = authRequest(http.MethodDelete, favPath, access, nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNoContent, rec.Code)

	wlPath := fmt.Sprintf("/api/v1/watchlist/%d", itemID)
	req = authRequest(http.MethodPost, wlPath, access, nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNoContent, rec.Code)

	req = authRequest(http.MethodGet, "/api/v1/watchlist", access, nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	req = authRequest(http.MethodDelete, wlPath, access, nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNoContent, rec.Code)

	req = authRequest(http.MethodGet, "/api/v1/watch-state/bad-id", access, nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAuthMiddleware_RequireRoleAndLocale(t *testing.T) {
	d := openTestDB(t)
	svc := newTestAuthService(t, d)
	ctx := context.Background()
	uid := uuid.NewString()
	repo := db.NewUserRepo(d.Querier())
	_, err := repo.Create(ctx, uid, "child-user", "hash", []string{auth.RoleChild})
	require.NoError(t, err)
	perms, err := repo.PermissionsForUser(ctx, uid)
	require.NoError(t, err)

	ac := auth.AuthContext{
		Claims: auth.Claims{
			Subject: auth.Subject{UserID: uid, Username: "child-user", Roles: []string{auth.RoleChild}},
		},
		Permissions: perms,
		Profile:     db.UserProfile{UserID: uid, Locale: "tr-TR", Theme: "cinematic"},
	}

	allowed := false
	handler := RequireRole(auth.RoleAdmin)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowed = true
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(auth.WithAuthContext(req.Context(), ac))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)
	assert.False(t, allowed)

	localeHandler := ProfileLocaleMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		locale, ok := AcceptLanguageFromContext(r.Context())
		require.True(t, ok)
		assert.Equal(t, "tr-TR", locale)
		assert.Equal(t, "tr-TR", r.Header.Get("Accept-Language"))
	}))
	rec = httptest.NewRecorder()
	localeHandler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	_, ok := AcceptLanguageFromContext(context.Background())
	assert.False(t, ok)

	_ = svc
}
