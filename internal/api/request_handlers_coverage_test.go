package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/auth"
	"github.com/somralab/somra-media/internal/jobs"
	"github.com/somralab/somra-media/internal/notifications"
	"github.com/somralab/somra-media/internal/platform/db"
	"github.com/somralab/somra-media/internal/platform/i18n"
	"github.com/somralab/somra-media/internal/requests"
)

func newDirectRequestHandlers(t *testing.T) (*RequestHandlers, *db.DB) {
	t.Helper()
	d := openTestDB(t)
	q := d.Querier()
	fixed := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	return &RequestHandlers{
		Repo:  db.NewRequestRepo(q),
		Users: db.NewUserRepo(q),
		Collision: &requests.CollisionChecker{
			Library:  &requests.DBLibraryLookup{Q: q},
			Requests: &requests.DBPendingRequestLookup{Q: q},
		},
		Policy: &requests.PolicyService{
			Policies: requests.DBPolicyStore{Repo: db.NewRequestRepo(q)},
			Counter:  requests.DBRequestCounter{Repo: db.NewRequestRepo(q)},
		},
		Handoff: requests.NoOpAutomationHandoff{},
		Now:     func() time.Time { return fixed },
		Locale:  func(*http.Request) string { return "tr-TR" },
	}, d
}

func serveRequestRoute(t *testing.T, h *RequestHandlers, method, path string, body []byte, ac *auth.AuthContext) *httptest.ResponseRecorder {
	t.Helper()
	r := chi.NewRouter()
	if ac != nil {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				next.ServeHTTP(w, req.WithContext(auth.WithAuthContext(req.Context(), *ac)))
			})
		})
	}
	h.Mount(r)
	var reader *bytes.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, reader)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func adminAuthContext(t *testing.T, d *db.DB) auth.AuthContext {
	t.Helper()
	uid := uuid.NewString()
	return auth.AuthContext{
		Claims: auth.Claims{
			Subject: auth.Subject{UserID: uid, Roles: []string{auth.RoleAdmin}},
		},
		Permissions: permissionsForRoles(t, d, []string{auth.RoleAdmin}),
	}
}

func userAuthContextWithPerms(userID string, perms []string) auth.AuthContext {
	return auth.AuthContext{
		Claims: auth.Claims{
			Subject: auth.Subject{UserID: userID, Roles: []string{auth.RoleUser}},
		},
		Permissions: perms,
	}
}

func TestRequestHandlers_DirectValidationPaths(t *testing.T) {
	h, d := newDirectRequestHandlers(t)
	admin := adminAuthContext(t, d)

	rec := serveRequestRoute(t, h, http.MethodGet, "/requests", nil, nil)
	require.Equal(t, http.StatusForbidden, rec.Code)

	rec = serveRequestRoute(t, h, http.MethodPost, "/requests", []byte(`{`), &admin)
	require.Equal(t, http.StatusBadRequest, rec.Code)

	rec = serveRequestRoute(t, h, http.MethodPost, "/requests", []byte(`{}`), &admin)
	require.Equal(t, http.StatusBadRequest, rec.Code)

	rec = serveRequestRoute(t, h, http.MethodGet, "/requests/discover", nil, &admin)
	require.Equal(t, http.StatusBadRequest, rec.Code)

	rec = serveRequestRoute(t, h, http.MethodGet, "/requests/not-a-number", nil, &admin)
	require.Equal(t, http.StatusBadRequest, rec.Code)

	rec = serveRequestRoute(t, h, http.MethodPatch, "/requests/policies", []byte(`{`), &admin)
	require.Equal(t, http.StatusBadRequest, rec.Code)

	patch, _ := json.Marshal(map[string]any{"userQuotaPerMonth": -5})
	rec = serveRequestRoute(t, h, http.MethodPatch, "/requests/policies", patch, &admin)
	require.Equal(t, http.StatusBadRequest, rec.Code)

	rec = serveRequestRoute(t, h, http.MethodGet, "/requests?limit=0", nil, &admin)
	require.Equal(t, http.StatusOK, rec.Code)

	ctx := context.Background()
	users := db.NewUserRepo(d.Querier())
	ownerID := uuid.NewString()
	otherID := uuid.NewString()
	_, err := users.Create(ctx, ownerID, "owner", "hash", []string{auth.RoleUser})
	require.NoError(t, err)
	_, err = users.Create(ctx, otherID, "other", "hash", []string{auth.RoleUser})
	require.NoError(t, err)

	id, err := h.Repo.Create(ctx, db.Request{
		UserID: ownerID, MediaKind: db.RequestMediaKindMovie,
		Provider: "tmdb", ExternalID: "direct-1", Title: "Direct",
	})
	require.NoError(t, err)

	userPerms := permissionsForRoles(t, d, []string{auth.RoleUser})
	other := userAuthContextWithPerms(otherID, userPerms)
	rec = serveRequestRoute(t, h, http.MethodGet, "/requests/"+jsonNumber(id), nil, &other)
	require.Equal(t, http.StatusForbidden, rec.Code)

	notePatch, _ := json.Marshal(map[string]any{"adminNote": "secret"})
	rec = serveRequestRoute(t, h, http.MethodPatch, "/requests/"+jsonNumber(id), notePatch, &other)
	require.Equal(t, http.StatusForbidden, rec.Code)

	rec = serveRequestRoute(t, h, http.MethodPatch, "/requests/"+jsonNumber(id), []byte(`{`), &admin)
	require.Equal(t, http.StatusBadRequest, rec.Code)

	rec = serveRequestRoute(t, h, http.MethodGet, "/requests/99999", nil, &admin)
	require.Equal(t, http.StatusNotFound, rec.Code)

	rec = serveRequestRoute(t, h, http.MethodGet, "/requests?userId="+ownerID, nil, &admin)
	require.Equal(t, http.StatusOK, rec.Code)
	var list map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &list))
	assert.NotEmpty(t, list["requests"])

	owner := userAuthContextWithPerms(ownerID, userPerms)
	rec = serveRequestRoute(t, h, http.MethodPost, "/requests/"+jsonNumber(id)+"/cancel", nil, &owner)
	require.Equal(t, http.StatusOK, rec.Code)

	id2, err := h.Repo.Create(ctx, db.Request{
		UserID: ownerID, MediaKind: db.RequestMediaKindMovie,
		Provider: "tmdb", ExternalID: "direct-2", Title: "Approve Twice",
	})
	require.NoError(t, err)
	rec = serveRequestRoute(t, h, http.MethodPost, "/requests/"+jsonNumber(id2)+"/approve", nil, &admin)
	require.Equal(t, http.StatusOK, rec.Code)
	rec = serveRequestRoute(t, h, http.MethodPost, "/requests/"+jsonNumber(id2)+"/approve", nil, &admin)
	require.Equal(t, http.StatusConflict, rec.Code)
}

func TestRequestHandlers_NotifyAndCollisionHelpers(t *testing.T) {
	bundle, err := i18n.NewBundle()
	require.NoError(t, err)
	queue := jobs.NewMemoryQueue(jobs.MemoryQueueConfig{Workers: 1, Buffer: 4})
	t.Cleanup(queue.Close)
	dispatcher := notifications.NewDispatcher(notifications.DispatcherConfig{
		Renderer: notifications.NewTemplateRenderer(bundle),
		Filter:   notifications.NewPreferenceFilter(nil),
		Queue:    queue,
	})
	h := &RequestHandlers{Notify: dispatcher}
	req := requests.Request{ID: 1, UserID: "u1", Title: "Title"}
	h.notifyCreated(context.Background(), req)
	h.notifyApproved(context.Background(), req)
	h.notifyRejected(context.Background(), req)
	h.notifyCompleted(context.Background(), req)

	h.Notify = nil
	h.notifyCreated(context.Background(), req)

	rec := httptest.NewRecorder()
	httpReq := httptest.NewRequest(http.MethodPost, "/", nil)
	h.writeCollisionError(rec, httpReq, requests.ErrCollisionInLibrary)
	require.Equal(t, http.StatusConflict, rec.Code)
	rec = httptest.NewRecorder()
	h.writeCollisionError(rec, httpReq, requests.ErrCollisionDuplicatePending)
	require.Equal(t, http.StatusConflict, rec.Code)
	rec = httptest.NewRecorder()
	h.writeCollisionError(rec, httpReq, assert.AnError)
	require.Equal(t, http.StatusInternalServerError, rec.Code)

	h.publishStatus(db.Request{ID: 9, Status: db.RequestStatusPending})
}

func TestRequestHandlers_HelperFunctions(t *testing.T) {
	ownerAC := auth.AuthContext{Claims: auth.Claims{Subject: auth.Subject{UserID: "owner"}}}
	adminAC := auth.AuthContext{Permissions: []string{auth.PermRequestsManage}}
	row := db.Request{UserID: "owner"}
	assert.True(t, canAccessRequest(ownerAC, row))
	assert.True(t, canAccessRequest(adminAC, db.Request{UserID: "other"}))
	assert.False(t, canAccessRequest(ownerAC, db.Request{UserID: "other"}))

	resp := policyResponse(db.RequestPolicy{
		AutoApproveRoles:  `["admin"]`,
		AdminSettings:     `{"notifyOnCreate":true}`,
		UserQuotaPerMonth: 7,
	})
	assert.Equal(t, 7, resp["userQuotaPerMonth"])
	assert.NotEmpty(t, resp["autoApproveRoles"])
	assert.NotNil(t, resp["adminSettings"])

	h := &RequestHandlers{}
	require.NoError(t, h.runHandoff(context.Background(), requests.Request{ID: 1}))
	assert.Nil(t, h.adminIDs(context.Background()))

	d := openTestDB(t)
	h.Users = db.NewUserRepo(d.Querier())
	ctx := context.Background()
	_, err := h.Users.Create(ctx, uuid.NewString(), "adm", "hash", []string{auth.RoleAdmin})
	require.NoError(t, err)
	assert.NotEmpty(t, h.adminIDs(ctx))

	pub := RequestStatusPublisher{Bus: NewEventBus()}
	pub.PublishRequestStatus(RequestStatusEvent{RequestID: 2, Status: "approved"})
}

func TestRequestHandlers_NowAndLocaleHelpers(t *testing.T) {
	fixed := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	h := &RequestHandlers{
		Now:    func() time.Time { return fixed },
		Locale: func(*http.Request) string { return "tr-TR" },
	}
	assert.Equal(t, fixed, h.now())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	assert.Equal(t, "tr-TR", h.locale(req))

	plain := &RequestHandlers{}
	assert.False(t, plain.now().IsZero())
	assert.Equal(t, "en-US", plain.locale(req))
}
