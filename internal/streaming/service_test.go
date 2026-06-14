package streaming

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/platform/db"
)

func TestService_StartPlay_DirectPlay(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	defer d.Close()

	libRepo := db.NewLibraryRepo(d.Querier())
	mediaRepo := db.NewMediaRepo(d.Querier())
	userRepo := db.NewUserRepo(d.Querier())

	dir := t.TempDir()
	cache := filepath.Join(t.TempDir(), "cache")
	lib, err := libRepo.Create(ctx, "Movies", db.LibraryKindMovie, []string{dir}, false)
	require.NoError(t, err)

	itemID, err := mediaRepo.CreateItem(ctx, lib.ID, db.LibraryKindMovie, "Clip", nil)
	require.NoError(t, err)

	src := filepath.Join(dir, "clip.mp4")
	require.NoError(t, writeTestFile(src, "fake mp4 content"))

	fileID, err := mediaRepo.UpsertFile(ctx, db.MediaFile{
		LibraryID: lib.ID, MediaItemID: &itemID,
		Path: src, FileName: "clip.mp4", SizeBytes: 20,
	})
	require.NoError(t, err)
	require.NoError(t, mediaRepo.UpsertTechnical(ctx, fileID, 1000, "mp4", "h264", 640, 360, "aac", 2, 0, `{}`))

	user, err := userRepo.Create(ctx, uuid.NewString(), "player", "hash", []string{"user"})
	require.NoError(t, err)
	userID := user.ID

	svc := NewService(ServiceConfig{CacheDir: cache, SessionTTL: time.Hour}, db.NewPlaybackRepo(d.Querier()), mediaRepo, nil)
	resp, err := svc.StartPlay(ctx, PlayRequest{UserID: userID, MediaItemID: itemID})
	require.NoError(t, err)
	assert.Equal(t, ModeDirectPlay, resp.Mode)
	assert.NotEmpty(t, resp.SessionID)

	time.Sleep(200 * time.Millisecond)
	sessDir := SessionCacheDir(cache, resp.SessionID)
	_, err = osStat(filepath.Join(sessDir, "master.m3u8"))
	assert.NoError(t, err)
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

func writeTestFile(path, content string) error {
	return osWriteFile(path, []byte(content), 0o644)
}

var osWriteFile = func(name string, data []byte, perm os.FileMode) error {
	return os.WriteFile(name, data, perm)
}
