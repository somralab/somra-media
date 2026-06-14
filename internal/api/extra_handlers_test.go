package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/jobs"
	"github.com/somralab/somra-media/internal/library"
	"github.com/somralab/somra-media/internal/metadata"
	"github.com/somralab/somra-media/internal/platform/db"
)

func TestMediaHandlers_ListAndMatch(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	dir := t.TempDir()

	queue := jobs.NewMemoryQueue(jobs.MemoryQueueConfig{Workers: 1, Buffer: 2})
	defer queue.Close()
	scanner := library.NewScanner(library.ScannerConfig{DB: d, Prober: stubProber{}, Progress: library.NoopProgressPublisher{}})
	svc := library.NewService(library.ServiceConfig{DB: d, Queue: queue, Scanner: scanner})
	lib, err := svc.CreateLibrary(ctx, "Films", db.LibraryKindMovie, []string{dir}, false)
	require.NoError(t, err)

	mediaRepo := db.NewMediaRepo(d.Querier())
	itemID, err := mediaRepo.CreateItem(ctx, lib.ID, db.LibraryKindMovie, "Test", nil)
	require.NoError(t, err)

	reg := metadata.NewRegistry()
	reg.Register(metadata.TestProvider{})
	meta := &metadata.Service{DB: &metadata.DBStore{DB: d}, Registry: reg, Matcher: &metadata.Matcher{Registry: reg}}

	h := testRouterWithAuth(New(Options{
		MediaHandlers: &MediaHandlers{
			DB: d, Metadata: meta,
			Locale: func(*http.Request) string { return "tr-TR" },
		},
		BrowseHandlers: &BrowseHandlers{
			Browse: db.NewBrowseRepo(d.Querier()),
			Locale: func(*http.Request) string { return "tr-TR" },
		},
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/libraries/"+jsonNumber(lib.ID)+"/items", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/media-items/"+jsonNumber(itemID), nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/media-items/"+jsonNumber(itemID)+"/match-candidates", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	body, _ := json.Marshal(map[string]string{"provider": "tmdb", "externalId": "1"})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/media-items/"+jsonNumber(itemID)+"/rematch", bytes.NewReader(body))
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNoContent, rec.Code)

	req = httptest.NewRequest(http.MethodPost, "/api/v1/libraries/"+jsonNumber(lib.ID)+"/match", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	req = httptest.NewRequest(http.MethodPost, "/api/v1/media-items/bad/rematch", bytes.NewReader([]byte(`{}`)))
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestEventBus_PublishSubscribe(t *testing.T) {
	bus := NewEventBus()
	ch := bus.Subscribe()
	bus.Publish("scan.progress", map[string]int{"filesDone": 1})
	select {
	case msg := <-ch:
		require.Contains(t, string(msg), "scan.progress")
	default:
		t.Fatal("expected event")
	}
	bus.Unsubscribe(ch)
}
