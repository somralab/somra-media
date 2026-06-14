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

func openTestDB(t *testing.T) *db.DB {
	t.Helper()
	cfg := db.Default()
	cfg.DataDir = t.TempDir()
	d, err := db.Initialize(context.Background(), cfg, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = d.Close() })
	return d
}
