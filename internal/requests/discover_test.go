package requests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/metadata"
)

type stubDiscoverProvider struct {
	name    string
	results []metadata.SearchResult
	err     error
}

func (s stubDiscoverProvider) Name() string {
	if s.name != "" {
		return s.name
	}
	return "tmdb"
}

func (s stubDiscoverProvider) Search(_ context.Context, _ metadata.SearchQuery) ([]metadata.SearchResult, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.results, nil
}

func (s stubDiscoverProvider) Detail(_ context.Context, _, _ string) (metadata.Detail, error) {
	return metadata.Detail{}, nil
}

func (s stubDiscoverProvider) Images(_ context.Context, _ string) (string, string, string, error) {
	return "", "", "", nil
}

func TestDiscoverer_Search_FiltersLibrary(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	reg := metadata.NewRegistry()
	reg.Register(stubDiscoverProvider{results: []metadata.SearchResult{
		{Provider: "tmdb", ExternalID: "111", Title: "New Movie"},
		{Provider: "tmdb", ExternalID: "222", Title: "Owned"},
	}})

	d := &Discoverer{Registry: reg, Library: mapLibraryLookup{items: map[string]int64{"tmdb:222": 99}}}
	hits, err := d.Search(ctx, DiscoverSearchParams{Query: "movie", Kind: MediaKindMovie, Locale: "en-US"})
	require.NoError(t, err)
	require.Len(t, hits, 1)
	assert.Equal(t, "111", hits[0].ExternalID)
	assert.Equal(t, "New Movie", hits[0].Title)
}

func TestDiscoverer_Search_Limit(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	results := make([]metadata.SearchResult, 5)
	for i := range results {
		results[i] = metadata.SearchResult{Provider: "tmdb", ExternalID: "1000", Title: "Title"}
	}
	reg := metadata.NewRegistry()
	reg.Register(stubDiscoverProvider{results: results})

	d := &Discoverer{Registry: reg}
	hits, err := d.Search(ctx, DiscoverSearchParams{Query: "x", Kind: MediaKindMovie, Limit: 2})
	require.NoError(t, err)
	assert.Len(t, hits, 2)
}

func TestDiscoverer_Search_MissingProvider(t *testing.T) {
	t.Parallel()
	d := &Discoverer{Registry: metadata.NewRegistry()}
	_, err := d.Search(context.Background(), DiscoverSearchParams{Query: "x", Kind: MediaKindMovie})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not registered")
}

func TestDiscoverer_Search_NilDiscoverer(t *testing.T) {
	t.Parallel()
	var d *Discoverer
	_, err := d.Search(context.Background(), DiscoverSearchParams{Query: "x"})
	require.Error(t, err)
}

type mapLibraryLookup struct {
	items map[string]int64
}

func (m mapLibraryLookup) ExistsByProviderID(_ context.Context, provider, externalID string) (bool, int64, error) {
	key := provider + ":" + externalID
	id, ok := m.items[key]
	return ok, id, nil
}
