package db

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
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

func TestLibraryRepo_DeleteClosedDB(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	repo := NewLibraryRepo(d.Querier())
	require.NoError(t, d.Close())

	err := repo.Delete(ctx, 1)
	require.Error(t, err)
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

func TestRequestRepo_SetStatusNotFound(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })
	repo := NewRequestRepo(d.Querier())
	require.ErrorIs(t, repo.SetStatus(ctx, 99999, RequestStatusApproved), ErrRequestNotFound)
}

func TestBrowseRepo_SearchFTS(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	libRepo := NewLibraryRepo(d.Querier())
	mediaRepo := NewMediaRepo(d.Querier())
	browseRepo := NewBrowseRepo(d.Querier())

	dir := t.TempDir()
	lib, err := libRepo.Create(ctx, "Browse", LibraryKindMovie, []string{dir}, false)
	require.NoError(t, err)
	itemID, err := mediaRepo.CreateItem(ctx, lib.ID, LibraryKindMovie, "Unique Title XYZ", nil)
	require.NoError(t, err)
	require.NoError(t, mediaRepo.IndexFTS(ctx, itemID, "Unique Title XYZ"))

	results, err := browseRepo.SearchFTS(ctx, "Unique", "en-US", 10)
	require.NoError(t, err)
	require.NotEmpty(t, results)
}

func TestMediaRepo_UpsertFileRichMetadata(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	libRepo := NewLibraryRepo(d.Querier())
	mediaRepo := NewMediaRepo(d.Querier())
	dir := t.TempDir()
	lib, err := libRepo.Create(ctx, "Rich", LibraryKindMovie, []string{dir}, false)
	require.NoError(t, err)
	itemID, err := mediaRepo.CreateItem(ctx, lib.ID, LibraryKindMovie, "Rich Movie", nil)
	require.NoError(t, err)

	year := 2020
	season := 1
	episode := 2
	path := dir + "/rich.mkv"
	id, err := mediaRepo.UpsertFile(ctx, MediaFile{
		LibraryID: lib.ID, MediaItemID: &itemID, Path: path, FileName: "rich.mkv",
		SizeBytes: 999, MtimeNs: 123, ContentHash: "abc",
		ParsedTitle: "Parsed", ParsedYear: &year, ParsedSeason: &season, ParsedEpisode: &episode,
	})
	require.NoError(t, err)
	require.Greater(t, id, int64(0))

	got, err := mediaRepo.GetFileByPath(ctx, path)
	require.NoError(t, err)
	require.NotNil(t, got.ParsedYear)
	assert.Equal(t, 2020, *got.ParsedYear)
}

func TestBrowseRepo_GetDetailWithArtwork(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	libRepo := NewLibraryRepo(d.Querier())
	mediaRepo := NewMediaRepo(d.Querier())
	browseRepo := NewBrowseRepo(d.Querier())

	dir := t.TempDir()
	lib, err := libRepo.Create(ctx, "Art", LibraryKindMovie, []string{dir}, false)
	require.NoError(t, err)
	itemID, err := mediaRepo.CreateItem(ctx, lib.ID, LibraryKindMovie, "Art Film", nil)
	require.NoError(t, err)
	require.NoError(t, mediaRepo.UpsertArtwork(ctx, itemID, "poster", "https://example.com/p.jpg", ""))
	require.NoError(t, mediaRepo.UpsertArtwork(ctx, itemID, "backdrop", "", "/local/backdrop.jpg"))

	detail, err := browseRepo.GetDetail(ctx, itemID, "en-US", "")
	require.NoError(t, err)
	require.NotEmpty(t, detail.Images)
	require.Equal(t, "/local/backdrop.jpg", detail.BackdropURL)
}

func TestUserRepo_ListMultiple(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })
	repo := NewUserRepo(d.Querier())

	for i := 0; i < 3; i++ {
		_, err := repo.Create(ctx, uuid.NewString(), "listuser"+string(rune('a'+i)), "hash", []string{"user"})
		require.NoError(t, err)
	}

	all, err := repo.List(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(all), 3)
}

func TestBrowseRepo_DiscoverHomeEmpty(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	users := NewUserRepo(d.Querier())
	userID := uuid.NewString()
	_, err := users.Create(ctx, userID, "home-user", "hash", []string{"user"})
	require.NoError(t, err)

	home, err := NewBrowseRepo(d.Querier()).DiscoverHome(ctx, userID, "en-US")
	require.NoError(t, err)
	require.NotNil(t, home.Shelves)
}

func TestBrowseRepo_DiscoverHomeWithProgress(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	users := NewUserRepo(d.Querier())
	watch := NewWatchRepo(d.Querier())
	libRepo := NewLibraryRepo(d.Querier())
	mediaRepo := NewMediaRepo(d.Querier())

	userID := uuid.NewString()
	_, err := users.Create(ctx, userID, "progress-user", "hash", []string{"user"})
	require.NoError(t, err)

	dir := t.TempDir()
	lib, err := libRepo.Create(ctx, "Progress", LibraryKindMovie, []string{dir}, false)
	require.NoError(t, err)
	itemID, err := mediaRepo.CreateItem(ctx, lib.ID, LibraryKindMovie, "In Progress", nil)
	require.NoError(t, err)
	require.NoError(t, watch.UpsertWatchState(ctx, WatchState{
		UserID: userID, MediaItemID: itemID, PositionMs: 120000, Completed: false,
	}))

	home, err := NewBrowseRepo(d.Querier()).DiscoverHome(ctx, userID, "en-US")
	require.NoError(t, err)
	var found bool
	for _, shelf := range home.Shelves {
		if len(shelf.Items) > 0 {
			found = true
			break
		}
	}
	require.True(t, found)
}

func TestBrowseRepo_GetDetailWithFavorite(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	users := NewUserRepo(d.Querier())
	watch := NewWatchRepo(d.Querier())
	libRepo := NewLibraryRepo(d.Querier())
	mediaRepo := NewMediaRepo(d.Querier())
	browseRepo := NewBrowseRepo(d.Querier())

	userID := uuid.NewString()
	_, err := users.Create(ctx, userID, "fav-user", "hash", []string{"user"})
	require.NoError(t, err)

	dir := t.TempDir()
	lib, err := libRepo.Create(ctx, "Fav", LibraryKindMovie, []string{dir}, false)
	require.NoError(t, err)
	itemID, err := mediaRepo.CreateItem(ctx, lib.ID, LibraryKindMovie, "Favorite Film", nil)
	require.NoError(t, err)
	require.NoError(t, watch.AddFavorite(ctx, userID, itemID))
	require.NoError(t, watch.AddWatchlist(ctx, userID, itemID))

	detail, err := browseRepo.GetDetail(ctx, itemID, "en-US", userID)
	require.NoError(t, err)
	require.True(t, detail.IsFavorite)
	require.True(t, detail.InWatchlist)
}

func TestMediaRepo_MetadataFields(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	libRepo := NewLibraryRepo(d.Querier())
	mediaRepo := NewMediaRepo(d.Querier())
	dir := t.TempDir()
	lib, err := libRepo.Create(ctx, "Meta", LibraryKindMovie, []string{dir}, false)
	require.NoError(t, err)
	itemID, err := mediaRepo.CreateItem(ctx, lib.ID, LibraryKindMovie, "Matched", nil)
	require.NoError(t, err)

	require.NoError(t, mediaRepo.SetLocalizedText(ctx, itemID, "en-US", "title", "Localized"))
	score := 0.95
	require.NoError(t, mediaRepo.SetMatch(ctx, itemID, MatchMatched, &score))
	require.NoError(t, mediaRepo.SetProviderID(ctx, itemID, "tmdb", "99"))

	item, err := mediaRepo.GetItemByID(ctx, itemID, "en-US")
	require.NoError(t, err)
	require.Equal(t, "Localized", item.Title)
	require.Equal(t, MatchMatched, item.MatchStatus)
}

func TestRequestRepo_UpdateAllFields(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	users := NewUserRepo(d.Querier())
	userID := uuid.NewString()
	_, err := users.Create(ctx, userID, "upd-user", "hash", []string{"user"})
	require.NoError(t, err)

	repo := NewRequestRepo(d.Querier())
	id, err := repo.Create(ctx, Request{
		UserID: userID, MediaKind: RequestMediaKindTV,
		Provider: "tmdb", ExternalID: "upd-1", Title: "Update Me",
	})
	require.NoError(t, err)

	res := RequestQuality720p
	profile := "web"
	note := "note"
	flag := true
	status := RequestStatusRejected
	require.NoError(t, repo.Update(ctx, id, RequestUpdate{
		QualityResolution: &res,
		QualityProfile:    &profile,
		AdminNote:         &note,
		Status:            &status,
		CollisionFlag:     &flag,
	}))

	got, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	require.Equal(t, RequestQuality720p, got.QualityResolution)
	require.Equal(t, "web", got.QualityProfile)
	require.Equal(t, "note", got.AdminNote)
	require.True(t, got.CollisionFlag)
}

func TestBrowseRepo_ListPaginatedWatchFilter(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	users := NewUserRepo(d.Querier())
	watch := NewWatchRepo(d.Querier())
	libRepo := NewLibraryRepo(d.Querier())
	mediaRepo := NewMediaRepo(d.Querier())
	browseRepo := NewBrowseRepo(d.Querier())

	userID := uuid.NewString()
	_, err := users.Create(ctx, userID, "page-user", "hash", []string{"user"})
	require.NoError(t, err)

	dir := t.TempDir()
	lib, err := libRepo.Create(ctx, "Page", LibraryKindMovie, []string{dir}, false)
	require.NoError(t, err)
	year := 2021
	itemID, err := mediaRepo.CreateItem(ctx, lib.ID, LibraryKindMovie, "Paged", &year)
	require.NoError(t, err)
	require.NoError(t, watch.UpsertWatchState(ctx, WatchState{
		UserID: userID, MediaItemID: itemID, PositionMs: 5000, Completed: false,
	}))

	page, err := browseRepo.ListPaginated(ctx, lib.ID, "en-US", BrowseFilters{
		Year:        &year,
		Sort:        "created_at",
		Limit:       10,
		UserID:      userID,
		WatchStatus: "in_progress",
	})
	require.NoError(t, err)
	require.Equal(t, 1, page.Total)
	require.Len(t, page.Items, 1)
}

func TestBrowseRepo_ListPaginatedGenreFilter(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	libRepo := NewLibraryRepo(d.Querier())
	mediaRepo := NewMediaRepo(d.Querier())
	browseRepo := NewBrowseRepo(d.Querier())

	dir := t.TempDir()
	lib, err := libRepo.Create(ctx, "GenreLib", LibraryKindMovie, []string{dir}, false)
	require.NoError(t, err)
	itemID, err := mediaRepo.CreateItem(ctx, lib.ID, LibraryKindMovie, "Action Hero", nil)
	require.NoError(t, err)

	_, err = d.SQL().ExecContext(ctx, `INSERT INTO genre (name) VALUES ('Action')`)
	require.NoError(t, err)
	var genreID int64
	require.NoError(t, d.SQL().QueryRowContext(ctx, `SELECT id FROM genre WHERE name = 'Action'`).Scan(&genreID))
	_, err = d.SQL().ExecContext(ctx, `INSERT INTO media_genre (media_item_id, genre_id) VALUES (?, ?)`, itemID, genreID)
	require.NoError(t, err)

	page, err := browseRepo.ListPaginated(ctx, lib.ID, "en-US", BrowseFilters{
		Genre: "Action", Sort: "title", Limit: 5,
	})
	require.NoError(t, err)
	require.Equal(t, 1, page.Total)

	got, err := libRepo.GetByID(ctx, lib.ID)
	require.NoError(t, err)
	require.Equal(t, lib.Name, got.Name)
}

func TestMediaRepo_ListAndGetNotFound(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	libRepo := NewLibraryRepo(d.Querier())
	mediaRepo := NewMediaRepo(d.Querier())
	dir := t.TempDir()
	lib, err := libRepo.Create(ctx, "ListLib", LibraryKindMovie, []string{dir}, false)
	require.NoError(t, err)
	itemID, err := mediaRepo.CreateItem(ctx, lib.ID, LibraryKindMovie, "Listed", nil)
	require.NoError(t, err)
	require.NoError(t, mediaRepo.SetLocalizedText(ctx, itemID, "en-US", "title", "Listed"))

	items, err := mediaRepo.ListItemsByLibrary(ctx, lib.ID, "en-US", 10, 0)
	require.NoError(t, err)
	require.Len(t, items, 1)

	_, err = mediaRepo.GetItemByID(ctx, 99999, "en-US")
	require.Error(t, err)
}

func TestMediaRepo_FileLookupNotFound(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })
	repo := NewMediaRepo(d.Querier())

	_, err := repo.GetFileByID(ctx, 99999)
	require.Error(t, err)

	_, err = repo.GetPrimaryFileByItemID(ctx, 99999)
	require.Error(t, err)

	_, err = repo.GetTechnicalByFileID(ctx, 99999)
	require.Error(t, err)

	_, err = repo.GetFileByPath(ctx, "/no/such/file.mkv")
	require.Error(t, err)
}

func TestLibraryAndScanRepo_NotFound(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	_, err := NewLibraryRepo(d.Querier()).GetByID(ctx, 99999)
	require.Error(t, err)

	_, err = NewScanRepo(d.Querier()).GetByID(ctx, 99999)
	require.Error(t, err)
}

func TestMediaRepo_UpsertArtworkReplaceLocalPath(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	libRepo := NewLibraryRepo(d.Querier())
	mediaRepo := NewMediaRepo(d.Querier())
	dir := t.TempDir()
	lib, err := libRepo.Create(ctx, "Art2", LibraryKindMovie, []string{dir}, false)
	require.NoError(t, err)
	itemID, err := mediaRepo.CreateItem(ctx, lib.ID, LibraryKindMovie, "Art Replace", nil)
	require.NoError(t, err)

	require.NoError(t, mediaRepo.UpsertArtwork(ctx, itemID, "poster", "https://a.com/1.jpg", ""))
	require.NoError(t, mediaRepo.UpsertArtwork(ctx, itemID, "poster", "https://a.com/2.jpg", "/local/p.jpg"))
}

func TestRequestRepo_HasActiveExcludesCompleted(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	users := NewUserRepo(d.Querier())
	userID := uuid.NewString()
	_, err := users.Create(ctx, userID, "active-user", "hash", []string{"user"})
	require.NoError(t, err)

	repo := NewRequestRepo(d.Querier())
	id, err := repo.Create(ctx, Request{
		UserID: userID, MediaKind: RequestMediaKindMovie,
		Provider: "tmdb", ExternalID: "active-1", Title: "Done",
		Status: RequestStatusCompleted,
	})
	require.NoError(t, err)
	_ = id

	active, err := repo.HasActiveByProviderExternal(ctx, "tmdb", "active-1")
	require.NoError(t, err)
	require.False(t, active)
}
