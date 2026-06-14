package db

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBrowseRepo_FiltersDiscoverAndDetail(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })
	sql := d.SQL()

	mediaRepo := NewMediaRepo(d.Querier())
	browseRepo := NewBrowseRepo(d.Querier())
	watchRepo := NewWatchRepo(d.Querier())
	userRepo := NewUserRepo(d.Querier())

	userID := uuid.NewString()
	_, err := userRepo.Create(ctx, userID, "browse-user", "hash", []string{"user"})
	require.NoError(t, err)

	libRepo := NewLibraryRepo(d.Querier())
	movieLib, err := libRepo.Create(ctx, "Films", LibraryKindMovie, []string{t.TempDir()}, false)
	require.NoError(t, err)
	tvLib, err := libRepo.Create(ctx, "Shows", LibraryKindTV, []string{t.TempDir()}, false)
	require.NoError(t, err)

	actionID, err := mediaRepo.CreateItem(ctx, movieLib.ID, LibraryKindMovie, "Action Hero", intPtr(2020))
	require.NoError(t, err)
	require.NoError(t, mediaRepo.SetLocalizedText(ctx, actionID, "en-US", "title", "Action Hero"))
	_, err = sql.ExecContext(ctx, `UPDATE media_item SET content_rating = ? WHERE id = ?`, "PG-13", actionID)
	require.NoError(t, err)

	dramaID, err := mediaRepo.CreateItem(ctx, movieLib.ID, LibraryKindMovie, "Quiet Drama", intPtr(2021))
	require.NoError(t, err)
	require.NoError(t, mediaRepo.SetLocalizedText(ctx, dramaID, "en-US", "title", "Quiet Drama"))

	_, err = sql.ExecContext(ctx, `INSERT INTO genre (name) VALUES ('Action')`)
	require.NoError(t, err)
	var genreID int64
	require.NoError(t, sql.QueryRowContext(ctx, `SELECT id FROM genre WHERE name = 'Action'`).Scan(&genreID))
	_, err = sql.ExecContext(ctx, `INSERT INTO media_genre (media_item_id, genre_id) VALUES (?, ?)`, actionID, genreID)
	require.NoError(t, err)

	require.NoError(t, watchRepo.UpsertWatchState(ctx, WatchState{
		UserID: userID, MediaItemID: actionID, PositionMs: 120_000, Completed: false,
	}))
	require.NoError(t, watchRepo.AddFavorite(ctx, userID, actionID))
	require.NoError(t, watchRepo.AddWatchlist(ctx, userID, dramaID))

	_, err = sql.ExecContext(ctx,
		`INSERT INTO artwork (media_item_id, kind, source_url) VALUES (?, 'backdrop', 'https://cdn.example/back.jpg')`,
		actionID,
	)
	require.NoError(t, err)
	_, err = sql.ExecContext(ctx, `INSERT INTO person (name) VALUES ('Jane Actor')`)
	require.NoError(t, err)
	var personID int64
	require.NoError(t, sql.QueryRowContext(ctx, `SELECT id FROM person WHERE name = 'Jane Actor'`).Scan(&personID))
	_, err = sql.ExecContext(ctx,
		`INSERT INTO media_person (media_item_id, person_id, role, sort_order) VALUES (?, ?, 'actor', 1)`,
		actionID, personID,
	)
	require.NoError(t, err)

	tvID, err := mediaRepo.CreateItem(ctx, tvLib.ID, LibraryKindTV, "Series One", nil)
	require.NoError(t, err)
	require.NoError(t, mediaRepo.SetLocalizedText(ctx, tvID, "en-US", "title", "Series One"))
	res, err := sql.ExecContext(ctx, `INSERT INTO season (media_item_id, season_number) VALUES (?, 1)`, tvID)
	require.NoError(t, err)
	seasonID, err := res.LastInsertId()
	require.NoError(t, err)
	_, err = sql.ExecContext(ctx,
		`INSERT INTO episode (season_id, episode_number, sort_title) VALUES (?, 1, 'Pilot')`,
		seasonID,
	)
	require.NoError(t, err)

	year := 2020
	page, err := browseRepo.ListPaginated(ctx, movieLib.ID, "en-US", BrowseFilters{
		Limit: 10, Sort: "year", Genre: "Action", Year: &year, WatchStatus: "in_progress", UserID: userID,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, page.Total)
	require.Len(t, page.Items, 1)
	assert.Equal(t, actionID, page.Items[0].ID)

	unwatched, err := browseRepo.ListPaginated(ctx, movieLib.ID, "en-US", BrowseFilters{
		Limit: 10, WatchStatus: "unwatched", UserID: userID,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, unwatched.Total)

	require.NoError(t, watchRepo.UpsertWatchState(ctx, WatchState{
		UserID: userID, MediaItemID: dramaID, PositionMs: 1, Completed: true,
	}))
	completed, err := browseRepo.ListPaginated(ctx, movieLib.ID, "en-US", BrowseFilters{
		Limit: 10, WatchStatus: "completed", UserID: userID,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, completed.Total)

	byTitle, err := browseRepo.ListPaginated(ctx, movieLib.ID, "en-US", BrowseFilters{Limit: 10, Sort: "title"})
	require.NoError(t, err)
	assert.Equal(t, 2, byTitle.Total)

	home, err := browseRepo.DiscoverHome(ctx, userID, "en-US")
	require.NoError(t, err)
	require.NotEmpty(t, home.Shelves)
	var hasContinue, hasRecent, hasRecommended bool
	for _, shelf := range home.Shelves {
		switch shelf.ID {
		case "continueWatching":
			hasContinue = true
			require.NotEmpty(t, shelf.Items)
			assert.NotNil(t, shelf.Items[0].WatchState)
		case "recentlyAdded":
			hasRecent = true
		case "recommended":
			hasRecommended = true
		}
	}
	assert.True(t, hasContinue)
	assert.True(t, hasRecent)
	assert.True(t, hasRecommended)

	detail, err := browseRepo.GetDetail(ctx, actionID, "en-US", userID)
	require.NoError(t, err)
	assert.Equal(t, "Action Hero", detail.Title)
	assert.Contains(t, detail.Genres, "Action")
	require.Len(t, detail.Cast, 1)
	assert.Equal(t, "https://cdn.example/back.jpg", detail.BackdropURL)
	assert.True(t, detail.IsFavorite)
	assert.False(t, detail.InWatchlist)
	require.NotNil(t, detail.WatchState)
	assert.Equal(t, int64(120_000), detail.WatchState.PositionMs)

	tvDetail, err := browseRepo.GetDetail(ctx, tvID, "en-US", "")
	require.NoError(t, err)
	require.Len(t, tvDetail.Seasons, 1)
	require.Len(t, tvDetail.Seasons[0].Episodes, 1)
	assert.Equal(t, "Pilot", tvDetail.Seasons[0].Episodes[0].Title)
}

func TestBrowseRepo_ListPaginatedDefaults(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	browseRepo := NewBrowseRepo(d.Querier())
	libRepo := NewLibraryRepo(d.Querier())
	mediaRepo := NewMediaRepo(d.Querier())

	lib, err := libRepo.Create(ctx, "Empty", LibraryKindMovie, []string{t.TempDir()}, false)
	require.NoError(t, err)
	_, err = mediaRepo.CreateItem(ctx, lib.ID, LibraryKindMovie, "Only", nil)
	require.NoError(t, err)

	page, err := browseRepo.ListPaginated(ctx, lib.ID, "en-US", BrowseFilters{Offset: -1, Limit: 0})
	require.NoError(t, err)
	assert.Equal(t, 50, page.Limit)
	assert.Equal(t, 1, page.Total)
}
