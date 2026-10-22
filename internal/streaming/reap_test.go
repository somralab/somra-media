package streaming

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/platform/db"
)

func TestReapIdle_SessionStaysExpiredWhenPackagingFailsAfterwards(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	playbackRepo := db.NewPlaybackRepo(d.Querier())
	mediaRepo := db.NewMediaRepo(d.Querier())

	dir := t.TempDir()
	lib, err := db.NewLibraryRepo(d.Querier()).Create(ctx, "Movies", db.LibraryKindMovie, []string{dir}, false)
	require.NoError(t, err)
	itemID, err := mediaRepo.CreateItem(ctx, lib.ID, db.LibraryKindMovie, "Idle", nil)
	require.NoError(t, err)
	src := filepath.Join(dir, "idle.mkv")
	require.NoError(t, writeTestFile(src, "data"))
	fileID, err := mediaRepo.UpsertFile(ctx, db.MediaFile{
		LibraryID: lib.ID, MediaItemID: &itemID, Path: src, FileName: "idle.mkv", SizeBytes: 4,
	})
	require.NoError(t, err)
	// HEVC forces a transcode, so the session owns an ffmpeg process.
	require.NoError(t, mediaRepo.UpsertTechnical(ctx, fileID, 1000, "matroska", "hevc", 1920, 1080, "aac", 2, 0, `{}`))
	user, err := db.NewUserRepo(d.Querier()).Create(ctx, uuid.NewString(), "idler", "hash", []string{"user"})
	require.NoError(t, err)

	svc := NewService(ServiceConfig{
		CacheDir: filepath.Join(t.TempDir(), "cache"), MaxConcurrent: 1,
		IdleTimeout: time.Minute, SessionTTL: time.Hour,
	}, playbackRepo, mediaRepo, nil)
	svc.ApplyRuntimeSettings(HWRuntimeConfig{Mode: HWModeOff, MaxHWSessions: 0, MaxTotalSessions: 1})
	// An ffmpeg that never writes a manifest, so the session cannot fail on
	// its own before the reaper runs.
	svc.procMgr = NewProcessManager(ProcessManagerConfig{MaxConcurrent: 1, FFmpegBin: fakeFFmpegScript(t, 30)})

	resp, err := svc.StartPlay(ctx, PlayRequest{UserID: user.ID, MediaItemID: itemID, Capabilities: DefaultBrowserCapabilities()})
	require.NoError(t, err)
	require.Eventually(t, func() bool { return svc.procMgr.RunningCount() == 1 }, 2*time.Second, 20*time.Millisecond)

	_, err = d.SQL().ExecContext(ctx,
		`UPDATE playback_session SET last_access_at = datetime('now', '-2 hours') WHERE id = ?`, resp.SessionID)
	require.NoError(t, err)
	require.NoError(t, svc.ReapIdle(ctx))
	require.Eventually(t, func() bool { return svc.procMgr.RunningCount() == 0 }, 3*time.Second, 20*time.Millisecond,
		"reaping must stop the session's ffmpeg")

	// The packaging goroutine notices its killed ffmpeg only after the reap,
	// and reports the failure.
	svc.runPackaging(resp.SessionID, func(context.Context) error { return errors.New("ffmpeg exited: signal: killed") })

	sess, err := playbackRepo.GetByID(ctx, resp.SessionID)
	require.NoError(t, err)
	assert.Equal(t, db.PlaybackExpired, sess.Status)
	assert.Equal(t, "idle_timeout", sess.ErrorMessage)
}
