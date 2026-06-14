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
	"time"

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

func TestStreamingHandlers_SessionLifecycle(t *testing.T) {
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
	playReq := httptest.NewRequest(http.MethodPost, "/api/v1/media-items/"+strconv.FormatInt(itemID, 10)+"/play", bytes.NewReader(body))
	playReq.Header.Set("Content-Type", "application/json")
	playRec := httptest.NewRecorder()
	h.ServeHTTP(playRec, playReq)
	require.Equal(t, http.StatusOK, playRec.Code, playRec.Body.String())

	var playResp playResponse
	require.NoError(t, json.Unmarshal(playRec.Body.Bytes(), &playResp))
	sessionID := playResp.SessionID
	require.NotEmpty(t, sessionID)

	waitForStreamingManifest(t, cacheDir, sessionID)

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/streaming/sessions", nil)
	listRec := httptest.NewRecorder()
	h.ServeHTTP(listRec, listReq)
	require.Equal(t, http.StatusOK, listRec.Code)
	var sessions []sessionSummary
	require.NoError(t, json.Unmarshal(listRec.Body.Bytes(), &sessions))
	require.Len(t, sessions, 1)
	assert.Equal(t, sessionID, sessions[0].SessionID)

	masterReq := httptest.NewRequest(http.MethodGet, "/api/v1/streaming/sessions/"+sessionID+"/master.m3u8", nil)
	masterRec := httptest.NewRecorder()
	h.ServeHTTP(masterRec, masterReq)
	require.Equal(t, http.StatusOK, masterRec.Code, masterRec.Body.String())
	assert.Contains(t, masterRec.Header().Get("Content-Type"), "mpegurl")

	segPath := filepath.Join(streaming.SessionCacheDir(cacheDir, sessionID), "seg_00001.m4s")
	require.NoError(t, os.WriteFile(segPath, []byte("segment"), 0o644))
	segReq := httptest.NewRequest(http.MethodGet, "/api/v1/streaming/sessions/"+sessionID+"/seg_00001.m4s", nil)
	segRec := httptest.NewRecorder()
	h.ServeHTTP(segRec, segReq)
	require.Equal(t, http.StatusOK, segRec.Code)

	sourceReq := httptest.NewRequest(http.MethodGet, "/api/v1/streaming/sessions/"+sessionID+"/source", nil)
	sourceRec := httptest.NewRecorder()
	h.ServeHTTP(sourceRec, sourceReq)
	require.Equal(t, http.StatusOK, sourceRec.Code)

	stopReq := httptest.NewRequest(http.MethodDelete, "/api/v1/streaming/sessions/"+sessionID, nil)
	stopRec := httptest.NewRecorder()
	h.ServeHTTP(stopRec, stopReq)
	require.Equal(t, http.StatusNoContent, stopRec.Code)
}

func TestStreamingHandlers_PlayInvalidItem(t *testing.T) {
	d := openTestDB(t)
	cacheDir := t.TempDir()
	stream := streaming.NewService(streaming.ServiceConfig{CacheDir: cacheDir}, db.NewPlaybackRepo(d.Querier()), db.NewMediaRepo(d.Querier()), nil)
	h := testRouterWithAuth(New(Options{
		StreamingHandlers: &StreamingHandlers{
			Streaming: stream,
			Media:     db.NewMediaRepo(d.Querier()),
			Library:   db.NewLibraryRepo(d.Querier()),
			Playback:  db.NewPlaybackRepo(d.Querier()),
			CacheRoot: cacheDir,
		},
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/media-items/not-a-number/play", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestStreamingHandlers_ServeSegmentInvalidPath(t *testing.T) {
	d := openTestDB(t)
	cacheDir := t.TempDir()
	stream := streaming.NewService(streaming.ServiceConfig{CacheDir: cacheDir}, db.NewPlaybackRepo(d.Querier()), db.NewMediaRepo(d.Querier()), nil)

	ctx := context.Background()
	libRepo := db.NewLibraryRepo(d.Querier())
	mediaRepo := db.NewMediaRepo(d.Querier())
	dir := t.TempDir()
	lib, err := libRepo.Create(ctx, "Movies", db.LibraryKindMovie, []string{dir}, false)
	require.NoError(t, err)
	itemID, err := mediaRepo.CreateItem(ctx, lib.ID, db.LibraryKindMovie, "Clip", nil)
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
	playReq := httptest.NewRequest(http.MethodPost, "/api/v1/media-items/"+strconv.FormatInt(itemID, 10)+"/play", bytes.NewReader(body))
	playReq.Header.Set("Content-Type", "application/json")
	playRec := httptest.NewRecorder()
	h.ServeHTTP(playRec, playReq)
	require.Equal(t, http.StatusOK, playRec.Code)

	var playResp playResponse
	require.NoError(t, json.Unmarshal(playRec.Body.Bytes(), &playResp))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/streaming/sessions/"+playResp.SessionID+"/evil.bin", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func waitForStreamingManifest(t *testing.T, cacheRoot, sessionID string) {
	t.Helper()
	dir := streaming.SessionCacheDir(cacheRoot, sessionID)
	for i := 0; i < 50; i++ {
		if _, err := os.Stat(filepath.Join(dir, "master.m3u8")); err == nil {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("streaming manifest not ready")
}
