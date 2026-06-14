package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/auth"
	"github.com/somralab/somra-media/internal/jobs"
	"github.com/somralab/somra-media/internal/library"
	"github.com/somralab/somra-media/internal/platform/db"
	"github.com/somralab/somra-media/internal/platform/diagnostics"
)

func TestLibraryHandlers_MountReadWrite(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	dir := t.TempDir()

	queue := jobs.NewMemoryQueue(jobs.MemoryQueueConfig{Workers: 1, Buffer: 2})
	t.Cleanup(queue.Close)
	scanner := library.NewScanner(library.ScannerConfig{DB: d, Prober: stubProber{}, Progress: library.NoopProgressPublisher{}})
	svc := library.NewService(library.ServiceConfig{DB: d, Queue: queue, Scanner: scanner})
	handlers := &LibraryHandlers{Service: svc}

	rWrite := chi.NewRouter()
	handlers.MountWrite(rWrite)

	rRead := chi.NewRouter()
	handlers.MountRead(rRead)

	body, _ := json.Marshal(map[string]any{
		"name": "Series", "kind": "tv", "paths": []string{dir}, "watchEnabled": false,
	})
	req := httptest.NewRequest(http.MethodPost, "/libraries", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	rWrite.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var created libraryResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))

	req = httptest.NewRequest(http.MethodGet, "/libraries", nil)
	rec = httptest.NewRecorder()
	rRead.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	req = httptest.NewRequest(http.MethodGet, "/libraries/"+jsonNumber(created.ID), nil)
	rec = httptest.NewRecorder()
	rRead.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	req = httptest.NewRequest(http.MethodGet, "/libraries/"+jsonNumber(created.ID)+"/scans", nil)
	rec = httptest.NewRecorder()
	rRead.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	updateBody, _ := json.Marshal(map[string]any{"name": "Renamed", "kind": "tv", "paths": []string{dir}})
	req = httptest.NewRequest(http.MethodPut, "/libraries/"+jsonNumber(created.ID), bytes.NewReader(updateBody))
	rec = httptest.NewRecorder()
	rWrite.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	req = httptest.NewRequest(http.MethodPost, "/libraries/"+jsonNumber(created.ID)+"/scan", nil)
	rec = httptest.NewRecorder()
	rWrite.ServeHTTP(rec, req)
	require.Equal(t, http.StatusAccepted, rec.Code)

	req = httptest.NewRequest(http.MethodDelete, "/libraries/"+jsonNumber(created.ID), nil)
	rec = httptest.NewRecorder()
	rWrite.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNoContent, rec.Code)

	_ = ctx
}

func TestFilterByParental_ChildProfile(t *testing.T) {
	rating := "R"
	items := []db.MediaItem{{ID: 1, ContentRating: &rating}}
	pg := "PG"
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(auth.WithAuthContext(req.Context(), auth.AuthContext{
		Profile: db.UserProfile{IsChild: true, MaxContentRating: &pg},
	}))

	filtered := filterByParental(req, items)
	assert.Empty(t, filtered)

	req = req.WithContext(context.Background())
	assert.Len(t, filterByParental(req, items), 1)
}

func TestAuthMiddleware_RejectsInvalidBearer(t *testing.T) {
	d := openTestDB(t)
	svc := newTestAuthService(t, d)
	h := New(Options{
		AuthMiddleware:  &AuthMiddleware{Service: svc},
		ProfileHandlers: &ProfileHandlers{Profiles: db.NewProfileRepo(d.Querier())},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/profile", nil)
	req.Header.Set("Authorization", "Bearer not-a-valid-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthHandlers_RevokeSessionForbidden(t *testing.T) {
	h, svc, _, access := newSprint03Router(t)
	ctx := context.Background()

	member, err := svc.Register(ctx, "member", "MemberPass1", []string{auth.RoleUser})
	require.NoError(t, err)
	_, memberPair, err := svc.Login(ctx, "member", "MemberPass1", "phone", "127.0.0.1")
	require.NoError(t, err)

	req := authRequest(http.MethodDelete, "/api/v1/auth/sessions/"+memberPair.SessionID, access, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)

	_ = member
}

func TestAuthHandlers_LoginWhenLocked(t *testing.T) {
	d := openTestDB(t)
	svc := newTestAuthService(t, d)
	h := New(Options{AuthHandlers: &AuthHandlers{Service: svc}})

	setupBody, _ := json.Marshal(map[string]string{"username": "admin", "password": "AdminPass1"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/setup/admin", bytes.NewReader(setupBody))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	badLogin, _ := json.Marshal(map[string]string{"username": "admin", "password": "wrong"})
	for i := 0; i < 6; i++ {
		req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(badLogin))
		rec = httptest.NewRecorder()
		h.ServeHTTP(rec, req)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(badLogin))
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusTooManyRequests, rec.Code)
}

func TestUserHandlers_InvalidJSON(t *testing.T) {
	h, _, _, access := newSprint03Router(t)
	req := authRequest(http.MethodPost, "/api/v1/users", access, []byte(`{`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)

	req = authRequest(http.MethodPut, "/api/v1/users/"+fmt.Sprint("missing"), access, []byte(`{`))
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestProfileHandlers_InvalidUpdate(t *testing.T) {
	h, _, _, access := newSprint03Router(t)
	req := authRequest(http.MethodPut, "/api/v1/profile", access, []byte(`{`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAuthHandlers_RevokeMissingSession(t *testing.T) {
	h, _, _, access := newSprint03Router(t)
	req := authRequest(http.MethodDelete, "/api/v1/auth/sessions/does-not-exist", access, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestAuthHandlers_SetupStatusAndValidation(t *testing.T) {
	d := openTestDB(t)
	svc := newTestAuthService(t, d)
	h := New(Options{AuthHandlers: &AuthHandlers{Service: svc}})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/setup/status", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	req = httptest.NewRequest(http.MethodPost, "/api/v1/setup/admin", bytes.NewReader([]byte(`{"username":"admin","password":"weak"}`)))
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)

	req = httptest.NewRequest(http.MethodPost, "/api/v1/setup/admin", bytes.NewReader([]byte(`{`)))
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAuthMiddleware_AcceptsValidBearer(t *testing.T) {
	h, _, _, access := newSprint03Router(t)
	req := authRequest(http.MethodGet, "/api/v1/profile", access, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestRequirePermission_DeniesMissingPermission(t *testing.T) {
	ac := auth.AuthContext{
		Claims:      auth.Claims{Subject: auth.Subject{Roles: []string{auth.RoleUser}}},
		Permissions: []string{auth.PermLibraryRead},
	}
	denied := false
	handler := RequirePermission(auth.PermUsersManage)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		denied = true
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(auth.WithAuthContext(req.Context(), ac))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)
	assert.False(t, denied)
}

func TestHealth_WithLegacyChecks(t *testing.T) {
	h := newTestHandler(t, Options{
		HealthCheck: []HealthCheck{
			stubHealthCheck{name: "cache", status: HealthStatus{Status: "ok"}},
		},
	})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthHandlers_LoginSuccessAndRefreshInvalid(t *testing.T) {
	h, _, _, _ := newSprint03Router(t)

	loginBody, _ := json.Marshal(map[string]string{"username": "admin", "password": "AdminPass1", "deviceLabel": "laptop"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: refreshTokenCookie, Value: "invalid-token"})
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestWatchHandlers_Unauthorized(t *testing.T) {
	d := openTestDB(t)
	h := New(Options{
		WatchHandlers: &WatchHandlers{Watch: db.NewWatchRepo(d.Querier())},
	})

	cases := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v1/watch-state"},
		{http.MethodGet, "/api/v1/watch-state/1"},
		{http.MethodPut, "/api/v1/watch-state/1"},
		{http.MethodGet, "/api/v1/favorites"},
		{http.MethodPost, "/api/v1/favorites/1"},
		{http.MethodDelete, "/api/v1/favorites/1"},
		{http.MethodGet, "/api/v1/watchlist"},
		{http.MethodPost, "/api/v1/watchlist/1"},
		{http.MethodDelete, "/api/v1/watchlist/1"},
	}
	for _, tc := range cases {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		require.Equal(t, http.StatusUnauthorized, rec.Code, tc.method+" "+tc.path)
	}
}

func TestLibraryHandlers_CreateInvalidJSON(t *testing.T) {
	d := openTestDB(t)
	queue := jobs.NewMemoryQueue(jobs.MemoryQueueConfig{Workers: 1, Buffer: 2})
	t.Cleanup(queue.Close)
	scanner := library.NewScanner(library.ScannerConfig{DB: d, Prober: stubProber{}, Progress: library.NoopProgressPublisher{}})
	svc := library.NewService(library.ServiceConfig{DB: d, Queue: queue, Scanner: scanner})
	r := chi.NewRouter()
	(&LibraryHandlers{Service: svc}).MountWrite(r)

	req := httptest.NewRequest(http.MethodPost, "/libraries", bytes.NewReader([]byte(`{`)))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAuthHandlers_LogoutWithoutCookie(t *testing.T) {
	h, _, _, _ := newSprint03Router(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNoContent, rec.Code)
}

func TestFilterByParental_DefaultChildPG(t *testing.T) {
	rating := "R"
	items := []db.MediaItem{{ID: 1, ContentRating: &rating}}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(auth.WithAuthContext(req.Context(), auth.AuthContext{
		Profile: db.UserProfile{IsChild: true},
	}))
	filtered := filterByParental(req, items)
	assert.Empty(t, filtered)
}

func TestFilterByParental_AdultWithLimit(t *testing.T) {
	rating := "R"
	items := []db.MediaItem{{ID: 1, ContentRating: &rating}}
	pg13 := "PG-13"
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(auth.WithAuthContext(req.Context(), auth.AuthContext{
		Profile: db.UserProfile{MaxContentRating: &pg13},
	}))
	filtered := filterByParental(req, items)
	assert.Empty(t, filtered)

	allowed := "PG"
	items = []db.MediaItem{{ID: 2, ContentRating: &allowed}}
	filtered = filterByParental(req, items)
	assert.Len(t, filtered, 1)
}

func TestUserHandlers_UpdateRolesOnly(t *testing.T) {
	h, _, _, access := newSprint03Router(t)

	createBody, _ := json.Marshal(map[string]any{
		"username": "roles-user",
		"password": "MemberPass1",
		"roles":    []string{"user"},
	})
	req := authRequest(http.MethodPost, "/api/v1/users", access, createBody)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var created map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))
	userID := created["id"].(string)

	updateBody, _ := json.Marshal(map[string]any{"roles": []string{"admin"}})
	req = authRequest(http.MethodPut, "/api/v1/users/"+userID, access, updateBody)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestProfileHandlers_Unauthorized(t *testing.T) {
	d := openTestDB(t)
	h := &ProfileHandlers{Profiles: db.NewProfileRepo(d.Querier())}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/profile", nil)
	rec := httptest.NewRecorder()
	h.get(rec, req)
	require.Equal(t, http.StatusUnauthorized, rec.Code)

	req = httptest.NewRequest(http.MethodPut, "/api/v1/profile", bytes.NewReader([]byte(`{"locale":"tr-TR"}`)))
	rec = httptest.NewRecorder()
	h.update(rec, req)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestWatchHandlers_InvalidBody(t *testing.T) {
	h, _, _, access := newSprint03Router(t)
	req := authRequest(http.MethodPut, "/api/v1/watch-state/1", access, []byte(`{`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHealth_WithAggregator(t *testing.T) {
	reg := diagnostics.NewRegistry()
	reg.Register(diagnostics.NewUptimeProvider())
	h := newTestHandler(t, Options{HealthAggregator: NewDiagnosticsAggregator(reg)})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var body HealthResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.NotEmpty(t, body.Checks)
}

func TestEventBus_ScanProgress(t *testing.T) {
	bus := NewEventBus()
	pub := ScanProgressPublisher{Bus: bus}
	ch := bus.Subscribe()
	pub.PublishScanProgress(context.Background(), library.ProgressEvent{FilesDone: 2, FilesTotal: 10})
	select {
	case msg := <-ch:
		require.Contains(t, string(msg), "scan.progress")
	default:
		t.Fatal("expected scan progress event")
	}
}

func TestEventBus_PublishMarshalError(t *testing.T) {
	bus := NewEventBus()
	ch := bus.Subscribe()
	bus.Publish("bad", make(chan int))
	select {
	case <-ch:
		t.Fatal("unexpected event for unmarshalable payload")
	default:
	}
}

func TestUserHandlers_ForbiddenForMember(t *testing.T) {
	h, svc, _, _ := newSprint03Router(t)
	ctx := context.Background()

	_, err := svc.Register(ctx, "member", "MemberPass1", []string{auth.RoleUser})
	require.NoError(t, err)
	_, pair, err := svc.Login(ctx, "member", "MemberPass1", "web", "127.0.0.1")
	require.NoError(t, err)

	req := authRequest(http.MethodGet, "/api/v1/users", pair.AccessToken, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)
}

func TestRequirePermission_AllowsGranted(t *testing.T) {
	ac := auth.AuthContext{
		Permissions: []string{auth.PermLibraryRead},
	}
	allowed := false
	handler := RequirePermission(auth.PermLibraryRead)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowed = true
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(auth.WithAuthContext(req.Context(), ac))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, allowed)
}
