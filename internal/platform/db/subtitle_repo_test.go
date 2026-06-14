package db

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	has, err = repo.HasLanguage(ctx, itemID, "tr")
	require.NoError(t, err)
	assert.False(t, has)

	files, err = repo.ListByMediaItem(ctx, itemID)
	require.NoError(t, err)
	require.Len(t, files, 1)
	assert.Equal(t, "en", files[0].Language)

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
}
