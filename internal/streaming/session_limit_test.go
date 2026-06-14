package streaming

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/platform/db"
)

func TestService_TranscodeQueueLimit(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	defer d.Close()

	libRepo := db.NewLibraryRepo(d.Querier())
	mediaRepo := db.NewMediaRepo(d.Querier())
	userRepo := db.NewUserRepo(d.Querier())
	playbackRepo := db.NewPlaybackRepo(d.Querier())

	dir := t.TempDir()
	cache := filepath.Join(t.TempDir(), "cache")
	lib, err := libRepo.Create(ctx, "Movies", db.LibraryKindMovie, []string{dir}, false)
	require.NoError(t, err)
	itemID, err := mediaRepo.CreateItem(ctx, lib.ID, db.LibraryKindMovie, "HEVC", nil)
	require.NoError(t, err)
	src := filepath.Join(dir, "hevc.mkv")
	require.NoError(t, writeTestFile(src, "data"))
	fileID, err := mediaRepo.UpsertFile(ctx, db.MediaFile{
		LibraryID: lib.ID, MediaItemID: &itemID, Path: src, FileName: "hevc.mkv", SizeBytes: 4,
	})
	require.NoError(t, err)
	require.NoError(t, mediaRepo.UpsertTechnical(ctx, fileID, 1000, "matroska", "hevc", 1920, 1080, "aac", 2, 0, `{}`))

	user, err := userRepo.Create(ctx, uuid.NewString(), "queueuser", "hash", []string{"user"})
	require.NoError(t, err)

	svc := NewService(ServiceConfig{
		CacheDir: cache, MaxConcurrent: 1, MaxTranscodeQueue: 1, SessionTTL: time.Hour,
	}, playbackRepo, mediaRepo, nil)

	caps := DefaultBrowserCapabilities()
	_, err = svc.StartPlay(ctx, PlayRequest{UserID: user.ID, MediaItemID: itemID, Capabilities: caps})
	require.NoError(t, err)

	n, err := playbackRepo.CountActiveTranscodes(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, n, 1)
}

func TestService_StopSession(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	defer d.Close()

	libRepo := db.NewLibraryRepo(d.Querier())
	mediaRepo := db.NewMediaRepo(d.Querier())
	userRepo := db.NewUserRepo(d.Querier())
	playbackRepo := db.NewPlaybackRepo(d.Querier())

	dir := t.TempDir()
	cache := filepath.Join(t.TempDir(), "cache")
	lib, err := libRepo.Create(ctx, "Movies", db.LibraryKindMovie, []string{dir}, false)
	require.NoError(t, err)
	itemID, err := mediaRepo.CreateItem(ctx, lib.ID, db.LibraryKindMovie, "Clip", nil)
	require.NoError(t, err)
	src := filepath.Join(dir, "clip.mp4")
	require.NoError(t, writeTestFile(src, "data"))
	fileID, err := mediaRepo.UpsertFile(ctx, db.MediaFile{
		LibraryID: lib.ID, MediaItemID: &itemID, Path: src, FileName: "clip.mp4", SizeBytes: 4,
	})
	require.NoError(t, err)
	require.NoError(t, mediaRepo.UpsertTechnical(ctx, fileID, 1000, "mp4", "h264", 640, 360, "aac", 2, 0, `{}`))
	user, err := userRepo.Create(ctx, uuid.NewString(), "stopuser", "hash", []string{"user"})
	require.NoError(t, err)

	svc := NewService(ServiceConfig{CacheDir: cache, SessionTTL: time.Hour}, playbackRepo, mediaRepo, nil)
	resp, err := svc.StartPlay(ctx, PlayRequest{UserID: user.ID, MediaItemID: itemID})
	require.NoError(t, err)
	require.NoError(t, svc.StopSession(ctx, resp.SessionID, user.ID))
}
