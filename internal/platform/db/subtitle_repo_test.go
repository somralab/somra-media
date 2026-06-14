package db

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubtitleRepoMultipleItemsMissing(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })
	repo := NewSubtitleRepo(d.Querier())
	libRepo := NewLibraryRepo(d.Querier())
	mediaRepo := NewMediaRepo(d.Querier())
	dir := t.TempDir()
	lib, err := libRepo.Create(ctx, "Films", LibraryKindMovie, []string{dir}, false)
	require.NoError(t, err)

	for _, title := range []string{"A", "B"} {
		itemID, err := mediaRepo.CreateItem(ctx, lib.ID, LibraryKindMovie, title, nil)
		require.NoError(t, err)
		_, err = mediaRepo.UpsertFile(ctx, MediaFile{
			MediaItemID: &itemID,
			LibraryID:   lib.ID,
			Path:        filepath.Join(dir, title+".mkv"),
		})
		require.NoError(t, err)
	}

	ids, err := repo.ListMediaItemsMissingLanguages(ctx, []string{"en"}, 1)
	require.NoError(t, err)
	assert.NotEmpty(t, ids)
}

func TestSubtitleRepo_CRUD(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })
	repo := NewSubtitleRepo(d.Querier())
	libRepo := NewLibraryRepo(d.Querier())
	mediaRepo := NewMediaRepo(d.Querier())

	dir := t.TempDir()
	lib, err := libRepo.Create(ctx, "Films", LibraryKindMovie, []string{dir}, false)
	require.NoError(t, err)
	itemID, err := mediaRepo.CreateItem(ctx, lib.ID, LibraryKindMovie, "Matrix", intPtr(1999))
	require.NoError(t, err)
	_, err = mediaRepo.UpsertFile(ctx, MediaFile{
		MediaItemID: &itemID,
		LibraryID:   lib.ID,
		Path:        filepath.Join(dir, "matrix.mkv"),
	})
	require.NoError(t, err)

	files, err := repo.ListByMediaItem(ctx, itemID)
	require.NoError(t, err)
	assert.Empty(t, files)

	id, err := repo.Create(ctx, SubtitleFile{
		MediaItemID: itemID,
		Language:    "en",
		Source:      SubtitleExternal,
		Path:        "/cache/subtitles/1/en.srt",
		Provider:    "mock",
		ExternalID:  "ext-1",
	})
	require.NoError(t, err)
	require.Greater(t, id, int64(0))

	has, err := repo.HasLanguage(ctx, itemID, "en")
	require.NoError(t, err)
	assert.True(t, has)

	has, err = repo.HasLanguage(ctx, itemID, "EN")
	require.NoError(t, err)
	assert.True(t, has)

	files, err = repo.ListByMediaItem(ctx, itemID)
	require.NoError(t, err)
	require.Len(t, files, 1)
	assert.False(t, files[0].CreatedAt.IsZero())

	ids, err := repo.ListMediaItemsMissingLanguages(ctx, []string{"en", "tr"}, 10)
	require.NoError(t, err)
	assert.Contains(t, ids, itemID)

	_, err = repo.Create(ctx, SubtitleFile{
		MediaItemID: itemID,
		Language:    "tr",
		Source:      SubtitleUploaded,
		Path:        "/cache/subtitles/1/tr.srt",
	})
	require.NoError(t, err)

	ids, err = repo.ListMediaItemsMissingLanguages(ctx, []string{"en", "tr"}, 10)
	require.NoError(t, err)
	assert.NotContains(t, ids, itemID)

	empty, err := repo.ListMediaItemsMissingLanguages(ctx, nil, 10)
	require.NoError(t, err)
	assert.Nil(t, empty)

	files, err = repo.ListByMediaItem(ctx, itemID)
	require.NoError(t, err)
	require.Len(t, files, 2)

	_, err = d.SQL().ExecContext(ctx, `
		INSERT INTO subtitle_file (media_item_id, language, source, path, created_at)
		VALUES (?, 'de', 'uploaded', '/x.srt', 'not-a-date')
	`, itemID)
	require.NoError(t, err)
	files, err = repo.ListByMediaItem(ctx, itemID)
	require.NoError(t, err)
	require.Len(t, files, 3)
	assert.Equal(t, "de", files[2].Language)
	assert.True(t, files[2].CreatedAt.IsZero())

	_, err = repo.Create(ctx, SubtitleFile{
		MediaItemID: 99999,
		Language:    "xx",
		Source:      SubtitleUploaded,
		Path:        "/missing",
	})
	require.Error(t, err)

	has, err = repo.HasLanguage(ctx, 99999, "en")
	require.NoError(t, err)
	assert.False(t, has)
}

func TestSubtitleRepo_ListAfterClose(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	repo := NewSubtitleRepo(d.Querier())
	require.NoError(t, d.Close())
	_, err := repo.ListByMediaItem(ctx, 1)
	require.Error(t, err)
}
