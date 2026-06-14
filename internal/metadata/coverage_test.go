package metadata

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/platform/db"
)

func TestScoreMatch_EdgeCases(t *testing.T) {
	assert.Equal(t, float64(0), ScoreMatch(SearchQuery{}, "Title", nil))
	assert.Equal(t, float64(0.1), ScoreMatch(SearchQuery{Title: "Alpha"}, "Beta", nil))

	year := 1999
	assert.Equal(t, float64(0.4), ScoreMatch(SearchQuery{Title: "X", Year: &year}, "Y", &year))
}

func TestSortResults_OrdersByScore(t *testing.T) {
	results := []SearchResult{
		{Title: "low", Score: 0.2},
		{Title: "high", Score: 0.9},
		{Title: "mid", Score: 0.5},
	}
	sortResults(results)
	assert.Equal(t, "high", results[0].Title)
}

func TestRegistry_GetAll(t *testing.T) {
	reg := NewRegistry()
	reg.Register(TestProvider{})
	reg.Register(&MockProvider{})
	got, ok := reg.Get("tmdb")
	require.True(t, ok)
	assert.Equal(t, "tmdb", got.Name())
	assert.Len(t, reg.All(), 2)
}

func TestResponseCache_Expired(t *testing.T) {
	c := NewResponseCache(time.Millisecond)
	c.Set("k", []byte("v"))
	time.Sleep(2 * time.Millisecond)
	_, ok := c.Get("k")
	assert.False(t, ok)
}

func TestRateLimiter_ContextCancel(t *testing.T) {
	l := NewRateLimiter(time.Second)
	ctx := context.Background()
	require.NoError(t, l.Wait(ctx, "tmdb"))

	ctx2, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- l.Wait(ctx2, "tmdb") }()
	time.Sleep(20 * time.Millisecond)
	cancel()
	require.Error(t, <-done)
}

func TestDBStore_ListUnmatched(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	store := &DBStore{DB: d}
	libRepo := db.NewLibraryRepo(d.Querier())
	mediaRepo := db.NewMediaRepo(d.Querier())
	dir := t.TempDir()
	lib, err := libRepo.Create(ctx, "U", db.LibraryKindMovie, []string{dir}, false)
	require.NoError(t, err)
	itemID, err := mediaRepo.CreateItem(ctx, lib.ID, db.LibraryKindMovie, "Unmatched", ptrInt(2021))
	require.NoError(t, err)
	_, err = mediaRepo.UpsertFile(ctx, db.MediaFile{
		LibraryID: lib.ID, MediaItemID: &itemID,
		Path: "/tmp/u.mkv", FileName: "u.mkv", ParsedTitle: "Unmatched",
	})
	require.NoError(t, err)

	items, err := store.ListUnmatched(ctx, lib.ID, 10)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "Unmatched", items[0].ParsedTitle)
}

func TestDBStore_ApplyMatchWithBackdrop(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	store := &DBStore{DB: d}
	libRepo := db.NewLibraryRepo(d.Querier())
	mediaRepo := db.NewMediaRepo(d.Querier())
	dir := t.TempDir()
	lib, err := libRepo.Create(ctx, "B", db.LibraryKindMovie, []string{dir}, false)
	require.NoError(t, err)
	itemID, err := mediaRepo.CreateItem(ctx, lib.ID, db.LibraryKindMovie, "T", nil)
	require.NoError(t, err)

	err = store.ApplyMatch(ctx, itemID, "tmdb", "1", "tr-TR", Detail{
		Provider: "tmdb", ExternalID: "1",
		Title: "T", Overview: "O",
		PosterURL:   "https://image.tmdb.org/t/p/w500/p.jpg",
		BackdropURL: "https://image.tmdb.org/t/p/original/b.jpg",
	})
	require.NoError(t, err)
}

func TestService_AutoMatchSkipsLowScore(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	libRepo := db.NewLibraryRepo(d.Querier())
	mediaRepo := db.NewMediaRepo(d.Querier())
	dir := t.TempDir()
	lib, err := libRepo.Create(ctx, "L", db.LibraryKindMovie, []string{dir}, false)
	require.NoError(t, err)
	itemID, err := mediaRepo.CreateItem(ctx, lib.ID, db.LibraryKindMovie, "Obscure", ptrInt(1900))
	require.NoError(t, err)
	_, err = mediaRepo.UpsertFile(ctx, db.MediaFile{
		LibraryID: lib.ID, MediaItemID: &itemID,
		Path: "/tmp/o.mkv", FileName: "o.mkv", ParsedTitle: "Obscure",
	})
	require.NoError(t, err)

	reg := NewRegistry()
	reg.Register(&MockProvider{Results: []SearchResult{{
		Provider: "mock", ExternalID: "1", Title: "Other", Score: 0.1,
	}}})
	svc := &Service{
		DB:       &DBStore{DB: d},
		Registry: reg,
		Matcher:  &Matcher{Registry: reg},
	}

	n, err := svc.AutoMatch(ctx, lib.ID, "en-US", 10)
	require.NoError(t, err)
	assert.Equal(t, 0, n)
}

func TestMockProvider_SearchWithPresetResults(t *testing.T) {
	p := &MockProvider{Results: []SearchResult{{Provider: "mock", ExternalID: "9", Title: "Preset", Score: 0.99}}}
	results, err := p.Search(context.Background(), SearchQuery{Title: "x"})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "Preset", results[0].Title)

	poster, _, _, err := p.Images(context.Background(), "9")
	require.NoError(t, err)
	assert.NotEmpty(t, poster)
}

func TestTestProvider_Images(t *testing.T) {
	_, _, _, err := TestProvider{}.Images(context.Background(), "1")
	require.NoError(t, err)
}

func TestMatcher_ProviderErrorContinues(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&errorProvider{})
	reg.Register(TestProvider{})
	m := &Matcher{Registry: reg}
	results, err := m.Match(context.Background(), SearchQuery{Title: "Inception", Kind: "movie"})
	require.NoError(t, err)
	require.NotEmpty(t, results)
}

type errorProvider struct{}

func (errorProvider) Name() string { return "err" }

func (errorProvider) Search(context.Context, SearchQuery) ([]SearchResult, error) {
	return nil, assert.AnError
}

func (errorProvider) Detail(context.Context, string, string) (Detail, error) {
	return Detail{}, assert.AnError
}

func (errorProvider) Images(context.Context, string) (string, string, string, error) {
	return "", "", "", assert.AnError
}

func TestService_RematchUnknownProvider(t *testing.T) {
	svc := &Service{Registry: NewRegistry()}
	err := svc.Rematch(context.Background(), 1, "missing", "1", "en-US")
	require.Error(t, err)
}
