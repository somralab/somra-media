package streaming

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/platform/db"
)

func TestStopSessionReleasesTranscodeSlot(t *testing.T) {
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
	itemID, err := mediaRepo.CreateItem(ctx, lib.ID, db.LibraryKindMovie, "Long", nil)
	require.NoError(t, err)
	src := filepath.Join(dir, "long.mkv")
	require.NoError(t, writeTestFile(src, "data"))
	fileID, err := mediaRepo.UpsertFile(ctx, db.MediaFile{
		LibraryID: lib.ID, MediaItemID: &itemID, Path: src, FileName: "long.mkv", SizeBytes: 4,
	})
	require.NoError(t, err)
	require.NoError(t, mediaRepo.UpsertTechnical(ctx, fileID, 1000, "matroska", "hevc", 1920, 1080, "aac", 2, 0, `{}`))
	user, err := userRepo.Create(ctx, uuid.NewString(), "slotuser", "hash", []string{"user"})
	require.NoError(t, err)

	svc := NewService(ServiceConfig{
		CacheDir: cache, MaxConcurrent: 1, SessionTTL: time.Hour,
	}, playbackRepo, mediaRepo, nil)
	svc.ApplyRuntimeSettings(HWRuntimeConfig{Mode: HWModeOff, MaxHWSessions: 0, MaxTotalSessions: 1})
	svc.procMgr = NewProcessManager(ProcessManagerConfig{MaxConcurrent: 1, FFmpegBin: fakeFFmpegScript(t, 30)})

	resp, err := svc.StartPlay(ctx, PlayRequest{UserID: user.ID, MediaItemID: itemID, Capabilities: DefaultBrowserCapabilities()})
	require.NoError(t, err)

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if svc.procMgr.RunningCount() > 0 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	assert.Equal(t, 1, svc.procMgr.RunningCount())

	require.NoError(t, svc.StopSession(ctx, resp.SessionID, user.ID))

	waitDeadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(waitDeadline) {
		if svc.procMgr.RunningCount() == 0 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	assert.Equal(t, 0, svc.procMgr.RunningCount())
	require.NoError(t, svc.procMgr.Acquire(context.Background()))
	svc.procMgr.Release()
}

func TestHWActiveDecrementsAfterCancel(t *testing.T) {
	svc := NewService(ServiceConfig{CacheDir: t.TempDir(), MaxConcurrent: 1}, nil, nil, nil)
	svc.incHWActive()
	svc.decHWActive()
	assert.Equal(t, 0, svc.activeHWCount())
}

func fakeFFmpegScript(t *testing.T, seconds int) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "ffmpeg")
	script := "#!/bin/sh\nsleep " + strconv.Itoa(seconds) + "\n"
	require.NoError(t, os.WriteFile(path, []byte(script), 0o755))
	return path
}
