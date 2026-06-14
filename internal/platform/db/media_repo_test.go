package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMediaRepo_FullLifecycle(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	defer d.Close()

	libRepo := NewLibraryRepo(d.Querier())
	mediaRepo := NewMediaRepo(d.Querier())
	dir := t.TempDir()
	lib, err := libRepo.Create(ctx, "Movies", LibraryKindMovie, []string{dir}, false)
	require.NoError(t, err)

	year := 2010
	itemID, err := mediaRepo.CreateItem(ctx, lib.ID, LibraryKindMovie, "Inception", &year)
	require.NoError(t, err)

	require.NoError(t, mediaRepo.IndexFTS(ctx, itemID, "Inception"))
	ids, err := mediaRepo.SearchTitleFTS(ctx, "Inception", 10)
	require.NoError(t, err)
	require.Contains(t, ids, itemID)

	require.NoError(t, mediaRepo.SetLocalizedText(ctx, itemID, "en-US", "title", "Inception"))
	require.NoError(t, mediaRepo.SetLocalizedText(ctx, itemID, "en-US", "overview", "A dream within a dream"))
	score := 0.95
	require.NoError(t, mediaRepo.SetMatch(ctx, itemID, MatchManual, &score))
	require.NoError(t, mediaRepo.SetProviderID(ctx, itemID, "tmdb", "27205"))
	require.NoError(t, mediaRepo.UpsertArtwork(ctx, itemID, "poster", "https://image.tmdb.org/t/p/w500/x.jpg", ""))

	err = mediaRepo.SetProviderID(ctx, itemID, "invalid-provider", "1")
	require.Error(t, err)

	fileID, err := mediaRepo.UpsertFile(ctx, MediaFile{
		LibraryID: lib.ID, MediaItemID: &itemID,
		Path: "/tmp/inception.mkv", FileName: "inception.mkv",
		SizeBytes: 1000, MtimeNs: 123, ParsedTitle: "Inception", ParsedYear: &year,
	})
	require.NoError(t, err)

	fileID2, err := mediaRepo.UpsertFile(ctx, MediaFile{
		LibraryID: lib.ID, MediaItemID: &itemID,
		Path: "/tmp/inception.mkv", FileName: "inception.mkv",
		SizeBytes: 2000, MtimeNs: 456, ParsedTitle: "Inception", ParsedYear: &year,
	})
	require.NoError(t, err)
	assert.Equal(t, fileID, fileID2)

	gotFile, err := mediaRepo.GetFileByPath(ctx, "/tmp/inception.mkv")
	require.NoError(t, err)
	assert.Equal(t, int64(2000), gotFile.SizeBytes)

	require.NoError(t, mediaRepo.UpsertTechnical(ctx, fileID, 7200000, "matroska", "h264", 1920, 1080, "aac", 2, 1, `{}`))

	items, err := mediaRepo.ListItemsByLibrary(ctx, lib.ID, "en-US", 10, 0)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "Inception", items[0].Title)
	assert.Equal(t, "A dream within a dream", items[0].Overview)
	assert.NotEmpty(t, items[0].PosterURL)

	item, err := mediaRepo.GetItemByID(ctx, itemID, "en-US")
	require.NoError(t, err)
	assert.Equal(t, MatchManual, item.MatchStatus)
	require.NotNil(t, item.MatchScore)

	require.NoError(t, mediaRepo.SetMatch(ctx, itemID, MatchUnmatched, nil))

	_, err = mediaRepo.GetItemByID(ctx, 99999, "en-US")
	require.Error(t, err)

	_, err = mediaRepo.GetFileByPath(ctx, "/missing.mkv")
	require.Error(t, err)

	_, err = mediaRepo.CreateItem(ctx, 99999, LibraryKindMovie, "orphan", nil)
	require.Error(t, err)

	empty, err := mediaRepo.SearchTitleFTS(ctx, "   ", 5)
	require.NoError(t, err)
	assert.Nil(t, empty)
}

func TestScanRepo_ListByLibrary(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	defer d.Close()

	libRepo := NewLibraryRepo(d.Querier())
	scanRepo := NewScanRepo(d.Querier())
	dir := t.TempDir()
	lib, err := libRepo.Create(ctx, "TV", LibraryKindTV, []string{dir}, true)
	require.NoError(t, err)

	runID, err := scanRepo.CreateRun(ctx, lib.ID, ScanIncremental, "task-2")
	require.NoError(t, err)
	require.NoError(t, scanRepo.Finish(ctx, runID, ScanFailed, "probe error"))

	_, err = scanRepo.CreateRun(ctx, 99999, ScanFull, "bad")
	require.Error(t, err)

	runs, err := scanRepo.ListByLibrary(ctx, lib.ID, 5)
	require.NoError(t, err)
	require.Len(t, runs, 1)
	assert.Equal(t, ScanFailed, runs[0].Status)
}

func TestLibraryRepo_NotFound(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	defer d.Close()
	repo := NewLibraryRepo(d.Querier())
	_, err := repo.GetByID(ctx, 404)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrLibraryNotFound)
	require.Error(t, repo.Delete(ctx, 404))
}

func TestMediaRepo_UpsertFileOptionalFields(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	defer d.Close()
	libRepo := NewLibraryRepo(d.Querier())
	mediaRepo := NewMediaRepo(d.Querier())
	dir := t.TempDir()
	lib, err := libRepo.Create(ctx, "M", LibraryKindMusic, []string{dir}, false)
	require.NoError(t, err)

	itemID, err := mediaRepo.CreateItem(ctx, lib.ID, LibraryKindMusic, "Album", nil)
	require.NoError(t, err)
	season, episode := 1, 2
	_, err = mediaRepo.UpsertFile(ctx, MediaFile{
		LibraryID: lib.ID, MediaItemID: &itemID,
		Path: "/tmp/track.flac", FileName: "track.flac", SizeBytes: 1,
		ParsedSeason: &season, ParsedEpisode: &episode, ContentHash: "abc",
	})
	require.NoError(t, err)
}

func TestMediaRepo_ForeignKeyErrors(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	defer d.Close()
	mediaRepo := NewMediaRepo(d.Querier())

	_, err := mediaRepo.UpsertFile(ctx, MediaFile{
		LibraryID: 99999, Path: "/tmp/orphan.mkv", FileName: "orphan.mkv", SizeBytes: 1,
	})
	require.Error(t, err)

	err = mediaRepo.UpsertTechnical(ctx, 99999, 1, "mkv", "h264", 0, 0, "", 0, 0, "")
	require.Error(t, err)
}

func TestScanRepo_GetByIDNotFound(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	defer d.Close()
	scanRepo := NewScanRepo(d.Querier())
	_, err := scanRepo.GetByID(ctx, 9999)
	require.Error(t, err)
}
