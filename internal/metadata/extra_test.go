package metadata

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/platform/db"
)

func TestService_SearchCandidates(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	libRepo := db.NewLibraryRepo(d.Querier())
	mediaRepo := db.NewMediaRepo(d.Querier())
	dir := t.TempDir()
	lib, err := libRepo.Create(ctx, "M", db.LibraryKindMovie, []string{dir}, false)
	require.NoError(t, err)
	itemID, err := mediaRepo.CreateItem(ctx, lib.ID, db.LibraryKindMovie, "Inception", ptrInt(2010))
	require.NoError(t, err)

	reg := NewRegistry()
	reg.Register(TestProvider{})
	svc := &Service{DB: &DBStore{DB: d}, Registry: reg, Matcher: &Matcher{Registry: reg, Limiter: NewRateLimiter(1 * time.Millisecond)}}

	results, err := svc.SearchCandidates(ctx, itemID, "en-US")
	require.NoError(t, err)
	require.NotEmpty(t, results)
}

func TestMatcher_Match(t *testing.T) {
	reg := NewRegistry()
	reg.Register(TestProvider{})
	m := &Matcher{Registry: reg, Limiter: NewRateLimiter(1 * time.Millisecond)}
	results, err := m.Match(context.Background(), SearchQuery{Title: "Inception", Year: ptrInt(2010), Kind: "movie"})
	require.NoError(t, err)
	require.NotEmpty(t, results)
}

func TestStubProviders(t *testing.T) {
	ctx := context.Background()
	for _, p := range []MetadataProvider{&TVDBProvider{}, &MusicBrainzProvider{}, &FanartProvider{}} {
		require.NotEmpty(t, p.Name())
		_, _ = p.Search(ctx, SearchQuery{Title: "x"})
		_, _ = p.Detail(ctx, "1", "en-US")
		_, _, _, _ = p.Images(ctx, "1")
	}
}

func TestDoRequest_BlocksBadURL(t *testing.T) {
	_, err := DoRequest(context.Background(), SafeHTTPClient(time.Second), "GET", "http://127.0.0.1/")
	require.Error(t, err)
}
