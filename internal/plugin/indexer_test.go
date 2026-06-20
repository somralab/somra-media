package plugin

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockIndexer struct {
	id           string
	capabilities Capabilities
	results      []SearchResult
	capErr       error
	searchErr    error
}

func (m *mockIndexer) ID() string              { return m.id }
func (m *mockIndexer) Type() PluginType        { return PluginTypeIndexer }
func (m *mockIndexer) ContractVersion() string { return ContractVersion }

func (m *mockIndexer) Capabilities(_ context.Context) (Capabilities, error) {
	if m.capErr != nil {
		return Capabilities{}, m.capErr
	}
	return m.capabilities, nil
}

func (m *mockIndexer) Search(_ context.Context, _ SearchQuery) ([]SearchResult, error) {
	if m.searchErr != nil {
		return nil, m.searchErr
	}
	return m.results, nil
}

var _ Indexer = (*mockIndexer)(nil)

func TestIndexerCapabilities(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	seeders := 10
	idx := &mockIndexer{
		id: "torznab-demo",
		capabilities: Capabilities{
			Protocols:      []Protocol{ProtocolTorrent},
			Categories:     []Category{{ID: 2000, Name: "Movies"}},
			Limits:         Limits{MaxPageSize: 100, MaxResults: 250},
			SupportsSearch: true,
			SupportsRSS:    true,
		},
		results: []SearchResult{{
			ReleaseID:  "rel-1",
			IndexerID:  "torznab-demo",
			Title:      "Example 1080p",
			Protocol:   ProtocolTorrent,
			SizeBytes:  4_000_000_000,
			Seeders:    &seeders,
			Resolution: "1080p",
			Codec:      "x265",
			Source:     "web-dl",
		}},
	}

	require.NoError(t, ValidateContract(idx))

	caps, err := idx.Capabilities(ctx)
	require.NoError(t, err)
	assert.True(t, caps.SupportsSearch)
	assert.Equal(t, []Protocol{ProtocolTorrent}, caps.Protocols)
	assert.Equal(t, 100, caps.Limits.MaxPageSize)

	year := 2024
	results, err := idx.Search(ctx, SearchQuery{
		Title:     "Example",
		Year:      &year,
		MediaKind: MediaKindMovie,
	})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "rel-1", results[0].ReleaseID)
	assert.Equal(t, int64(4_000_000_000), results[0].SizeBytes)
}

func TestIndexerSearchTVQuery(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	season, episode := 1, 5
	idx := &mockIndexer{
		id: "newznab-demo",
		capabilities: Capabilities{
			Protocols:      []Protocol{ProtocolUsenet},
			SupportsSearch: true,
		},
		results: []SearchResult{{
			ReleaseID: "nzb-42",
			IndexerID: "newznab-demo",
			Title:     "Show S01E05",
			Protocol:  ProtocolUsenet,
			SizeBytes: 800_000_000,
		}},
	}

	results, err := idx.Search(ctx, SearchQuery{
		Title:     "Show",
		MediaKind: MediaKindTV,
		Season:    &season,
		Episode:   &episode,
	})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, ProtocolUsenet, results[0].Protocol)
}

func TestIndexerCapabilitiesError(t *testing.T) {
	t.Parallel()
	idx := &mockIndexer{id: "broken", capErr: ErrUnsupportedCapability}
	_, err := idx.Capabilities(context.Background())
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnsupportedCapability)
}

func TestSearchResultPublishAt(t *testing.T) {
	t.Parallel()
	published := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	r := SearchResult{
		ReleaseID: "rel-ts",
		PublishAt: &published,
	}
	assert.Equal(t, "rel-ts", r.ReleaseID)
	require.NotNil(t, r.PublishAt)
	assert.Equal(t, 2024, r.PublishAt.Year())
}
