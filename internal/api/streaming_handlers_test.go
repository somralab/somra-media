package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/platform/db"
	"github.com/somralab/somra-media/internal/streaming"
)

func TestStreamingHandlers_PlayRequiresAuth(t *testing.T) {
	t.Parallel()
	d := openTestDB(t)
	cacheDir := t.TempDir()
	stream := streaming.NewService(streaming.ServiceConfig{CacheDir: cacheDir}, db.NewPlaybackRepo(d.Querier()), db.NewMediaRepo(d.Querier()), nil)
	h := New(Options{
		AuthMiddleware: &AuthMiddleware{Service: newTestAuthService(t, d)},
		StreamingHandlers: &StreamingHandlers{
			Streaming: stream,
			Media:     db.NewMediaRepo(d.Querier()),
			Library:   db.NewLibraryRepo(d.Querier()),
			Playback:  db.NewPlaybackRepo(d.Querier()),
			CacheRoot: cacheDir,
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/media-items/1/play", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestStreamingHandlers_PlayFlow(t *testing.T) {
	d := openTestDB(t)
	cacheDir := t.TempDir()
	stream := streaming.NewService(streaming.ServiceConfig{CacheDir: cacheDir}, db.NewPlaybackRepo(d.Querier()), db.NewMediaRepo(d.Querier()), nil)

	ctx := context.Background()
	libRepo := db.NewLibraryRepo(d.Querier())
	mediaRepo := db.NewMediaRepo(d.Querier())
	dir := t.TempDir()
	lib, err := libRepo.Create(ctx, "Movies", db.LibraryKindMovie, []string{dir}, false)
	require.NoError(t, err)
	itemID, err := mediaRepo.CreateItem(ctx, lib.ID, db.LibraryKindMovie, "Playable", nil)
	require.NoError(t, err)
	src := filepath.Join(dir, "clip.mp4")
	require.NoError(t, os.WriteFile(src, []byte("fake mp4"), 0o644))
	fileID, err := mediaRepo.UpsertFile(ctx, db.MediaFile{
		LibraryID: lib.ID, MediaItemID: &itemID, Path: src, FileName: "clip.mp4", SizeBytes: 8,
	})
	require.NoError(t, err)
	require.NoError(t, mediaRepo.UpsertTechnical(ctx, fileID, 1000, "mp4", "h264", 640, 360, "aac", 2, 0, `{}`))

	_, err = db.NewUserRepo(d.Querier()).Create(ctx, "test-user", "player", "hash", []string{"user"})
	require.NoError(t, err)

	h := testRouterWithAuth(New(Options{
		StreamingHandlers: &StreamingHandlers{
			Streaming: stream,
			Media:     mediaRepo,
			Library:   libRepo,
			Playback:  db.NewPlaybackRepo(d.Querier()),
			CacheRoot: cacheDir,
		},
	}))

	body, _ := json.Marshal(map[string]any{"capabilities": streaming.DefaultBrowserCapabilities()})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/media-items/"+strconv.FormatInt(itemID, 10)+"/play", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var playResp playResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &playResp))
	assert.NotEmpty(t, playResp.SessionID)
	assert.Equal(t, streaming.ModeDirectPlay, playResp.Mode)
}
