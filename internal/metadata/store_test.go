package metadata

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/platform/db"
)

func TestDBStore_GetItemAndApplyMatch(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	store := &DBStore{DB: d}
	libRepo := db.NewLibraryRepo(d.Querier())
	mediaRepo := db.NewMediaRepo(d.Querier())
	dir := t.TempDir()
	lib, err := libRepo.Create(ctx, "X", db.LibraryKindMovie, []string{dir}, false)
	require.NoError(t, err)
	itemID, err := mediaRepo.CreateItem(ctx, lib.ID, db.LibraryKindMovie, "Title", ptrInt(2020))
	require.NoError(t, err)

	view, err := store.GetItem(ctx, itemID, "en-US")
	require.NoError(t, err)
	require.Equal(t, itemID, view.ID)

	require.NoError(t, store.ApplyMatch(ctx, itemID, "tmdb", "5", "en-US", Detail{
		Provider: "tmdb", ExternalID: "5", Title: "Title", Overview: "O", PosterURL: "https://image.tmdb.org/t/p/w500/x.jpg",
	}))

	items, err := store.ListUnmatched(ctx, lib.ID, 10)
	require.NoError(t, err)
	require.Empty(t, items)
}

func TestMockProvider_Detail(t *testing.T) {
	p := &MockProvider{Detail_: Detail{Provider: "mock", ExternalID: "9", Title: "T"}}
	d, err := p.Detail(context.Background(), "9", "en-US")
	require.NoError(t, err)
	require.Equal(t, "T", d.Title)
}
