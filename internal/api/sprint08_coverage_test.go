package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/auth"
	"github.com/somralab/somra-media/internal/jobs"
	"github.com/somralab/somra-media/internal/library"
	"github.com/somralab/somra-media/internal/metadata"
	"github.com/somralab/somra-media/internal/platform/db"
	"github.com/somralab/somra-media/internal/requests"
)

func TestNotificationHandlers_TestChannelSuccess(t *testing.T) {
	h, ac := newDirectNotificationHandlers(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	createBody, _ := json.Marshal(map[string]any{
		"channelType": "webhook",
		"name":        "Live hook",
		"config":      map[string]any{"url": srv.URL},
		"enabled":     true,
	})
	rec := serveNotificationRoute(t, h, http.MethodPost, "/notifications/channels", createBody, &ac)
	require.Equal(t, http.StatusCreated, rec.Code)

	var created map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))
	chID := int64(created["id"].(float64))

	testBody, _ := json.Marshal(map[string]string{"message": "hello from test"})
	rec = serveNotificationRoute(t, h, http.MethodPost, "/notifications/channels/"+jsonNumber(chID)+"/test", testBody, &ac)
	require.Equal(t, http.StatusNoContent, rec.Code)
}

func TestRequestHandlers_ListByStatus(t *testing.T) {
	h, d := newDirectRequestHandlers(t)
	admin := adminAuthContext(t, d)
	ctx := context.Background()

	users := db.NewUserRepo(d.Querier())
	ownerID := admin.Claims.UserID
	_, err := users.Create(ctx, ownerID, "list-admin", "hash", []string{auth.RoleAdmin})
	require.NoError(t, err)

	_, err = h.Repo.Create(ctx, db.Request{
		UserID: ownerID, MediaKind: db.RequestMediaKindMovie,
		Provider: "tmdb", ExternalID: "list-1", Title: "Pending One",
	})
	require.NoError(t, err)
	approvedID, err := h.Repo.Create(ctx, db.Request{
		UserID: ownerID, MediaKind: db.RequestMediaKindTV,
		Provider: "tmdb", ExternalID: "list-2", Title: "Approved One",
		Status: db.RequestStatusApproved,
	})
	require.NoError(t, err)
	_ = approvedID

	rec := serveRequestRoute(t, h, http.MethodGet, "/requests?status=pending", nil, &admin)
	require.Equal(t, http.StatusOK, rec.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	reqs, ok := body["requests"].([]any)
	require.True(t, ok)
	require.NotEmpty(t, reqs)
}

func TestRequestHandlers_DiscoverWithTestProvider(t *testing.T) {
	h, d := newDirectRequestHandlers(t)
	admin := adminAuthContext(t, d)
	reg := metadata.NewRegistry()
	reg.Register(metadata.TestProvider{})
	h.Discoverer = &requests.Discoverer{
		Registry: reg,
		Library:  &requests.DBLibraryLookup{Q: d.Querier()},
		Provider: "tmdb",
	}

	rec := serveRequestRoute(t, h, http.MethodGet, "/requests/discover?q=matrix&kind=movie", nil, &admin)
	require.Equal(t, http.StatusOK, rec.Code)

	h.Discoverer = nil
	rec = serveRequestRoute(t, h, http.MethodGet, "/requests/discover?q=matrix", nil, &admin)
	require.Equal(t, http.StatusServiceUnavailable, rec.Code)
}

func TestLibraryHandlers_DeleteNotFoundAndListScans(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	queue := jobs.NewMemoryQueue(jobs.MemoryQueueConfig{Workers: 1, Buffer: 2})
	t.Cleanup(queue.Close)
	scanner := library.NewScanner(library.ScannerConfig{DB: d, Prober: stubProber{}, Progress: library.NoopProgressPublisher{}})
	svc := library.NewService(library.ServiceConfig{DB: d, Queue: queue, Scanner: scanner})
	r := chi.NewRouter()
	(&LibraryHandlers{Service: svc}).MountWrite(r)

	req := httptest.NewRequest(http.MethodDelete, "/libraries/99999", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNotFound, rec.Code)

	dir := t.TempDir()
	lib, err := svc.CreateLibrary(ctx, "ScanLib", db.LibraryKindMovie, []string{dir}, false)
	require.NoError(t, err)

	rRead := chi.NewRouter()
	(&LibraryHandlers{Service: svc}).MountRead(rRead)
	req = httptest.NewRequest(http.MethodGet, "/libraries/"+jsonNumber(lib.ID)+"/scans", nil)
	rec = httptest.NewRecorder()
	rRead.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthHandlers_ListSessions(t *testing.T) {
	h, _, _, access := newSprint03Router(t)
	req := authRequest(http.MethodGet, "/api/v1/auth/sessions", access, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestRequestHandlers_CreateWithQuality(t *testing.T) {
	h, d := newDirectRequestHandlers(t)
	admin := adminAuthContext(t, d)
	ctx := context.Background()
	_, err := db.NewUserRepo(d.Querier()).Create(ctx, admin.Claims.UserID, "create-admin", "hash", []string{auth.RoleAdmin})
	require.NoError(t, err)

	body, _ := json.Marshal(map[string]any{
		"mediaKind": "movie", "provider": "tmdb", "externalId": "create-99",
		"title": "New Film", "posterUrl": "https://example.com/p.jpg",
		"qualityResolution": "1080p", "qualityProfile": "web-dl",
	})
	rec := serveRequestRoute(t, h, http.MethodPost, "/requests", body, &admin)
	require.Equal(t, http.StatusCreated, rec.Code)
}
