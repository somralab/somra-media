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

func TestService_InvalidPlayRequest(t *testing.T) {
	svc := NewService(ServiceConfig{CacheDir: t.TempDir()}, nil, nil, nil)
	_, err := svc.StartPlay(context.Background(), PlayRequest{})
	assert.Error(t, err)
}

func TestService_ReapIdle(t *testing.T) {
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

	svc := NewService(ServiceConfig{CacheDir: cache, IdleTimeout: time.Hour, SessionTTL: time.Hour}, playbackRepo, mediaRepo, nil)
	resp, err := svc.StartPlay(ctx, PlayRequest{UserID: user.ID, MediaItemID: itemID})
	require.NoError(t, err)
	require.NoError(t, svc.ReapIdle(ctx))
	_, err = svc.GetSession(ctx, resp.SessionID, user.ID)
	require.NoError(t, err)
}

func TestService_GetSessionWrongUser(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	playbackRepo := db.NewPlaybackRepo(d.Querier())
	mediaRepo := db.NewMediaRepo(d.Querier())
	svc := NewService(ServiceConfig{CacheDir: t.TempDir()}, playbackRepo, mediaRepo, nil)
	_, err := svc.GetSession(ctx, "missing", "user")
	assert.Error(t, err)
}

func TestBuildFFmpegArgs_DirectStream(t *testing.T) {
	args := BuildFFmpegArgs(PackagerOptions{
		SourcePath: "/m.mkv", OutputDir: "/out", Mode: ModeDirectStream,
		StartPositionMs: 1000, AudioStreamIndex: intPtr(0),
	})
	assert.Contains(t, args, "-c")
	assert.Contains(t, args, "copy")
}

func TestBuildFFmpegArgs_DirectPlayEmpty(t *testing.T) {
	args := BuildFFmpegArgs(PackagerOptions{Mode: ModeDirectPlay, OutputDir: t.TempDir()})
	assert.Contains(t, args, "-map")
}

func TestStartPackaging_DirectPlay(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, StartPackaging(context.Background(), NewProcessManager(ProcessManagerConfig{}), PackagerOptions{
		Mode: ModeDirectPlay, OutputDir: dir,
	}))
	_, err := os.Stat(filepath.Join(dir, "master.m3u8"))
	assert.NoError(t, err)
}

func TestSymlinkOrCopy(t *testing.T) {
	src := filepath.Join(t.TempDir(), "src.bin")
	dst := filepath.Join(t.TempDir(), "dst.bin")
	require.NoError(t, os.WriteFile(src, []byte("payload"), 0o644))
	assert.NoError(t, symlinkOrCopyDefault(src, dst))
	data, err := os.ReadFile(dst)
	require.NoError(t, err)
	assert.Equal(t, "payload", string(data))
}

func TestRemoveDir_Actual(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "f"), []byte("x"), 0o644))
	require.NoError(t, RemoveDir(dir))
	_, err := os.Stat(dir)
	assert.True(t, os.IsNotExist(err))
}

func TestContainsContainer(t *testing.T) {
	assert.True(t, containsContainer([]string{"mp4"}, "mp4"))
	assert.False(t, containsContainer([]string{"mp4"}, "matroska"))
}

func TestIsDirectPlayContainer(t *testing.T) {
	assert.True(t, isDirectPlayContainer("mp4,mov"))
	assert.False(t, isDirectPlayContainer("matroska"))
}

func TestClientCapabilities_AudioChannelsSupported(t *testing.T) {
	caps := DefaultBrowserCapabilities()
	out := caps.AudioChannelsSupported(6)
	assert.NotEmpty(t, out)
}

func TestProcessManager_StopUnknown(t *testing.T) {
	pm := NewProcessManager(ProcessManagerConfig{MaxConcurrent: 1})
	pm.Stop("unknown")
	pm.Release()
}

func TestService_MetricsExposure(t *testing.T) {
	svc := NewService(ServiceConfig{CacheDir: t.TempDir()}, nil, nil, nil)
	m := svc.Metrics()
	m.incActive()
	m.decActive()
	assert.Equal(t, int64(0), m.ActiveSessions())
}

func TestService_TouchSession(t *testing.T) {
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
	require.NoError(t, svc.TouchSession(ctx, resp.SessionID))
}

func TestService_TranscodeQueued(t *testing.T) {
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

	svc := NewService(ServiceConfig{
		CacheDir: cache, MaxConcurrent: 1, MaxTranscodeQueue: 2, SessionTTL: time.Hour,
	}, playbackRepo, mediaRepo, nil)

	caps := DefaultBrowserCapabilities()
	_, err := svc.StartPlay(ctx, PlayRequest{UserID: user.ID, MediaItemID: itemID, Capabilities: caps})
	require.NoError(t, err)
	resp2, err := svc.StartPlay(ctx, PlayRequest{UserID: user.ID, MediaItemID: itemID, Capabilities: caps})
	require.NoError(t, err)
	assert.NotEmpty(t, resp2.SessionID)
}

func TestService_ReapIdleExpires(t *testing.T) {
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
	svc := NewService(ServiceConfig{CacheDir: cache, IdleTimeout: time.Hour, SessionTTL: time.Hour}, playbackRepo, mediaRepo, nil)
	resp, err := svc.StartPlay(ctx, PlayRequest{UserID: user.ID, MediaItemID: itemID})
	require.NoError(t, err)
	_, err = d.SQL().ExecContext(ctx, `UPDATE playback_session SET last_access_at = datetime('now', '-2 hours') WHERE id = ?`, resp.SessionID)
	require.NoError(t, err)
	require.NoError(t, svc.ReapIdle(ctx))
	sess, err := playbackRepo.GetByID(ctx, resp.SessionID)
	require.NoError(t, err)
	assert.Equal(t, db.PlaybackExpired, sess.Status)
}

func TestOsLinkSourceExists(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "source")
	require.NoError(t, os.WriteFile(dst, []byte("x"), 0o644))
	assert.NoError(t, osLinkSource(context.Background(), dir, dir))
}

func TestSymlinkOrCopy_Symlink(t *testing.T) {
	if os.Getenv("GOOS") == "windows" {
		t.Skip("symlink test skipped on windows")
	}
	src := filepath.Join(t.TempDir(), "src.bin")
	dst := filepath.Join(t.TempDir(), "link.bin")
	require.NoError(t, os.WriteFile(src, []byte("linked"), 0o644))
	require.NoError(t, os.Symlink(src, dst))
	assert.NoError(t, symlinkOrCopyDefault(src, filepath.Join(t.TempDir(), "other")))
}

func TestProcessManager_RunningCount(t *testing.T) {
	pm := NewProcessManager(ProcessManagerConfig{MaxConcurrent: 1})
	assert.Equal(t, 0, pm.RunningCount())
}

func TestNormalizeCodecBranches(t *testing.T) {
	assert.Equal(t, "ac3", normalizeCodec("eac3"))
	assert.Equal(t, "dts", normalizeCodec("dts-hd"))
	assert.Equal(t, "vp9", normalizeCodec("vp9"))
}

func TestEstimateVideoBitrateBranches(t *testing.T) {
	assert.Equal(t, int64(5_000_000), estimateVideoBitrate(1920, 1080))
	assert.Equal(t, int64(2_500_000), estimateVideoBitrate(1280, 720))
	assert.Equal(t, int64(1_200_000), estimateVideoBitrate(640, 360))
}

func TestProbeEstimatedBitrateHeuristic(t *testing.T) {
	p := MediaProbe{VideoWidth: 100, VideoHeight: 100, DurationMs: 3600000}
	assert.Equal(t, int64(30000), p.EstimatedBitrate())
}

func TestDecisionAudioChannelDownmix(t *testing.T) {
	eng := NewDecisionEngine()
	caps := ClientCapabilities{
		VideoCodecs: []string{"h264"}, AudioCodecs: []string{"aac"},
		Containers: []string{"mp4"}, MaxAudioChannels: 2,
	}
	got := eng.Decide(caps, MediaProbe{
		Container: "mp4", VideoCodec: "h264", AudioCodec: "aac",
		VideoWidth: 1280, VideoHeight: 720, AudioChannels: 6,
	})
	assert.Equal(t, ModeTranscode, got.Mode)
}

func intPtr(v int) *int { return &v }
