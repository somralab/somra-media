package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/auth"
	"github.com/somralab/somra-media/internal/jobs"
	"github.com/somralab/somra-media/internal/metadata"
	"github.com/somralab/somra-media/internal/notifications"
	"github.com/somralab/somra-media/internal/platform/db"
	"github.com/somralab/somra-media/internal/platform/i18n"
	"github.com/somralab/somra-media/internal/requests"
)

func newRequestTestRouter(t *testing.T) (http.Handler, *db.DB, string, *EventBus) {
	t.Helper()
	d := openTestDB(t)
	svc := newTestAuthService(t, d)
	bus := NewEventBus()
	q := d.Querier()
	reqRepo := db.NewRequestRepo(q)
	users := db.NewUserRepo(q)
	reg := metadata.NewRegistry()
	reg.Register(&metadata.MockProvider{})

	bundle, err := i18n.NewBundle()
	require.NoError(t, err)
	queue := jobs.NewMemoryQueue(jobs.MemoryQueueConfig{Workers: 1, Buffer: 4})
	t.Cleanup(queue.Close)
	dispatcher := notifications.NewDispatcher(notifications.DispatcherConfig{
		Renderer: notifications.NewTemplateRenderer(bundle),
		Filter:   notifications.NewPreferenceFilter(nil),
		Queue:    queue,
	})

	h := New(Options{
		EventBus:       bus,
		AuthHandlers:   &AuthHandlers{Service: svc},
		AuthMiddleware: &AuthMiddleware{Service: svc},
		RequestHandlers: &RequestHandlers{
			Repo:  reqRepo,
			Users: users,
			Collision: &requests.CollisionChecker{
				Library:  &requests.DBLibraryLookup{Q: q},
				Requests: &requests.DBPendingRequestLookup{Q: q},
			},
			Policy: &requests.PolicyService{
				Policies: requests.DBPolicyStore{Repo: reqRepo},
				Counter:  requests.DBRequestCounter{Repo: reqRepo},
			},
			Discoverer: &requests.Discoverer{
				Registry: reg,
				Library:  &requests.DBLibraryLookup{Q: q},
				Provider: "mock",
			},
			Handoff:   requests.NoOpAutomationHandoff{},
			Notify:    dispatcher,
			StatusPub: RequestStatusPublisher{Bus: bus},
			Locale:    func(*http.Request) string { return "en-US" },
		},
		NotificationHandlers: &NotificationHandlers{
			Channels:   db.NewNotificationChannelRepo(q),
			Prefs:      db.NewNotificationPreferenceRepo(q),
			Dispatcher: dispatcher,
		},
	})

	setupBody, _ := json.Marshal(map[string]string{"username": "admin", "password": "AdminPass1"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/setup/admin", bytes.NewReader(setupBody))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var tok map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &tok))
	return h, d, tok["accessToken"].(string), bus
}

func TestRequestHandlers_CreateListApprove(t *testing.T) {
	h, d, token, bus := newRequestTestRouter(t)
	ch := bus.Subscribe()
	t.Cleanup(func() { bus.Unsubscribe(ch) })

	body, _ := json.Marshal(map[string]any{
		"mediaKind":  "movie",
		"provider":   "tmdb",
		"externalId": "12345",
		"title":      "Test Movie",
	})
	req := authRequest(http.MethodPost, "/api/v1/requests", token, body)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var created map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))
	assert.Equal(t, "approved", created["status"])
	id := int64(created["id"].(float64))

	select {
	case msg := <-ch:
		require.Contains(t, string(msg), "request.status")
		require.Contains(t, string(msg), `"requestId":`+fmt.Sprint(id))
	default:
		t.Fatal("expected request.status SSE event")
	}

	req = authRequest(http.MethodGet, "/api/v1/requests", token, nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var list map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &list))
	items := list["requests"].([]any)
	require.Len(t, items, 1)

	ctx := context.Background()
	users := db.NewUserRepo(d.Querier())
	userID := uuid.NewString()
	_, err := users.Create(ctx, userID, "viewer", "hash", []string{auth.RoleUser})
	require.NoError(t, err)
	reqRepo := db.NewRequestRepo(d.Querier())
	pendingID, err := reqRepo.Create(ctx, db.Request{
		UserID: userID, MediaKind: db.RequestMediaKindTV,
		Provider: "tmdb", ExternalID: "999", Title: "Pending Show",
	})
	require.NoError(t, err)

	approveBody, _ := json.Marshal(map[string]string{"adminNote": "looks good"})
	req = authRequest(http.MethodPost, fmt.Sprintf("/api/v1/requests/%d/approve", pendingID), token, approveBody)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var approved map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &approved))
	assert.Equal(t, "approved", approved["status"])
	assert.Equal(t, userID, approved["userId"])
}

func TestRequestHandlers_CollisionInLibrary(t *testing.T) {
	h, d, token, _ := newRequestTestRouter(t)
	ctx := context.Background()
	q := d.Querier()
	libRepo := db.NewLibraryRepo(q)
	mediaRepo := db.NewMediaRepo(q)
	dir := t.TempDir()
	lib, err := libRepo.Create(ctx, "Films", db.LibraryKindMovie, []string{dir}, false)
	require.NoError(t, err)
	itemID, err := mediaRepo.CreateItem(ctx, lib.ID, db.LibraryKindMovie, "Owned", nil)
	require.NoError(t, err)
	require.NoError(t, mediaRepo.SetProviderID(ctx, itemID, "tmdb", "777"))

	body, _ := json.Marshal(map[string]any{
		"mediaKind": "movie", "provider": "tmdb", "externalId": "777", "title": "Owned",
	})
	req := authRequest(http.MethodPost, "/api/v1/requests", token, body)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusConflict, rec.Code)
}

func TestRequestHandlers_DiscoverAndPolicies(t *testing.T) {
	h, _, token, _ := newRequestTestRouter(t)

	req := authRequest(http.MethodGet, "/api/v1/requests/discover?q=matrix", token, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var discover map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &discover))
	results := discover["results"].([]any)
	assert.NotEmpty(t, results)

	req = authRequest(http.MethodGet, "/api/v1/requests/policies", token, nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	patch, _ := json.Marshal(map[string]any{"userQuotaPerMonth": 5})
	req = authRequest(http.MethodPatch, "/api/v1/requests/policies", token, patch)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestRequestHandlers_RejectAndCancel(t *testing.T) {
	h, d, token, _ := newRequestTestRouter(t)
	ctx := context.Background()
	users := db.NewUserRepo(d.Querier())
	userID := uuid.NewString()
	_, err := users.Create(ctx, userID, "requser", "hash", []string{auth.RoleUser})
	require.NoError(t, err)
	reqRepo := db.NewRequestRepo(d.Querier())
	id, err := reqRepo.Create(ctx, db.Request{
		UserID: userID, MediaKind: db.RequestMediaKindMovie,
		Provider: "tmdb", ExternalID: "555", Title: "Cancel Me",
	})
	require.NoError(t, err)

	rejectBody, _ := json.Marshal(map[string]string{"adminNote": "no"})
	req := authRequest(http.MethodPost, fmt.Sprintf("/api/v1/requests/%d/reject", id), token, rejectBody)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	id2, err := reqRepo.Create(ctx, db.Request{
		UserID: userID, MediaKind: db.RequestMediaKindMovie,
		Provider: "tmdb", ExternalID: "556", Title: "Cancel Me 2",
	})
	require.NoError(t, err)

	req = authRequest(http.MethodPost, fmt.Sprintf("/api/v1/requests/%d/cancel", id2), token, nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestRequestStatusPublisher_Publishes(t *testing.T) {
	bus := NewEventBus()
	ch := bus.Subscribe()
	pub := RequestStatusPublisher{Bus: bus}
	pub.PublishRequestStatus(RequestStatusEvent{
		RequestID: 42,
		Status:    "pending",
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	})
	select {
	case msg := <-ch:
		require.Contains(t, string(msg), "request.status")
		require.Contains(t, string(msg), `"requestId":42`)
	default:
		t.Fatal("expected request.status event")
	}
}

func TestRequestHandlers_QuotaExceeded(t *testing.T) {
	h, _, token, _ := newRequestTestRouter(t)

	patch, _ := json.Marshal(map[string]any{"userQuotaPerMonth": 1})
	req := authRequest(http.MethodPatch, "/api/v1/requests/policies", token, patch)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	body, _ := json.Marshal(map[string]any{
		"mediaKind": "movie", "provider": "tmdb", "externalId": "quota-1", "title": "First",
	})
	req = authRequest(http.MethodPost, "/api/v1/requests", token, body)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	body, _ = json.Marshal(map[string]any{
		"mediaKind": "movie", "provider": "tmdb", "externalId": "quota-2", "title": "Second",
	})
	req = authRequest(http.MethodPost, "/api/v1/requests", token, body)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)
}

func TestRequestHandlers_GetAndPatch(t *testing.T) {
	h, d, token, _ := newRequestTestRouter(t)
	ctx := context.Background()
	users := db.NewUserRepo(d.Querier())
	userID := uuid.NewString()
	_, err := users.Create(ctx, userID, "getuser", "hash", []string{auth.RoleUser})
	require.NoError(t, err)
	reqRepo := db.NewRequestRepo(d.Querier())
	id, err := reqRepo.Create(ctx, db.Request{
		UserID: userID, MediaKind: db.RequestMediaKindMovie,
		Provider: "tmdb", ExternalID: "get-1", Title: "Get Me",
	})
	require.NoError(t, err)

	req := authRequest(http.MethodGet, fmt.Sprintf("/api/v1/requests/%d", id), token, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	patch, _ := json.Marshal(map[string]any{
		"qualityResolution": "1080p",
		"qualityProfile":    "hd",
		"adminNote":         "reviewed",
	})
	req = authRequest(http.MethodPatch, fmt.Sprintf("/api/v1/requests/%d", id), token, patch)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var updated map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &updated))
	assert.Equal(t, "1080p", updated["qualityResolution"])
	assert.Equal(t, "reviewed", updated["adminNote"])

	req = authRequest(http.MethodGet, "/api/v1/requests/99999", token, nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNotFound, rec.Code)

	req = authRequest(http.MethodGet, "/api/v1/requests?status=pending", token, nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestRequestHandlers_DuplicatePendingCollision(t *testing.T) {
	h, _, token, _ := newRequestTestRouter(t)

	patch, _ := json.Marshal(map[string]any{"userQuotaPerMonth": 10, "autoApproveRoles": []string{}})
	req := authRequest(http.MethodPatch, "/api/v1/requests/policies", token, patch)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	body, _ := json.Marshal(map[string]any{
		"mediaKind": "movie", "provider": "tmdb", "externalId": "dup-1", "title": "Dup",
	})
	req = authRequest(http.MethodPost, "/api/v1/requests", token, body)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var created map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))
	assert.Equal(t, "pending", created["status"])

	req = authRequest(http.MethodPost, "/api/v1/requests", token, body)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusConflict, rec.Code)
}

func TestRequestHandlers_PoliciesValidation(t *testing.T) {
	h, _, token, _ := newRequestTestRouter(t)

	patch, _ := json.Marshal(map[string]any{
		"userQuotaPerMonth": -1,
		"adminSettings":     map[string]any{"notifyOnCreate": true},
	})
	req := authRequest(http.MethodPatch, "/api/v1/requests/policies", token, patch)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)

	patch, _ = json.Marshal(map[string]any{
		"autoApproveRoles":  []string{"admin", "user"},
		"adminSettings":     map[string]any{"notifyOnCreate": true},
		"userQuotaPerMonth": 3,
	})
	req = authRequest(http.MethodPatch, "/api/v1/requests/policies", token, patch)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestRequestHandlers_UserGetsOwnRequest(t *testing.T) {
	h, d, _, _ := newRequestTestRouter(t)
	ctx := context.Background()
	svc := newTestAuthService(t, d)
	_, err := svc.Register(ctx, "ownreq", "UserPass1", []string{auth.RoleUser})
	require.NoError(t, err)
	user, pair, err := svc.Login(ctx, "ownreq", "UserPass1", "web", "127.0.0.1")
	require.NoError(t, err)

	reqRepo := db.NewRequestRepo(d.Querier())
	id, err := reqRepo.Create(ctx, db.Request{
		UserID: user.ID, MediaKind: db.RequestMediaKindMovie,
		Provider: "tmdb", ExternalID: "own-1", Title: "Mine",
	})
	require.NoError(t, err)

	req := authRequest(http.MethodGet, fmt.Sprintf("/api/v1/requests/%d", id), pair.AccessToken, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	req = authRequest(http.MethodGet, fmt.Sprintf("/api/v1/requests?userId=%s", user.ID), pair.AccessToken, nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)
}

func TestRequestHandlers_AdminListByUser(t *testing.T) {
	h, d, token, _ := newRequestTestRouter(t)
	ctx := context.Background()
	users := db.NewUserRepo(d.Querier())
	userID := uuid.NewString()
	_, err := users.Create(ctx, userID, "listed", "hash", []string{auth.RoleUser})
	require.NoError(t, err)
	reqRepo := db.NewRequestRepo(d.Querier())
	_, err = reqRepo.Create(ctx, db.Request{
		UserID: userID, MediaKind: db.RequestMediaKindMovie,
		Provider: "tmdb", ExternalID: "listed-1", Title: "Listed",
	})
	require.NoError(t, err)

	req := authRequest(http.MethodGet, fmt.Sprintf("/api/v1/requests?userId=%s&status=pending", userID), token, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	req = authRequest(http.MethodGet, "/api/v1/requests/not-a-number", token, nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRequestHandlers_PatchNoOp(t *testing.T) {
	h, d, token, _ := newRequestTestRouter(t)
	ctx := context.Background()
	reqRepo := db.NewRequestRepo(d.Querier())
	users := db.NewUserRepo(d.Querier())
	adminID := uuid.NewString()
	_, err := users.Create(ctx, adminID, "admin2", "hash", []string{auth.RoleAdmin})
	require.NoError(t, err)
	id, err := reqRepo.Create(ctx, db.Request{
		UserID: adminID, MediaKind: db.RequestMediaKindMovie,
		Provider: "tmdb", ExternalID: "noop-1", Title: "Noop",
	})
	require.NoError(t, err)

	req := authRequest(http.MethodPatch, fmt.Sprintf("/api/v1/requests/%d", id), token, []byte(`{}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestRequestStatusPublisher_NilBus(t *testing.T) {
	pub := RequestStatusPublisher{}
	pub.PublishRequestStatus(RequestStatusEvent{RequestID: 1, Status: "pending"})
}

func TestRequestHandlers_NotifyCompleted(t *testing.T) {
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
	h.notifyCompleted(context.Background(), requests.Request{ID: 1, UserID: "u1", Title: "Done"})
}

func TestRequestHandlers_UserRBAC(t *testing.T) {
	h, d, adminToken, _ := newRequestTestRouter(t)
	ctx := context.Background()
	svc := newTestAuthService(t, d)

	_, err := svc.Register(ctx, "requser", "UserPass1", []string{auth.RoleUser})
	require.NoError(t, err)
	user, pair, err := svc.Login(ctx, "requser", "UserPass1", "web", "127.0.0.1")
	require.NoError(t, err)
	userToken := pair.AccessToken

	patch, _ := json.Marshal(map[string]any{"userQuotaPerMonth": 5})
	req := authRequest(http.MethodPatch, "/api/v1/requests/policies", userToken, patch)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)

	reqRepo := db.NewRequestRepo(d.Querier())
	pendingID, err := reqRepo.Create(ctx, db.Request{
		UserID: user.ID, MediaKind: db.RequestMediaKindMovie,
		Provider: "tmdb", ExternalID: "rbac-1", Title: "RBAC Test",
	})
	require.NoError(t, err)

	req = authRequest(http.MethodPost, fmt.Sprintf("/api/v1/requests/%d/approve", pendingID), userToken, nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)

	req = authRequest(http.MethodGet, "/api/v1/requests", adminToken, nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}
