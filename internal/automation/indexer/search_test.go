package indexer

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/automation/automationtest"
	"github.com/somralab/somra-media/internal/plugin"
)

func TestSearchService_RequiresManager(t *testing.T) {
	var svc *SearchService
	_, err := svc.Search(context.Background(), SearchRequest{})
	require.Error(t, err)
}

func TestSearchService_SearchWithStubIndexer(t *testing.T) {
	ctx := context.Background()
	d := automationtest.OpenDB(t)
	mgr := automationtest.NewManager(t, d)
	idxID := automationtest.CreateStubIndexer(t, mgr, "search-idx")

	svc := &SearchService{Manager: mgr}
	resp, err := svc.Search(ctx, SearchRequest{
		Query: plugin.SearchQuery{Title: "Demo Movie", MediaKind: plugin.MediaKindMovie},
	})
	require.NoError(t, err)
	require.NotEmpty(t, resp.Results)

	filtered, err := svc.Search(ctx, SearchRequest{
		Query:      plugin.SearchQuery{Title: "Demo Movie", MediaKind: plugin.MediaKindMovie},
		IndexerIDs: []int64{idxID},
	})
	require.NoError(t, err)
	require.NotEmpty(t, filtered.Results)

	empty, err := svc.Search(ctx, SearchRequest{
		Query:      plugin.SearchQuery{Title: "Demo Movie", MediaKind: plugin.MediaKindMovie},
		IndexerIDs: []int64{99999},
	})
	require.NoError(t, err)
	require.Empty(t, empty.Results)
}

func TestDedupeResults(t *testing.T) {
	in := []plugin.SearchResult{
		{Title: "Same", Protocol: plugin.ProtocolTorrent, SizeBytes: 100},
		{Title: "same", Protocol: plugin.ProtocolTorrent, SizeBytes: 100},
		{Title: "Other", Protocol: plugin.ProtocolUsenet, SizeBytes: 200},
	}
	out := dedupeResults(in)
	require.Len(t, out, 2)
	require.Equal(t, "Same", out[0].Title)
	require.Equal(t, "Other", out[1].Title)
}

func TestDedupeKeyUsesIndexerFields(t *testing.T) {
	key := dedupeKey(plugin.SearchResult{
		Title:     " Release ",
		SizeBytes: 42,
		Protocol:  plugin.ProtocolTorrent,
		IndexerID: strconv.Itoa(1),
	})
	require.Contains(t, key, "release")
}
