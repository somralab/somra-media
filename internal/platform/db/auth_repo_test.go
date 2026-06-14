package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/platform/db"
)

func TestUserRepo_CRUD(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	repo := db.NewUserRepo(d.Querier())

	id := uuid.NewString()
	user, err := repo.Create(ctx, id, "alice", "hash", []string{"user"})
	require.NoError(t, err)
	assert.Equal(t, "alice", user.Username)
	assert.Contains(t, user.Roles, "user")

	got, err := repo.GetByUsername(ctx, "alice")
	require.NoError(t, err)
	assert.Equal(t, id, got.ID)

	perms, err := repo.PermissionsForUser(ctx, id)
	require.NoError(t, err)
	assert.Contains(t, perms, "library:read")

	n, err := repo.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(1), n)
}

func TestSessionRepo_CreateAndRevoke(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	users := db.NewUserRepo(d.Querier())
	sessions := db.NewSessionRepo(d.Querier())

	id := uuid.NewString()
	_, err := users.Create(ctx, id, "bob", "hash", []string{"user"})
	require.NoError(t, err)

	sid := uuid.NewString()
	err = sessions.Create(ctx, db.SessionRecord{
		ID: sid, UserID: id, TokenHash: "abc123", ExpiresAt: time.Now().Add(time.Hour),
	})
	require.NoError(t, err)

	rec, err := sessions.GetByTokenHash(ctx, "abc123")
	require.NoError(t, err)
	assert.Equal(t, sid, rec.ID)

	require.NoError(t, sessions.RevokeByID(ctx, sid))
}

func TestUserRepo_ListUpdateAndErrors(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	repo := db.NewUserRepo(d.Querier())

	id := uuid.NewString()
	_, err := repo.Create(ctx, id, "dave", "hash", []string{"user"})
	require.NoError(t, err)

	users, err := repo.List(ctx)
	require.NoError(t, err)
	require.Len(t, users, 1)

	require.NoError(t, repo.SetDisabled(ctx, id, true))
	require.NoError(t, repo.UpdatePassword(ctx, id, "new-hash"))
	require.NoError(t, repo.SetRoles(ctx, id, []string{"admin"}))

	got, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	assert.True(t, got.Disabled)
	assert.Contains(t, got.Roles, "admin")

	_, err = repo.Create(ctx, "", "empty", "hash", []string{"user"})
	require.Error(t, err)

	dupID := uuid.NewString()
	_, err = repo.Create(ctx, dupID, "dave", "hash", []string{"user"})
	require.ErrorIs(t, err, db.ErrUserAlreadyExists)
}

func TestProfileRepo_GetUpdate(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	users := db.NewUserRepo(d.Querier())
	profiles := db.NewProfileRepo(d.Querier())

	id := uuid.NewString()
	_, err := users.Create(ctx, id, "eve", "hash", []string{"user"})
	require.NoError(t, err)

	p, err := profiles.Get(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, id, p.UserID)

	rating := "PG-13"
	p.Locale = "tr-TR"
	p.Theme = "aurora"
	p.AvatarURL = "https://example/avatar.png"
	p.MaxContentRating = &rating
	p.IsChild = true
	require.NoError(t, profiles.Update(ctx, p))

	got, err := profiles.Get(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, "tr-TR", got.Locale)
	assert.Equal(t, "aurora", got.Theme)
	require.NotNil(t, got.MaxContentRating)
	assert.Equal(t, "PG-13", *got.MaxContentRating)

	_, err = profiles.Get(ctx, "missing")
	require.ErrorIs(t, err, db.ErrProfileNotFound)
}

func TestLoginAttemptRepo_LockoutCycle(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	repo := db.NewLoginAttemptRepo(d.Querier())

	la, err := repo.Get(ctx, "203.0.113.1", db.LoginAttemptIP)
	require.NoError(t, err)
	assert.Equal(t, 0, la.FailedCount)

	la, err = repo.RecordFailure(ctx, "203.0.113.1", db.LoginAttemptIP, 3, time.Minute)
	require.NoError(t, err)
	assert.Equal(t, 1, la.FailedCount)

	la, err = repo.RecordFailure(ctx, "203.0.113.1", db.LoginAttemptIP, 1, time.Minute)
	require.NoError(t, err)
	require.NotNil(t, la.LockedUntil)

	got, err := repo.Get(ctx, "203.0.113.1", db.LoginAttemptIP)
	require.NoError(t, err)
	assert.Equal(t, 2, got.FailedCount)

	require.NoError(t, repo.Reset(ctx, "203.0.113.1", db.LoginAttemptIP))
	got, err = repo.Get(ctx, "203.0.113.1", db.LoginAttemptIP)
	require.NoError(t, err)
	assert.Equal(t, 0, got.FailedCount)
}

func TestSessionRepo_ListTouchAndRevoke(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	users := db.NewUserRepo(d.Querier())
	sessions := db.NewSessionRepo(d.Querier())

	id := uuid.NewString()
	_, err := users.Create(ctx, id, "frank", "hash", []string{"user"})
	require.NoError(t, err)

	sid := uuid.NewString()
	err = sessions.Create(ctx, db.SessionRecord{
		ID: sid, UserID: id, DeviceLabel: "phone", TokenHash: "hash1",
		ExpiresAt: time.Now().Add(time.Hour),
	})
	require.NoError(t, err)

	got, err := sessions.GetByID(ctx, sid)
	require.NoError(t, err)
	assert.Equal(t, "phone", got.DeviceLabel)

	list, err := sessions.ListByUser(ctx, id)
	require.NoError(t, err)
	require.Len(t, list, 1)

	require.NoError(t, sessions.TouchLastUsed(ctx, sid))
	require.NoError(t, sessions.RevokeSession(ctx, sid))

	sid2 := uuid.NewString()
	require.NoError(t, sessions.Create(ctx, db.SessionRecord{
		ID: sid2, UserID: id, TokenHash: "hash2", ExpiresAt: time.Now().Add(time.Hour),
	}))
	require.NoError(t, sessions.RevokeByID(ctx, sid2))
}

func TestSessionRepo_GetByTokenHashWithRevoked(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	users := db.NewUserRepo(d.Querier())
	sessions := db.NewSessionRepo(d.Querier())

	id := uuid.NewString()
	_, err := users.Create(ctx, id, "session-user", "hash", []string{"user"})
	require.NoError(t, err)

	sid := uuid.NewString()
	lastUsed := time.Now().Add(-time.Hour)
	err = sessions.Create(ctx, db.SessionRecord{
		ID: sid, UserID: id, DeviceLabel: "tablet", TokenHash: "rev-hash",
		ExpiresAt: time.Now().Add(time.Hour), LastUsedAt: &lastUsed,
	})
	require.NoError(t, err)

	rec, err := sessions.GetByTokenHash(ctx, "rev-hash")
	require.NoError(t, err)
	assert.Equal(t, "tablet", rec.DeviceLabel)
	require.NotNil(t, rec.LastUsedAt)

	require.NoError(t, sessions.RevokeByID(ctx, sid))
	rec, err = sessions.GetByTokenHash(ctx, "rev-hash")
	require.NoError(t, err)
	require.NotNil(t, rec.RevokedAt)
}

func TestUserRepo_CreateValidation(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	repo := db.NewUserRepo(d.Querier())

	_, err := repo.Create(ctx, "", "user", "hash", []string{"user"})
	require.Error(t, err)

	_, err = repo.Create(ctx, uuid.NewString(), "", "hash", []string{"user"})
	require.Error(t, err)

	_, err = repo.Create(ctx, uuid.NewString(), "user", "", []string{"user"})
	require.Error(t, err)

	_, err = repo.GetByUsername(ctx, "missing")
	require.ErrorIs(t, err, db.ErrUserNotFound)
}

func TestLoginAttemptRepo_GetMissing(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	repo := db.NewLoginAttemptRepo(d.Querier())

	la, err := repo.Get(ctx, "nobody", db.LoginAttemptUsername)
	require.NoError(t, err)
	assert.Equal(t, 0, la.FailedCount)
}

func TestWatchRepo_WatchlistAndListStates(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	users := db.NewUserRepo(d.Querier())
	watch := db.NewWatchRepo(d.Querier())
	lib := db.NewLibraryRepo(d.Querier())
	media := db.NewMediaRepo(d.Querier())

	uid := uuid.NewString()
	_, err := users.Create(ctx, uid, "grace", "hash", []string{"user"})
	require.NoError(t, err)

	libRec, err := lib.Create(ctx, "Shows", db.LibraryKindTV, []string{"/tmp"}, false)
	require.NoError(t, err)
	itemID, err := media.CreateItem(ctx, libRec.ID, db.LibraryKindTV, "Episode", nil)
	require.NoError(t, err)

	require.NoError(t, watch.UpsertWatchState(ctx, db.WatchState{
		UserID: uid, MediaItemID: itemID, PositionMs: 2000, Completed: true,
	}))
	states, err := watch.ListWatchStates(ctx, uid)
	require.NoError(t, err)
	require.Len(t, states, 1)

	require.NoError(t, watch.AddWatchlist(ctx, uid, itemID))
	ids, err := watch.ListWatchlist(ctx, uid)
	require.NoError(t, err)
	assert.Equal(t, []int64{itemID}, ids)

	require.NoError(t, watch.RemoveWatchlist(ctx, uid, itemID))
	ids, err = watch.ListWatchlist(ctx, uid)
	require.NoError(t, err)
	assert.Empty(t, ids)

	require.NoError(t, watch.AddFavorite(ctx, uid, itemID))
	require.NoError(t, watch.RemoveFavorite(ctx, uid, itemID))

	_, err = watch.GetWatchState(ctx, uid, itemID+999)
	require.Error(t, err)
}

func TestWatchRepo_UpsertAndFavorites(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	users := db.NewUserRepo(d.Querier())
	watch := db.NewWatchRepo(d.Querier())
	lib := db.NewLibraryRepo(d.Querier())
	media := db.NewMediaRepo(d.Querier())

	uid := uuid.NewString()
	_, err := users.Create(ctx, uid, "carol", "hash", []string{"user"})
	require.NoError(t, err)

	libRec, err := lib.Create(ctx, "Films", db.LibraryKindMovie, []string{"/tmp"}, false)
	require.NoError(t, err)
	itemID, err := media.CreateItem(ctx, libRec.ID, db.LibraryKindMovie, "Title", nil)
	require.NoError(t, err)

	require.NoError(t, watch.UpsertWatchState(ctx, db.WatchState{
		UserID: uid, MediaItemID: itemID, PositionMs: 1000, Completed: false,
	}))
	ws, err := watch.GetWatchState(ctx, uid, itemID)
	require.NoError(t, err)
	assert.Equal(t, int64(1000), ws.PositionMs)

	require.NoError(t, watch.AddFavorite(ctx, uid, itemID))
	ids, err := watch.ListFavorites(ctx, uid)
	require.NoError(t, err)
	assert.Equal(t, []int64{itemID}, ids)
}

func TestSessionRepo_ValidationAndRevokeMissing(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	sessions := db.NewSessionRepo(d.Querier())

	err := sessions.Create(ctx, db.SessionRecord{})
	require.Error(t, err)

	err = sessions.RevokeByID(ctx, "missing")
	require.ErrorIs(t, err, db.ErrSessionNotFound)
}

func TestWatchRepo_ErrNotFound(t *testing.T) {
	t.Parallel()
	err := db.ErrNotFound("watch_state")
	assert.Contains(t, err.Error(), "watch_state")
}

func TestProfileRepo_UpdateMissing(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	profiles := db.NewProfileRepo(d.Querier())

	err := profiles.Update(ctx, db.UserProfile{UserID: "missing", Locale: "en-US", Theme: "cinematic"})
	require.ErrorIs(t, err, db.ErrProfileNotFound)
}

func TestUserRepo_AdditionalErrorPaths(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	repo := db.NewUserRepo(d.Querier())

	err := repo.UpdatePassword(ctx, "missing", "hash")
	require.ErrorIs(t, err, db.ErrUserNotFound)

	err = repo.SetDisabled(ctx, "missing", true)
	require.ErrorIs(t, err, db.ErrUserNotFound)

	badID := uuid.NewString()
	_, err = repo.Create(ctx, badID, "henry", "hash", []string{"nonexistent-role"})
	require.Error(t, err)
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
