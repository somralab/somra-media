package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/jobs"
	"github.com/somralab/somra-media/internal/library"
	"github.com/somralab/somra-media/internal/metadata"
	"github.com/somralab/somra-media/internal/platform/db"
)

func TestLibraryHandlers_CRUD(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	dir := t.TempDir()

	queue := jobs.NewMemoryQueue(jobs.MemoryQueueConfig{Workers: 1, Buffer: 2})
	defer queue.Close()

	scanner := library.NewScanner(library.ScannerConfig{DB: d, Prober: stubProber{}, Progress: library.NoopProgressPublisher{}})
	svc := library.NewService(library.ServiceConfig{DB: d, Queue: queue, Scanner: scanner})

	reg := metadata.NewRegistry()
	reg.Register(&metadata.MockProvider{})
	meta := &metadata.Service{DB: &metadata.DBStore{DB: d}, Registry: reg, Matcher: &metadata.Matcher{Registry: reg}}

	h := testRouterWithAuth(New(Options{
		LibraryHandlers: &LibraryHandlers{Service: svc},
		MediaHandlers:   &MediaHandlers{DB: d, Metadata: meta},
	}))

	body, _ := json.Marshal(map[string]any{
		"name": "Films", "kind": "movie", "paths": []string{dir}, "watchEnabled": false,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/libraries", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var created map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))
	id := int64(created["id"].(float64))

	req = httptest.NewRequest(http.MethodGet, "/api/v1/libraries/"+jsonNumber(id), nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	updateBody, _ := json.Marshal(map[string]any{
		"name": "Films Updated", "kind": "movie", "paths": []string{dir}, "watchEnabled": true,
	})
	req = httptest.NewRequest(http.MethodPut, "/api/v1/libraries/"+jsonNumber(id), bytes.NewReader(updateBody))
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/libraries", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	req = httptest.NewRequest(http.MethodPost, "/api/v1/libraries/"+jsonNumber(id)+"/scan", bytes.NewReader([]byte(`{"type":"full"}`)))
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusAccepted, rec.Code)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/libraries/"+jsonNumber(id)+"/scans", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	req = httptest.NewRequest(http.MethodPost, "/api/v1/libraries/"+jsonNumber(id)+"/scan", bytes.NewReader([]byte(`{"type":"incremental"}`)))
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusAccepted, rec.Code)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/libraries/not-a-number", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	req = httptest.NewRequest(http.MethodPost, "/api/v1/libraries", bytes.NewReader([]byte(`not-json`)))
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)

	req = httptest.NewRequest(http.MethodDelete, "/api/v1/libraries/"+jsonNumber(id), nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNoContent, rec.Code)

	_ = ctx
	_ = id
}

type stubProber struct{}

func (stubProber) Probe(context.Context, string) (library.ProbeResult, error) {
	return library.ProbeResult{DurationMs: 1}, nil
}

func jsonNumber(id int64) string {
	return fmt.Sprintf("%d", id)
}

func openTestDB(t *testing.T) *db.DB {
	t.Helper()
	cfg := db.Default()
	cfg.DataDir = t.TempDir()
	d, err := db.Initialize(context.Background(), cfg, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = d.Close() })
	return d
}

func testRouterWithAuth(h http.Handler) http.Handler {
	return testAuthMiddleware(h)
}
