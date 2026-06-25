package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLibraryRepo_CRUD(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	defer d.Close()

	dir := t.TempDir()
	repo := NewLibraryRepo(d.Querier())

	lib, err := repo.Create(ctx, "Movies", LibraryKindMovie, []string{dir}, true)
	require.NoError(t, err)
	assert.Equal(t, "Movies", lib.Name)
	assert.Equal(t, LibraryKindMovie, lib.Kind)
	assert.Len(t, lib.Paths, 1)

	got, err := repo.GetByID(ctx, lib.ID)
	require.NoError(t, err)
	assert.Equal(t, lib.Name, got.Name)

	paths, err := repo.ListPaths(ctx, lib.ID)
	require.NoError(t, err)
	assert.Equal(t, []string{dir}, paths)

	libs, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Len(t, libs, 1)

	dir2 := t.TempDir()
	updated, err := repo.Update(ctx, lib.ID, "Films", []string{dir, dir2}, false)
	require.NoError(t, err)
	assert.Equal(t, "Films", updated.Name)
	assert.False(t, updated.WatchEnabled)
	assert.Len(t, updated.Paths, 2)

	require.NoError(t, repo.Delete(ctx, lib.ID))
	_, err = repo.GetByID(ctx, lib.ID)
	assert.ErrorIs(t, err, ErrLibraryNotFound)
}

func TestMediaRepo_UpsertFile(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	defer d.Close()

	libRepo := NewLibraryRepo(d.Querier())
	mediaRepo := NewMediaRepo(d.Querier())
	dir := t.TempDir()
	lib, err := libRepo.Create(ctx, "TV", LibraryKindTV, []string{dir}, true)
	require.NoError(t, err)

	itemID, err := mediaRepo.CreateItem(ctx, lib.ID, LibraryKindTV, "Show", nil)
	require.NoError(t, err)

	fileID, err := mediaRepo.UpsertFile(ctx, MediaFile{
		LibraryID: lib.ID, MediaItemID: &itemID,
		Path: "/tmp/show.mkv", FileName: "show.mkv", SizeBytes: 100,
	})
	require.NoError(t, err)
	assert.Positive(t, fileID)

	require.NoError(t, mediaRepo.UpsertTechnical(ctx, fileID, 120000, "matroska", "h264", 1920, 1080, "aac", 2, 1, `{}`))
}

func TestScanRepo_Lifecycle(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	defer d.Close()

	libRepo := NewLibraryRepo(d.Querier())
	scanRepo := NewScanRepo(d.Querier())
	dir := t.TempDir()
	lib, err := libRepo.Create(ctx, "Music", LibraryKindMusic, []string{dir}, true)
	require.NoError(t, err)

	runID, err := scanRepo.CreateRun(ctx, lib.ID, ScanFull, "task-1")
	require.NoError(t, err)
	require.NoError(t, scanRepo.MarkRunning(ctx, runID))
	require.NoError(t, scanRepo.UpdateProgress(ctx, runID, 10, 5))
	require.NoError(t, scanRepo.Finish(ctx, runID, ScanSucceeded, ""))

	run, err := scanRepo.GetByID(ctx, runID)
	require.NoError(t, err)
	assert.Equal(t, ScanSucceeded, run.Status)
	assert.Equal(t, 10, run.FilesTotal)
}

func TestLibraryRepo_CreateValidation(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	defer d.Close()
	repo := NewLibraryRepo(d.Querier())
	dir := t.TempDir()

	_, err := repo.Create(ctx, "", LibraryKindMovie, []string{dir}, false)
	require.Error(t, err)

	_, err = repo.Create(ctx, "Name", LibraryKindMovie, nil, false)
	require.Error(t, err)
}

func TestLibraryRepo_UpdateNotFound(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	defer d.Close()
	repo := NewLibraryRepo(d.Querier())
	dir := t.TempDir()

	_, err := repo.Update(ctx, 9999, "X", []string{dir}, false)
	require.Error(t, err)
}

func TestMediaRepo_ListItemsDefaultLimit(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	defer d.Close()
	libRepo := NewLibraryRepo(d.Querier())
	mediaRepo := NewMediaRepo(d.Querier())
	dir := t.TempDir()
	lib, err := libRepo.Create(ctx, "L", LibraryKindMovie, []string{dir}, false)
	require.NoError(t, err)
	itemID, err := mediaRepo.CreateItem(ctx, lib.ID, LibraryKindMovie, "Title", nil)
	require.NoError(t, err)
	_, err = mediaRepo.UpsertFile(ctx, MediaFile{
		LibraryID: lib.ID, MediaItemID: &itemID,
		Path: "/tmp/t.mkv", FileName: "t.mkv", SizeBytes: 1,
	})
	require.NoError(t, err)

	items, err := mediaRepo.ListItemsByLibrary(ctx, lib.ID, "en-US", 0, 0)
	require.NoError(t, err)
	require.Len(t, items, 1)
}

func TestLibraryRepo_UpdateValidation(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	defer d.Close()
	repo := NewLibraryRepo(d.Querier())
	dir := t.TempDir()
	lib, err := repo.Create(ctx, "A", LibraryKindMovie, []string{dir}, false)
	require.NoError(t, err)

	_, err = repo.Update(ctx, lib.ID, "", []string{dir}, false)
	require.Error(t, err)

	_, err = repo.Update(ctx, lib.ID, "B", nil, false)
	require.Error(t, err)
}

func TestLibraryRepo_ListMultiple(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	defer d.Close()
	repo := NewLibraryRepo(d.Querier())
	dir1 := t.TempDir()
	dir2 := t.TempDir()
	_, err := repo.Create(ctx, "One", LibraryKindMovie, []string{dir1}, false)
	require.NoError(t, err)
	_, err = repo.Create(ctx, "Two", LibraryKindTV, []string{dir2}, true)
	require.NoError(t, err)

	libs, err := repo.List(ctx)
	require.NoError(t, err)
	require.Len(t, libs, 2)
}

func TestScanRepo_DefaultLimit(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	defer d.Close()
	libRepo := NewLibraryRepo(d.Querier())
	scanRepo := NewScanRepo(d.Querier())
	dir := t.TempDir()
	lib, err := libRepo.Create(ctx, "S", LibraryKindMusic, []string{dir}, false)
	require.NoError(t, err)
	runID, err := scanRepo.CreateRun(ctx, lib.ID, ScanFull, "t")
	require.NoError(t, err)
	require.NoError(t, scanRepo.MarkRunning(ctx, runID))

	runs, err := scanRepo.ListByLibrary(ctx, lib.ID, 0)
	require.NoError(t, err)
	require.Len(t, runs, 1)
}

func TestMediaRepo_UpsertArtworkReplace(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	defer d.Close()
	libRepo := NewLibraryRepo(d.Querier())
	mediaRepo := NewMediaRepo(d.Querier())
	dir := t.TempDir()
	lib, err := libRepo.Create(ctx, "A", LibraryKindMovie, []string{dir}, false)
	require.NoError(t, err)
	itemID, err := mediaRepo.CreateItem(ctx, lib.ID, LibraryKindMovie, "T", nil)
	require.NoError(t, err)

	require.NoError(t, mediaRepo.UpsertArtwork(ctx, itemID, "poster", "https://example.com/a.jpg", ""))
	require.NoError(t, mediaRepo.UpsertArtwork(ctx, itemID, "poster", "https://example.com/b.jpg", "/local/b.jpg"))
}

func openTestDB(tb testing.TB) *DB {
	tb.Helper()
	cfg := Default()
	cfg.DataDir = tb.TempDir()
	ctx := context.Background()
	d, err := Initialize(ctx, cfg, nil)
	require.NoError(tb, err)
	tb.Cleanup(func() { _ = d.Close() })
	return d
}
