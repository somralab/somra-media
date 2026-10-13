package db

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlaybackRepo_Lifecycle(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	defer d.Close()

	libRepo := NewLibraryRepo(d.Querier())
	mediaRepo := NewMediaRepo(d.Querier())
	userRepo := NewUserRepo(d.Querier())
	playbackRepo := NewPlaybackRepo(d.Querier())

	dir := t.TempDir()
	lib, err := libRepo.Create(ctx, "Movies", LibraryKindMovie, []string{dir}, false)
	require.NoError(t, err)

	year := 2010
	itemID, err := mediaRepo.CreateItem(ctx, lib.ID, LibraryKindMovie, "Test Movie", &year)
	require.NoError(t, err)

	fileID, err := mediaRepo.UpsertFile(ctx, MediaFile{
		LibraryID: lib.ID, MediaItemID: &itemID,
		Path: dir + "/movie.mp4", FileName: "movie.mp4", SizeBytes: 1000,
	})
	require.NoError(t, err)

	user, err := userRepo.Create(ctx, uuid.NewString(), "streamuser", "hash", []string{"user"})
	require.NoError(t, err)
	userID := user.ID

	sessionID := uuid.NewString()
	expires := time.Now().UTC().Add(2 * time.Hour)
	s := PlaybackSession{
		ID: sessionID, UserID: userID, MediaItemID: itemID, MediaFileID: fileID,
		Mode: PlaybackDirectPlay, Status: PlaybackActive,
		CachePath: "", StartPositionMs: 5000, ExpiresAt: expires,
	}
	require.NoError(t, playbackRepo.Create(ctx, s))

	got, err := playbackRepo.GetByID(ctx, sessionID)
	require.NoError(t, err)
	assert.Equal(t, PlaybackDirectPlay, got.Mode)
	assert.Equal(t, int64(5000), got.StartPositionMs)

	owned, err := playbackRepo.GetByIDForUser(ctx, sessionID, userID)
	require.NoError(t, err)
	assert.Equal(t, sessionID, owned.ID)

	_, err = playbackRepo.GetByIDForUser(ctx, sessionID, "other-user")
	require.Error(t, err)

	active, err := playbackRepo.ListActiveByUser(ctx, userID)
	require.NoError(t, err)
	require.Len(t, active, 1)

	require.NoError(t, playbackRepo.TouchLastAccess(ctx, sessionID))
	require.NoError(t, playbackRepo.Stop(ctx, sessionID))

	got, err = playbackRepo.GetByID(ctx, sessionID)
	require.NoError(t, err)
	assert.Equal(t, PlaybackStopped, got.Status)

	_, err = mediaRepo.GetPrimaryFileByItemID(ctx, itemID)
	require.NoError(t, err)

	require.NoError(t, mediaRepo.UpsertTechnical(ctx, fileID, 60000, "mp4", "h264", 1280, 720, "aac", 2, 0, `{}`))
	tech, err := mediaRepo.GetTechnicalByFileID(ctx, fileID)
	require.NoError(t, err)
	assert.Equal(t, "h264", tech.VideoCodec)
}

func TestPlaybackRepo_TranscodeCountAndIdle(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	defer d.Close()

	libRepo := NewLibraryRepo(d.Querier())
	mediaRepo := NewMediaRepo(d.Querier())
	userRepo := NewUserRepo(d.Querier())
	playbackRepo := NewPlaybackRepo(d.Querier())

	dir := t.TempDir()
	lib, err := libRepo.Create(ctx, "Movies", LibraryKindMovie, []string{dir}, false)
	require.NoError(t, err)
	itemID, err := mediaRepo.CreateItem(ctx, lib.ID, LibraryKindMovie, "HEVC", nil)
	require.NoError(t, err)
	fileID, err := mediaRepo.UpsertFile(ctx, MediaFile{
		LibraryID: lib.ID, MediaItemID: &itemID,
		Path: dir + "/hevc.mkv", FileName: "hevc.mkv", SizeBytes: 1000,
	})
	require.NoError(t, err)
	user, err := userRepo.Create(ctx, uuid.NewString(), "tcuser", "hash", []string{"user"})
	require.NoError(t, err)
	userID := user.ID

	for i := 0; i < 2; i++ {
		require.NoError(t, playbackRepo.Create(ctx, PlaybackSession{
			ID: uuid.NewString(), UserID: userID, MediaItemID: itemID, MediaFileID: fileID,
			Mode: PlaybackTranscode, Status: PlaybackActive,
			ExpiresAt: time.Now().UTC().Add(time.Hour),
		}))
	}

	n, err := playbackRepo.CountActiveTranscodes(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, n)

	cutoff := time.Now().UTC().Add(time.Hour)
	idle, err := playbackRepo.ListIdleSessions(ctx, cutoff)
	require.NoError(t, err)
	assert.Len(t, idle, 2)
}

func TestPlaybackRepo_ExpiredSessions(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	defer d.Close()

	libRepo := NewLibraryRepo(d.Querier())
	mediaRepo := NewMediaRepo(d.Querier())
	userRepo := NewUserRepo(d.Querier())
	playbackRepo := NewPlaybackRepo(d.Querier())

	dir := t.TempDir()
	lib, err := libRepo.Create(ctx, "Movies", LibraryKindMovie, []string{dir}, false)
	require.NoError(t, err)
	itemID, err := mediaRepo.CreateItem(ctx, lib.ID, LibraryKindMovie, "Old Session", nil)
	require.NoError(t, err)
	fileID, err := mediaRepo.UpsertFile(ctx, MediaFile{
		LibraryID: lib.ID, MediaItemID: &itemID,
		Path: dir + "/old.mp4", FileName: "old.mp4", SizeBytes: 100,
	})
	require.NoError(t, err)
	user, err := userRepo.Create(ctx, uuid.NewString(), "expuser", "hash", []string{"user"})
	require.NoError(t, err)

	sessionID := uuid.NewString()
	require.NoError(t, playbackRepo.Create(ctx, PlaybackSession{
		ID: sessionID, UserID: user.ID, MediaItemID: itemID, MediaFileID: fileID,
		Mode: PlaybackDirectPlay, Status: PlaybackActive,
		ExpiresAt: time.Now().UTC().Add(-time.Hour),
	}))

	expired, err := playbackRepo.ListExpired(ctx, time.Now().UTC())
	require.NoError(t, err)
	require.Len(t, expired, 1)
	assert.Equal(t, sessionID, expired[0].ID)

	require.NoError(t, playbackRepo.MarkExpired(ctx, []string{sessionID}))
	got, err := playbackRepo.GetByID(ctx, sessionID)
	require.NoError(t, err)
	assert.Equal(t, PlaybackExpired, got.Status)
}

func TestPlaybackRepo_CreateValidation(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })
	repo := NewPlaybackRepo(d.Querier())
	require.Error(t, repo.Create(ctx, PlaybackSession{}))
	require.ErrorIs(t, repo.UpdateStatus(ctx, "missing", PlaybackFailed, "boom"), ErrPlaybackSessionNotFound)
}

func TestPlaybackRepo_EndedSessionKeepsItsStatus(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	lib, err := NewLibraryRepo(d.Querier()).Create(ctx, "Movies", LibraryKindMovie, []string{t.TempDir()}, false)
	require.NoError(t, err)
	mediaRepo := NewMediaRepo(d.Querier())
	itemID, err := mediaRepo.CreateItem(ctx, lib.ID, LibraryKindMovie, "Film", nil)
	require.NoError(t, err)
	fileID, err := mediaRepo.UpsertFile(ctx, MediaFile{LibraryID: lib.ID, MediaItemID: &itemID, Path: "/m/film.mkv", FileName: "film.mkv"})
	require.NoError(t, err)
	user, err := NewUserRepo(d.Querier()).Create(ctx, uuid.NewString(), "ender", "hash", []string{"user"})
	require.NoError(t, err)
	repo := NewPlaybackRepo(d.Querier())

	newSession := func() string {
		id := uuid.NewString()
		require.NoError(t, repo.Create(ctx, PlaybackSession{
			ID: id, UserID: user.ID, MediaItemID: itemID, MediaFileID: fileID,
			Mode: PlaybackTranscode, Status: PlaybackActive, ExpiresAt: time.Now().UTC().Add(time.Hour),
		}))
		return id
	}
	get := func(id string) PlaybackSession {
		s, err := repo.GetByID(ctx, id)
		require.NoError(t, err)
		return s
	}

	// The reaper expires a session, then its ffmpeg exits with an error and
	// the packaging goroutine reports failure. The session must stay expired.
	reaped := newSession()
	require.NoError(t, repo.UpdateStatus(ctx, reaped, PlaybackExpired, "idle_timeout"))
	require.NoError(t, repo.UpdateStatus(ctx, reaped, PlaybackFailed, "ffmpeg exited"))
	assert.Equal(t, PlaybackExpired, get(reaped).Status)
	assert.Equal(t, "idle_timeout", get(reaped).ErrorMessage)

	// A late "manifest ready" must not revive a session the user stopped.
	stopped := newSession()
	require.NoError(t, repo.Stop(ctx, stopped))
	require.NoError(t, repo.UpdateStatus(ctx, stopped, PlaybackActive, ""))
	assert.Equal(t, PlaybackStopped, get(stopped).Status)

	// Stopping a session that already failed keeps the failure visible.
	failed := newSession()
	require.NoError(t, repo.UpdateStatus(ctx, failed, PlaybackFailed, "unreadable source"))
	require.NoError(t, repo.Stop(ctx, failed))
	assert.Equal(t, PlaybackFailed, get(failed).Status)
	assert.Equal(t, "unreadable source", get(failed).ErrorMessage)

	// Live sessions still move between non-terminal states and can end.
	live := newSession()
	require.NoError(t, repo.UpdateStatus(ctx, live, PlaybackQueued, ""))
	require.NoError(t, repo.UpdateStatus(ctx, live, PlaybackActive, ""))
	require.NoError(t, repo.UpdateStatus(ctx, live, PlaybackFailed, "boom"))
	assert.Equal(t, PlaybackFailed, get(live).Status)
}

func TestRequestRepo_ListWithStatusFilter(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	users := NewUserRepo(d.Querier())
	userID := uuid.NewString()
	_, err := users.Create(ctx, userID, "status-user", "hash", []string{"user"})
	require.NoError(t, err)

	repo := NewRequestRepo(d.Querier())
	_, err = repo.Create(ctx, Request{
		UserID: userID, MediaKind: RequestMediaKindMovie,
		Provider: "tmdb", ExternalID: "st-1", Title: "Pending",
	})
	require.NoError(t, err)
	_, err = repo.Create(ctx, Request{
		UserID: userID, MediaKind: RequestMediaKindMovie,
		Provider: "tmdb", ExternalID: "st-2", Title: "Done",
		Status: RequestStatusCompleted,
	})
	require.NoError(t, err)

	pending, err := repo.List(ctx, RequestListFilter{UserID: userID, Status: RequestStatusPending})
	require.NoError(t, err)
	require.Len(t, pending, 1)
}
