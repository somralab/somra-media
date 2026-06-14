package metadata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScoreMatch_ExactTitleAndYear(t *testing.T) {
	year := 2020
	q := SearchQuery{Title: "Inception", Year: &year}
	score := ScoreMatch(q, "Inception", &year)
	assert.GreaterOrEqual(t, score, 1.0)
}

func TestScoreMatch_PartialTitle(t *testing.T) {
	q := SearchQuery{Title: "Matrix"}
	score := ScoreMatch(q, "The Matrix", nil)
	assert.GreaterOrEqual(t, score, 0.7)
}

func TestValidateOutboundURL_BlocksPrivate(t *testing.T) {
	err := ValidateOutboundURL("https://127.0.0.1/movie")
	assert.Error(t, err)
}

func TestValidateOutboundURL_BlocksUnknownHost(t *testing.T) {
	err := ValidateOutboundURL("https://evil.example/movie")
	assert.Error(t, err)
}

func TestValidateOutboundURL_AllowsTMDB(t *testing.T) {
	err := ValidateOutboundURL("https://api.themoviedb.org/3/search/movie")
	assert.NoError(t, err)
}

func TestMockProvider_Search(t *testing.T) {
	p := &MockProvider{}
	results, err := p.Search(t.Context(), SearchQuery{Title: "Test Movie", Year: ptr(2020)})
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "mock", results[0].Provider)
}

func ptr(v int) *int { return &v }
