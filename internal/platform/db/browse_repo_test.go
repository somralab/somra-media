package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBrowseRepo_ListPaginatedAndDiscover(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	mediaRepo := NewMediaRepo(d.Querier())
	browseRepo := NewBrowseRepo(d.Querier())

	libRepo := NewLibraryRepo(d.Querier())
	lib, err := libRepo.Create(ctx, "Films", LibraryKindMovie, []string{"/tmp"}, true)
	require.NoError(t, err)
	libID := lib.ID

	itemID, err := mediaRepo.CreateItem(ctx, libID, LibraryKindMovie, "Alpha", intPtr(2020))
	require.NoError(t, err)
	require.NoError(t, mediaRepo.SetLocalizedText(ctx, itemID, "en-US", "title", "Alpha Movie"))
	require.NoError(t, mediaRepo.IndexFTS(ctx, itemID, "Alpha Movie"))

	item2, err := mediaRepo.CreateItem(ctx, libID, LibraryKindMovie, "Beta", intPtr(2021))
	require.NoError(t, err)
	require.NoError(t, mediaRepo.SetLocalizedText(ctx, item2, "en-US", "title", "Beta Movie"))

	page, err := browseRepo.ListPaginated(ctx, libID, "en-US", BrowseFilters{Limit: 10})
	require.NoError(t, err)
	assert.Equal(t, 2, page.Total)
	assert.Len(t, page.Items, 2)

	results, err := browseRepo.SearchFTS(ctx, "Alpha", "en-US", 5)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, itemID, results[0].ID)

	home, err := browseRepo.DiscoverHome(ctx, "user-1", "en-US")
	require.NoError(t, err)
	assert.NotEmpty(t, home.Shelves)

	detail, err := browseRepo.GetDetail(ctx, itemID, "en-US", "")
	require.NoError(t, err)
	assert.Equal(t, "Alpha Movie", detail.Title)
	assert.NotNil(t, detail.Genres)
	assert.NotNil(t, detail.Cast)
	assert.NotNil(t, detail.Images)
	assert.Len(t, detail.Genres, 0)
	assert.Len(t, detail.Cast, 0)
	assert.Len(t, detail.Images, 0)
}

func intPtr(n int) *int { return &n }
