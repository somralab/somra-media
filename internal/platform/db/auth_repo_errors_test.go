package db

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestAuthRepos_ErrorsAfterDBClose(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)

	userID := uuid.NewString()
	users := NewUserRepo(d.Querier())
	sessions := NewSessionRepo(d.Querier())
	profiles := NewProfileRepo(d.Querier())
	watch := NewWatchRepo(d.Querier())
	attempts := NewLoginAttemptRepo(d.Querier())

	_, err := users.Create(ctx, userID, "closed-db", "hash", []string{"user"})
	require.NoError(t, err)

	require.NoError(t, d.Close())

	_, err = users.Count(ctx)
	require.Error(t, err)

	_, err = users.Create(ctx, uuid.NewString(), "x", "hash", []string{"user"})
	require.Error(t, err)

	_, err = users.GetByID(ctx, userID)
	require.Error(t, err)

	_, err = users.GetByUsername(ctx, "closed-db")
	require.Error(t, err)

	_, err = users.List(ctx)
	require.Error(t, err)

	require.Error(t, users.UpdatePassword(ctx, userID, "new"))
	require.Error(t, users.SetDisabled(ctx, userID, true))
	require.Error(t, users.SetRoles(ctx, userID, []string{"user"}))

	_, err = users.PermissionsForUser(ctx, userID)
	require.Error(t, err)

	require.Error(t, sessions.Create(ctx, SessionRecord{
		ID: uuid.NewString(), UserID: userID, TokenHash: "x", ExpiresAt: time.Now().Add(time.Hour),
	}))

	_, err = sessions.GetByTokenHash(ctx, "x")
	require.Error(t, err)

	_, err = sessions.GetByID(ctx, uuid.NewString())
	require.Error(t, err)

	_, err = sessions.ListByUser(ctx, userID)
	require.Error(t, err)

	require.Error(t, sessions.RevokeByID(ctx, uuid.NewString()))
	require.Error(t, sessions.TouchLastUsed(ctx, uuid.NewString()))

	_, err = profiles.Get(ctx, userID)
	require.Error(t, err)

	require.Error(t, profiles.Update(ctx, UserProfile{UserID: userID, Locale: "en-US"}))

	require.Error(t, watch.UpsertWatchState(ctx, WatchState{UserID: userID, MediaItemID: 1}))

	_, err = watch.GetWatchState(ctx, userID, 1)
	require.Error(t, err)

	_, err = watch.ListWatchStates(ctx, userID)
	require.Error(t, err)

	require.Error(t, watch.AddFavorite(ctx, userID, 1))
	require.Error(t, watch.RemoveFavorite(ctx, userID, 1))

	_, err = watch.ListFavorites(ctx, userID)
	require.Error(t, err)

	require.Error(t, watch.AddWatchlist(ctx, userID, 1))
	require.Error(t, watch.RemoveWatchlist(ctx, userID, 1))

	_, err = watch.ListWatchlist(ctx, userID)
	require.Error(t, err)

	_, err = attempts.Get(ctx, "1.2.3.4", LoginAttemptIP)
	require.Error(t, err)

	_, err = attempts.RecordFailure(ctx, "1.2.3.4", LoginAttemptIP, 1, time.Hour)
	require.Error(t, err)

	require.Error(t, attempts.Reset(ctx, "1.2.3.4", LoginAttemptIP))
}

func TestSessionRepo_GetByIDWithRevokedAndLastUsed(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	users := NewUserRepo(d.Querier())
	sessions := NewSessionRepo(d.Querier())

	userID := uuid.NewString()
	_, err := users.Create(ctx, userID, "session-full", "hash", []string{"user"})
	require.NoError(t, err)

	sid := uuid.NewString()
	lastUsed := time.Now().Add(-30 * time.Minute)
	require.NoError(t, sessions.Create(ctx, SessionRecord{
		ID: sid, UserID: userID, DeviceLabel: "desktop", TokenHash: "full-hash",
		ExpiresAt: time.Now().Add(time.Hour), LastUsedAt: &lastUsed,
	}))
	require.NoError(t, sessions.TouchLastUsed(ctx, sid))
	require.NoError(t, sessions.RevokeByID(ctx, sid))

	got, err := sessions.GetByID(ctx, sid)
	require.NoError(t, err)
	require.Equal(t, "desktop", got.DeviceLabel)
	require.NotNil(t, got.LastUsedAt)
	require.NotNil(t, got.RevokedAt)

	list, err := sessions.ListByUser(ctx, userID)
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.NotNil(t, list[0].LastUsedAt)
	require.NotNil(t, list[0].RevokedAt)
}

func TestMediaRepo_UpsertFileExistingPath(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	libRepo := NewLibraryRepo(d.Querier())
	mediaRepo := NewMediaRepo(d.Querier())

	lib, err := libRepo.Create(ctx, "Movies", LibraryKindMovie, []string{t.TempDir()}, false)
	require.NoError(t, err)

	path := "/media/existing.mkv"
	firstID, err := mediaRepo.UpsertFile(ctx, MediaFile{
		LibraryID: lib.ID, Path: path, FileName: "existing.mkv", SizeBytes: 100,
	})
	require.NoError(t, err)
	require.Positive(t, firstID)

	secondID, err := mediaRepo.UpsertFile(ctx, MediaFile{
		LibraryID: lib.ID, Path: path, FileName: "existing.mkv", SizeBytes: 200,
	})
	require.NoError(t, err)
	require.Equal(t, firstID, secondID)

	got, err := mediaRepo.GetFileByPath(ctx, path)
	require.NoError(t, err)
	require.Equal(t, int64(200), got.SizeBytes)
}
