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

func TestStartPlay_MissingTechnical(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	mediaRepo := db.NewMediaRepo(d.Querier())
	libRepo := db.NewLibraryRepo(d.Querier())
	userRepo := db.NewUserRepo(d.Querier())
	dir := t.TempDir()
	lib, _ := libRepo.Create(ctx, "M", db.LibraryKindMovie, []string{dir}, false)
	itemID, _ := mediaRepo.CreateItem(ctx, lib.ID, db.LibraryKindMovie, "T", nil)
	fileID, _ := mediaRepo.UpsertFile(ctx, db.MediaFile{
		LibraryID: lib.ID, MediaItemID: &itemID, Path: filepath.Join(dir, "a.mp4"), FileName: "a.mp4",
	})
	_ = fileID
	user, _ := userRepo.Create(ctx, uuid.NewString(), "u", "h", []string{"user"})
	svc := NewService(ServiceConfig{CacheDir: t.TempDir()}, db.NewPlaybackRepo(d.Querier()), mediaRepo, nil)
	_, err := svc.StartPlay(ctx, PlayRequest{UserID: user.ID, MediaItemID: itemID})
	assert.Error(t, err)
}

func TestEnqueueRespectsMax(t *testing.T) {
	svc := NewService(ServiceConfig{CacheDir: t.TempDir(), MaxTranscodeQueue: 0}, nil, nil, nil)
	svc.enqueue("a", func(context.Context) error { return nil })
	assert.Equal(t, int64(0), svc.Metrics().QueueDepth())
}

func TestDrainQueueEmpty(t *testing.T) {
	svc := NewService(ServiceConfig{CacheDir: t.TempDir()}, nil, nil, nil)
	svc.drainQueue()
}

func TestValidateCacheSegmentPathMaster(t *testing.T) {
	root := t.TempDir()
	session := "sess-1"
	dir := SessionCacheDir(root, session)
	require.NoError(t, EnsureOutputDir(dir))
	path, err := ValidateCacheSegmentPath(root, session, "stream.m3u8")
	require.NoError(t, err)
	assert.Contains(t, path, "stream.m3u8")
}

func TestWriteDirectPlayManifestCreatesFile(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, WriteDirectPlayManifest(dir))
	_, err := os.Stat(filepath.Join(dir, "master.m3u8"))
	require.NoError(t, err)
}

func TestDrainQueueWithJob(t *testing.T) {
	svc := NewService(ServiceConfig{CacheDir: t.TempDir()}, nil, nil, nil)
	svc.queue = append(svc.queue, queuedJob{sessionID: "queued-1", fn: func(context.Context) error { return nil }})
	svc.drainQueue()
}

func TestService_DirectStreamStart(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	playbackRepo := db.NewPlaybackRepo(d.Querier())
	mediaRepo := db.NewMediaRepo(d.Querier())
	libRepo := db.NewLibraryRepo(d.Querier())
	userRepo := db.NewUserRepo(d.Querier())
	dir := t.TempDir()
	cache := t.TempDir()
	lib, _ := libRepo.Create(ctx, "M", db.LibraryKindMovie, []string{dir}, false)
	itemID, _ := mediaRepo.CreateItem(ctx, lib.ID, db.LibraryKindMovie, "MKV", nil)
	src := filepath.Join(dir, "movie.mkv")
	require.NoError(t, os.WriteFile(src, []byte("x"), 0o644))
	fileID, _ := mediaRepo.UpsertFile(ctx, db.MediaFile{LibraryID: lib.ID, MediaItemID: &itemID, Path: src, FileName: "movie.mkv"})
	_ = mediaRepo.UpsertTechnical(ctx, fileID, 1000, "matroska", "h264", 1280, 720, "aac", 2, 0, `{}`)
	user, _ := userRepo.Create(ctx, uuid.NewString(), "u", "h", []string{"user"})
	svc := NewService(ServiceConfig{CacheDir: cache, SessionTTL: time.Hour, FFmpegBin: "false"}, playbackRepo, mediaRepo, nil)
	resp, err := svc.StartPlay(ctx, PlayRequest{UserID: user.ID, MediaItemID: itemID, Capabilities: DefaultBrowserCapabilities()})
	require.NoError(t, err)
	assert.Equal(t, ModeDirectStream, resp.Mode)
}

func TestService_ReapExpiredSessions(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	playbackRepo := db.NewPlaybackRepo(d.Querier())
	mediaRepo := db.NewMediaRepo(d.Querier())
	libRepo := db.NewLibraryRepo(d.Querier())
	userRepo := db.NewUserRepo(d.Querier())
	dir := t.TempDir()
	cache := t.TempDir()
	lib, _ := libRepo.Create(ctx, "M", db.LibraryKindMovie, []string{dir}, false)
	itemID, _ := mediaRepo.CreateItem(ctx, lib.ID, db.LibraryKindMovie, "T", nil)
	src := filepath.Join(dir, "a.mp4")
	require.NoError(t, os.WriteFile(src, []byte("x"), 0o644))
	fileID, _ := mediaRepo.UpsertFile(ctx, db.MediaFile{LibraryID: lib.ID, MediaItemID: &itemID, Path: src, FileName: "a.mp4"})
	_ = mediaRepo.UpsertTechnical(ctx, fileID, 1000, "mp4", "h264", 640, 360, "aac", 2, 0, `{}`)
	user, _ := userRepo.Create(ctx, uuid.NewString(), "u", "h", []string{"user"})
	svc := NewService(ServiceConfig{CacheDir: cache, SessionTTL: time.Hour}, playbackRepo, mediaRepo, nil)
	resp, err := svc.StartPlay(ctx, PlayRequest{UserID: user.ID, MediaItemID: itemID})
	require.NoError(t, err)
	_, err = d.SQL().ExecContext(ctx, `UPDATE playback_session SET expires_at = datetime('now', '-1 hour') WHERE id = ?`, resp.SessionID)
	require.NoError(t, err)
	require.NoError(t, svc.ReapIdle(ctx))
}

func TestBuildLadderSmallSource(t *testing.T) {
	tiers := BuildLadder(320, 240)
	assert.GreaterOrEqual(t, len(tiers), 2)
}

func TestPathValidationRejectsTraversal(t *testing.T) {
	root := t.TempDir()
	_, err := ValidateCacheSegmentPath(root, "s1", "../secret")
	assert.Error(t, err)
}

func TestSessionCacheDirPath(t *testing.T) {
	assert.Contains(t, SessionCacheDir("/cache", "abc"), "abc")
}

func TestAcquireCancelled(t *testing.T) {
	pm := NewProcessManager(ProcessManagerConfig{MaxConcurrent: 1})
	require.NoError(t, pm.Acquire(context.Background()))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	assert.Error(t, pm.Acquire(ctx))
	pm.Release()
}

func TestRunPackagingFailureUpdatesStatus(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	playbackRepo := db.NewPlaybackRepo(d.Querier())
	mediaRepo := db.NewMediaRepo(d.Querier())
	libRepo := db.NewLibraryRepo(d.Querier())
	userRepo := db.NewUserRepo(d.Querier())
	dir := t.TempDir()
	lib, _ := libRepo.Create(ctx, "M", db.LibraryKindMovie, []string{dir}, false)
	itemID, _ := mediaRepo.CreateItem(ctx, lib.ID, db.LibraryKindMovie, "T", nil)
	fileID, _ := mediaRepo.UpsertFile(ctx, db.MediaFile{LibraryID: lib.ID, MediaItemID: &itemID, Path: filepath.Join(dir, "a.mp4"), FileName: "a.mp4"})
	user, _ := userRepo.Create(ctx, uuid.NewString(), "u", "h", []string{"user"})
	svc := NewService(ServiceConfig{CacheDir: t.TempDir()}, playbackRepo, mediaRepo, nil)
	id := uuid.NewString()
	require.NoError(t, playbackRepo.Create(ctx, db.PlaybackSession{
		ID: id, UserID: user.ID, MediaItemID: itemID, MediaFileID: fileID,
		Mode: db.PlaybackTranscode, Status: db.PlaybackActive,
		ExpiresAt: time.Now().Add(time.Hour),
	}))
	svc.runPackaging(id, func(context.Context) error { return assert.AnError })
	sess, err := playbackRepo.GetByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, db.PlaybackFailed, sess.Status)
}

func TestStopSessionTranscode(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	playbackRepo := db.NewPlaybackRepo(d.Querier())
	mediaRepo := db.NewMediaRepo(d.Querier())
	libRepo := db.NewLibraryRepo(d.Querier())
	userRepo := db.NewUserRepo(d.Querier())
	dir := t.TempDir()
	cache := t.TempDir()
	lib, _ := libRepo.Create(ctx, "M", db.LibraryKindMovie, []string{dir}, false)
	itemID, _ := mediaRepo.CreateItem(ctx, lib.ID, db.LibraryKindMovie, "HEVC", nil)
	src := filepath.Join(dir, "hevc.mkv")
	require.NoError(t, os.WriteFile(src, []byte("x"), 0o644))
	fileID, _ := mediaRepo.UpsertFile(ctx, db.MediaFile{LibraryID: lib.ID, MediaItemID: &itemID, Path: src, FileName: "hevc.mkv"})
	_ = mediaRepo.UpsertTechnical(ctx, fileID, 1000, "matroska", "hevc", 1920, 1080, "aac", 2, 0, `{}`)
	user, _ := userRepo.Create(ctx, uuid.NewString(), "u", "h", []string{"user"})
	svc := NewService(ServiceConfig{CacheDir: cache, SessionTTL: time.Hour, FFmpegBin: "false"}, playbackRepo, mediaRepo, nil)
	resp, err := svc.StartPlay(ctx, PlayRequest{UserID: user.ID, MediaItemID: itemID, Capabilities: DefaultBrowserCapabilities()})
	require.NoError(t, err)
	require.NoError(t, svc.StopSession(ctx, resp.SessionID, user.ID))
}

func TestStartPlayMissingFile(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	mediaRepo := db.NewMediaRepo(d.Querier())
	libRepo := db.NewLibraryRepo(d.Querier())
	userRepo := db.NewUserRepo(d.Querier())
	lib, _ := libRepo.Create(ctx, "M", db.LibraryKindMovie, []string{t.TempDir()}, false)
	itemID, _ := mediaRepo.CreateItem(ctx, lib.ID, db.LibraryKindMovie, "orphan", nil)
	user, _ := userRepo.Create(ctx, uuid.NewString(), "u", "h", []string{"user"})
	svc := NewService(ServiceConfig{CacheDir: t.TempDir()}, db.NewPlaybackRepo(d.Querier()), mediaRepo, nil)
	_, err := svc.StartPlay(ctx, PlayRequest{UserID: user.ID, MediaItemID: itemID})
	assert.Error(t, err)
}
