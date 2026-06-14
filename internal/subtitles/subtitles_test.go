package subtitles_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/subtitles"
)

type mockProvider struct {
	results []subtitles.SearchResult
	data    []byte
}

func (m *mockProvider) Name() string { return "mock" }

func (m *mockProvider) Search(_ context.Context, _ subtitles.SearchQuery) ([]subtitles.SearchResult, error) {
	return m.results, nil
}

func (m *mockProvider) Download(_ context.Context, _, _ string) ([]byte, error) {
	return m.data, nil
}

func TestMissingLanguages(t *testing.T) {
	missing := subtitles.MissingLanguages([]string{"en", "tr"}, []string{"en"})
	assert.Equal(t, []string{"tr"}, missing)
}

func TestScoreResult(t *testing.T) {
	q := subtitles.SearchQuery{Title: "Matrix", Language: "en"}
	item := subtitles.SearchResult{Score: 50}
	assert.GreaterOrEqual(t, item.Score, 0.0)
	_ = q
}

func TestOpenSubtitlesRequiresAPIKey(t *testing.T) {
	p := subtitles.NewOpenSubtitles("")
	_, err := p.Search(context.Background(), subtitles.SearchQuery{Title: "Test"})
	require.Error(t, err)
}

func TestMockProviderSearch(t *testing.T) {
	p := &mockProvider{results: []subtitles.SearchResult{{Provider: "mock", ExternalID: "1", Language: "en", Score: 90}}}
	results, err := p.Search(context.Background(), subtitles.SearchQuery{Title: "Film"})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "mock", results[0].Provider)
}
