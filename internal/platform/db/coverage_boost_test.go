package db

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUserRepo_AssignRoleNotFound(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })
	repo := NewUserRepo(d.Querier())

	userID := uuid.NewString()
	_, err := repo.Create(ctx, userID, "role-user", "hash", []string{"user"})
	require.NoError(t, err)

	err = repo.SetRoles(ctx, userID, []string{"missing-role"})
	require.Error(t, err)
}

func TestSessionRepo_GetByTokenHashNotFound(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	_, err := NewSessionRepo(d.Querier()).GetByTokenHash(ctx, "nope")
	require.ErrorIs(t, err, ErrSessionNotFound)
}

func TestWatchRepo_EmptyLists(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })
	users := NewUserRepo(d.Querier())
	watch := NewWatchRepo(d.Querier())

	userID := uuid.NewString()
	_, err := users.Create(ctx, userID, "empty", "hash", []string{"user"})
	require.NoError(t, err)

	states, err := watch.ListWatchStates(ctx, userID)
	require.NoError(t, err)
	require.Empty(t, states)

	favs, err := watch.ListFavorites(ctx, userID)
	require.NoError(t, err)
	require.Empty(t, favs)

	wl, err := watch.ListWatchlist(ctx, userID)
	require.NoError(t, err)
	require.Empty(t, wl)
}

func TestMediaRepo_SearchEmpty(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	ids, err := NewMediaRepo(d.Querier()).SearchTitleFTS(ctx, "zzzznotfound", 5)
	require.NoError(t, err)
	require.Empty(t, ids)
}

func TestUserRepo_GetByIDNotFound(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	_, err := NewUserRepo(d.Querier()).GetByID(ctx, "missing")
	require.ErrorIs(t, err, ErrUserNotFound)
}

func TestUserRepo_SetDisabledNotFound(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	err := NewUserRepo(d.Querier()).SetDisabled(ctx, "missing", true)
	require.ErrorIs(t, err, ErrUserNotFound)
}

func TestLoginAttemptRepo_RecordFailureLocked(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })
	repo := NewLoginAttemptRepo(d.Querier())

	for i := 0; i < 3; i++ {
		_, err := repo.RecordFailure(ctx, "5.6.7.8", LoginAttemptIP, 2, time.Hour)
		require.NoError(t, err)
	}
	la, err := repo.Get(ctx, "5.6.7.8", LoginAttemptIP)
	require.NoError(t, err)
	require.NotNil(t, la.LockedUntil)
}

func TestMediaRepo_GetItemNotFound(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	_, err := NewMediaRepo(d.Querier()).GetItemByID(ctx, 99999, "en-US")
	require.Error(t, err)
}

func TestLibraryRepo_GetNotFound(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	_, err := NewLibraryRepo(d.Querier()).GetByID(ctx, 404)
	require.ErrorIs(t, err, ErrLibraryNotFound)
}

func TestUserRepo_CountUsers(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })
	repo := NewUserRepo(d.Querier())

	n, err := repo.Count(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(0), n)

	_, err = repo.Create(ctx, uuid.NewString(), "solo", "hash", []string{"user"})
	require.NoError(t, err)
	n, err = repo.Count(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(1), n)
}

func TestSessionRepo_CreateWithoutOptionalFields(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })
	users := NewUserRepo(d.Querier())
	sessions := NewSessionRepo(d.Querier())

	userID := uuid.NewString()
	_, err := users.Create(ctx, userID, "minimal-session", "hash", []string{"user"})
	require.NoError(t, err)

	sid := uuid.NewString()
	require.NoError(t, sessions.Create(ctx, SessionRecord{
		ID: sid, UserID: userID, TokenHash: "tok", ExpiresAt: time.Now().Add(time.Hour),
	}))
	_, err = sessions.GetByTokenHash(ctx, "tok")
	require.NoError(t, err)
}

func TestWatchRepo_UpsertUpdatesExisting(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })
	users := NewUserRepo(d.Querier())
	watch := NewWatchRepo(d.Querier())
	lib := NewLibraryRepo(d.Querier())
	media := NewMediaRepo(d.Querier())

	userID := uuid.NewString()
	_, err := users.Create(ctx, userID, "watcher", "hash", []string{"user"})
	require.NoError(t, err)
	libRec, err := lib.Create(ctx, "Lib", LibraryKindMovie, []string{t.TempDir()}, false)
	require.NoError(t, err)
	itemID, err := media.CreateItem(ctx, libRec.ID, LibraryKindMovie, "Film", nil)
	require.NoError(t, err)

	require.NoError(t, watch.UpsertWatchState(ctx, WatchState{UserID: userID, MediaItemID: itemID, PositionMs: 100}))
	require.NoError(t, watch.UpsertWatchState(ctx, WatchState{UserID: userID, MediaItemID: itemID, PositionMs: 200, Completed: true}))
	got, err := watch.GetWatchState(ctx, userID, itemID)
	require.NoError(t, err)
	require.Equal(t, int64(200), got.PositionMs)
	require.True(t, got.Completed)
}

func TestMediaRepo_ListItemsWithLocalizedFields(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	libRepo := NewLibraryRepo(d.Querier())
	mediaRepo := NewMediaRepo(d.Querier())
	dir := t.TempDir()
	lib, err := libRepo.Create(ctx, "Showcase", LibraryKindMovie, []string{dir}, false)
	require.NoError(t, err)
	itemID, err := mediaRepo.CreateItem(ctx, lib.ID, LibraryKindMovie, "Title EN", nil)
	require.NoError(t, err)
	require.NoError(t, mediaRepo.SetLocalizedText(ctx, itemID, "en-US", "title", "Title EN"))
	require.NoError(t, mediaRepo.SetLocalizedText(ctx, itemID, "tr-TR", "title", "Baslik TR"))
	require.NoError(t, mediaRepo.SetLocalizedText(ctx, itemID, "en-US", "overview", "Overview EN"))

	items, err := mediaRepo.ListItemsByLibrary(ctx, lib.ID, "tr-TR", 10, 0)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, "Baslik TR", items[0].Title)
}

func TestUserRepo_CreateWithBadRole(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })
	repo := NewUserRepo(d.Querier())

	_, err := repo.Create(ctx, uuid.NewString(), "bad-role-user", "hash", []string{"missing-role"})
	require.Error(t, err)
}

func TestUserRepo_GetByUsernameCaseInsensitive(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })
	repo := NewUserRepo(d.Querier())

	id := uuid.NewString()
	_, err := repo.Create(ctx, id, "CaseUser", "hash", []string{"user"})
	require.NoError(t, err)

	got, err := repo.GetByUsername(ctx, "caseuser")
	require.NoError(t, err)
	require.Equal(t, id, got.ID)
}

func TestLoginAttemptRepo_GetExistingLocked(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })
	repo := NewLoginAttemptRepo(d.Querier())

	la, err := repo.RecordFailure(ctx, "9.9.9.9", LoginAttemptIP, 1, 2*time.Hour)
	require.NoError(t, err)
	require.NotNil(t, la.LockedUntil)

	got, err := repo.Get(ctx, "9.9.9.9", LoginAttemptIP)
	require.NoError(t, err)
	require.Equal(t, la.FailedCount, got.FailedCount)
	require.NotNil(t, got.LockedUntil)
}

func TestWatchRepo_FavoriteAndWatchlistIdempotent(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })
	users := NewUserRepo(d.Querier())
	watch := NewWatchRepo(d.Querier())
	lib := NewLibraryRepo(d.Querier())
	media := NewMediaRepo(d.Querier())

	userID := uuid.NewString()
	_, err := users.Create(ctx, userID, "idempotent", "hash", []string{"user"})
	require.NoError(t, err)
	libRec, err := lib.Create(ctx, "Lib", LibraryKindMovie, []string{t.TempDir()}, false)
	require.NoError(t, err)
	itemID, err := media.CreateItem(ctx, libRec.ID, LibraryKindMovie, "Film", nil)
	require.NoError(t, err)

	require.NoError(t, watch.AddFavorite(ctx, userID, itemID))
	require.NoError(t, watch.AddFavorite(ctx, userID, itemID))
	require.NoError(t, watch.AddWatchlist(ctx, userID, itemID))
	require.NoError(t, watch.AddWatchlist(ctx, userID, itemID))
	require.NoError(t, watch.RemoveFavorite(ctx, userID, itemID))
	require.NoError(t, watch.RemoveWatchlist(ctx, userID, itemID))
}

func TestLibraryRepo_UpdatePaths(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })
	repo := NewLibraryRepo(d.Querier())
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	lib, err := repo.Create(ctx, "Original", LibraryKindTV, []string{dir1}, true)
	require.NoError(t, err)
	updated, err := repo.Update(ctx, lib.ID, "Updated", []string{dir1, dir2}, false)
	require.NoError(t, err)
	require.Len(t, updated.Paths, 2)
	require.False(t, updated.WatchEnabled)
}
