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

func TestWatchRepo_MultipleEntries(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	users := db.NewUserRepo(d.Querier())
	watch := db.NewWatchRepo(d.Querier())
	lib := db.NewLibraryRepo(d.Querier())
	media := db.NewMediaRepo(d.Querier())

	uid := uuid.NewString()
	_, err := users.Create(ctx, uid, "multi", "hash", []string{"user"})
	require.NoError(t, err)

	libRec, err := lib.Create(ctx, "All", db.LibraryKindMovie, []string{"/tmp"}, false)
	require.NoError(t, err)
	item1, err := media.CreateItem(ctx, libRec.ID, db.LibraryKindMovie, "One", nil)
	require.NoError(t, err)
	item2, err := media.CreateItem(ctx, libRec.ID, db.LibraryKindMovie, "Two", nil)
	require.NoError(t, err)

	require.NoError(t, watch.UpsertWatchState(ctx, db.WatchState{UserID: uid, MediaItemID: item1, PositionMs: 1}))
	require.NoError(t, watch.UpsertWatchState(ctx, db.WatchState{UserID: uid, MediaItemID: item2, PositionMs: 2}))

	states, err := watch.ListWatchStates(ctx, uid)
	require.NoError(t, err)
	require.Len(t, states, 2)

	require.NoError(t, watch.AddFavorite(ctx, uid, item1))
	require.NoError(t, watch.AddFavorite(ctx, uid, item2))
	favs, err := watch.ListFavorites(ctx, uid)
	require.NoError(t, err)
	require.Len(t, favs, 2)

	require.NoError(t, watch.AddWatchlist(ctx, uid, item1))
	require.NoError(t, watch.AddWatchlist(ctx, uid, item2))
	wl, err := watch.ListWatchlist(ctx, uid)
	require.NoError(t, err)
	require.Len(t, wl, 2)

	require.NoError(t, watch.RemoveFavorite(ctx, uid, item1))
	require.NoError(t, watch.RemoveWatchlist(ctx, uid, item2))
}

func TestUserRepo_ListMultiple(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	repo := db.NewUserRepo(d.Querier())

	for _, name := range []string{"amy", "ben", "cara"} {
		_, err := repo.Create(ctx, uuid.NewString(), name, "hash", []string{"user"})
		require.NoError(t, err)
	}

	users, err := repo.List(ctx)
	require.NoError(t, err)
	require.Len(t, users, 3)
	assert.Equal(t, "amy", users[0].Username)
}

func TestProfileRepo_UpdateAvatar(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	users := db.NewUserRepo(d.Querier())
	profiles := db.NewProfileRepo(d.Querier())

	id := uuid.NewString()
	_, err := users.Create(ctx, id, "avatar-user", "hash", []string{"user"})
	require.NoError(t, err)

	p, err := profiles.Get(ctx, id)
	require.NoError(t, err)
	p.AvatarURL = "https://cdn.example/avatar.png"
	p.Locale = "tr-TR"
	require.NoError(t, profiles.Update(ctx, p))

	got, err := profiles.Get(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, "https://cdn.example/avatar.png", got.AvatarURL)
	assert.Equal(t, "tr-TR", got.Locale)
}

func TestSessionRepo_RevokeSessionGroup(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	users := db.NewUserRepo(d.Querier())
	sessions := db.NewSessionRepo(d.Querier())

	id := uuid.NewString()
	_, err := users.Create(ctx, id, "sess-user", "hash", []string{"user"})
	require.NoError(t, err)

	sid := uuid.NewString()
	require.NoError(t, sessions.Create(ctx, db.SessionRecord{
		ID: sid, UserID: id, TokenHash: "group-hash", ExpiresAt: time.Now().Add(time.Hour),
	}))
	require.NoError(t, sessions.TouchLastUsed(ctx, sid))
	require.NoError(t, sessions.RevokeSession(ctx, sid))

	list, err := sessions.ListByUser(ctx, id)
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.NotNil(t, list[0].RevokedAt)
}
