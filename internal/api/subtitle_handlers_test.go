package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/platform/db"
	"github.com/somralab/somra-media/internal/settings"
	"github.com/somralab/somra-media/internal/subtitles"
)

type handlerMockProvider struct {
	results []subtitles.SearchResult
	data    []byte
}

func (m *handlerMockProvider) Name() string { return "mock" }

func (m *handlerMockProvider) Search(_ context.Context, _ subtitles.SearchQuery) ([]subtitles.SearchResult, error) {
	return m.results, nil
}

func (m *handlerMockProvider) Download(_ context.Context, _, _ string) ([]byte, error) {
	return m.data, nil
}

func TestSubtitleHandlers(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	authSvc := newTestAuthService(t, d)
	settingsRepo := db.NewSettingsRepo(d.Querier())
	settingsSvc := settings.NewService(settingsRepo)

	libRepo := db.NewLibraryRepo(d.Querier())
	mediaRepo := db.NewMediaRepo(d.Querier())
	dir := t.TempDir()
	lib, err := libRepo.Create(ctx, "Films", db.LibraryKindMovie, []string{dir}, false)
	require.NoError(t, err)
	year := 2010
	itemID, err := mediaRepo.CreateItem(ctx, lib.ID, db.LibraryKindMovie, "Inception", &year)
	require.NoError(t, err)

	mock := &handlerMockProvider{
		results: []subtitles.SearchResult{{Provider: "mock", ExternalID: "1", Language: "en", Score: 90}},
		data:    []byte("subtitle"),
	}
	subSvc := &subtitles.Service{
		Repo:     db.NewSubtitleRepo(d.Querier()),
		Media:    testMediaLookup{repo: mediaRepo},
		Settings: settingsSvc,
		Storage:  &subtitles.Storage{Root: t.TempDir()},
		Provider: mock,
	}

	h := testRouterWithAuth(New(Options{
		AuthHandlers:     &AuthHandlers{Service: authSvc},
		SubtitleHandlers: &SubtitleHandlers{Service: subSvc},
	}))

	token := setupAdminToken(t, h)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/subtitles/search?mediaItemId=%d&language=en", itemID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	searchBody, _ := json.Marshal(map[string]any{"mediaItemId": itemID, "language": "en"})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/subtitles/search", bytes.NewReader(searchBody))
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	dlBody, _ := json.Marshal(map[string]any{
		"mediaItemId": itemID, "provider": "mock", "externalId": "1", "language": "en",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/subtitles/download", bytes.NewReader(dlBody))
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	require.NoError(t, writer.WriteField("mediaItemId", fmt.Sprintf("%d", itemID)))
	require.NoError(t, writer.WriteField("language", "tr"))
	part, err := writer.CreateFormFile("file", "sub.srt")
	require.NoError(t, err)
	_, err = part.Write([]byte("1\n00:00:01,000 --> 00:00:02,000\nHi"))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	req = httptest.NewRequest(http.MethodPost, "/api/v1/subtitles/upload", &body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/media-items/%d/subtitles", itemID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/subtitles/search?mediaItemId=bad", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func setupAdminToken(t *testing.T, h http.Handler) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"username": "admin", "password": "AdminPass1"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/setup/admin", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)
	var tok map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &tok))
	return tok["accessToken"].(string)
}

type testMediaLookup struct {
	repo *db.MediaRepo
}

func (m testMediaLookup) GetItem(ctx context.Context, id int64) (db.MediaItem, error) {
	return m.repo.GetItemByID(ctx, id, "en-US")
}
