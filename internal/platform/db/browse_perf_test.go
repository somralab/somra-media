package db

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// seedLargeLibrary inserts count synthetic movie items for browse/search perf tests.
func seedLargeLibrary(tb testing.TB, d *DB, count int) (libraryID int64) {
	tb.Helper()
	ctx := context.Background()
	libRepo := NewLibraryRepo(d.Querier())
	mediaRepo := NewMediaRepo(d.Querier())

	dir := tb.TempDir()
	lib, err := libRepo.Create(ctx, "Large", LibraryKindMovie, []string{dir}, false)
	require.NoError(tb, err)

	for i := 0; i < count; i++ {
		title := fmt.Sprintf("Movie %05d", i)
		year := 1980 + (i % 40)
		itemID, err := mediaRepo.CreateItem(ctx, lib.ID, LibraryKindMovie, title, &year)
		require.NoError(tb, err)
		require.NoError(tb, mediaRepo.IndexFTS(ctx, itemID, title))
	}

	return lib.ID
}

func largeLibrarySize() int {
	if v := os.Getenv("SOMRA_LARGE_LIB_SIZE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return 2000
}

func explainPlan(t *testing.T, d *DB, query string, args ...any) string {
	t.Helper()
	ctx := context.Background()
	rows, err := d.Querier().QueryContext(ctx, "EXPLAIN QUERY PLAN "+query, args...)
	require.NoError(t, err)
	defer func() { _ = rows.Close() }()
	var parts []string
	for rows.Next() {
		var id, parent, notused int
		var detail string
		require.NoError(t, rows.Scan(&id, &parent, &notused, &detail))
		parts = append(parts, detail)
	}
	require.NoError(t, rows.Err())
	return strings.Join(parts, "\n")
}

func TestBrowseRepo_LargeLibraryQueryPlans(t *testing.T) {
	if testing.Short() {
		t.Skip("large library query plan check")
	}
	d := openTestDB(t)
	libID := seedLargeLibrary(t, d, largeLibrarySize())
	ctx := context.Background()
	browseRepo := NewBrowseRepo(d.Querier())

	listPlan := explainPlan(t, d, `
		SELECT mi.id FROM media_item mi
		WHERE mi.library_id = ?
		ORDER BY mi.created_at DESC
		LIMIT 50 OFFSET 0
	`, libID)
	require.Contains(t, listPlan, "idx_media_item_library_created")

	sortPlan := explainPlan(t, d, `
		SELECT mi.id FROM media_item mi
		WHERE mi.library_id = ?
		ORDER BY mi.sort_title COLLATE NOCASE
		LIMIT 50 OFFSET 0
	`, libID)
	require.Contains(t, sortPlan, "idx_media_item_library_sort")

	_, err := browseRepo.ListPaginated(ctx, libID, "en-US", BrowseFilters{Limit: 50, Sort: "created_at"})
	require.NoError(t, err)

	results, err := browseRepo.SearchFTS(ctx, "Movie", "en-US", 20)
	require.NoError(t, err)
	require.NotEmpty(t, results)
}

func BenchmarkBrowseRepo_ListPaginated(b *testing.B) {
	d := openTestDB(b)
	libID := seedLargeLibrary(b, d, largeLibrarySize())
	ctx := context.Background()
	repo := NewBrowseRepo(d.Querier())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.ListPaginated(ctx, libID, "en-US", BrowseFilters{Limit: 50, Sort: "title"})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBrowseRepo_SearchFTS(b *testing.B) {
	d := openTestDB(b)
	seedLargeLibrary(b, d, largeLibrarySize())
	ctx := context.Background()
	repo := NewBrowseRepo(d.Querier())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.SearchFTS(ctx, "Movie 01234", "en-US", 20)
		if err != nil {
			b.Fatal(err)
		}
	}
}
