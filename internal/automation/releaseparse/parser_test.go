package releaseparse

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/somralab/somra-media/internal/plugin"
)

func TestEnrich(t *testing.T) {
	seeders := 42
	r := Enrich(plugin.SearchResult{
		Title:   "Movie.Name.2024.1080p.WEB-DL.x265-GROUP",
		Seeders: &seeders,
	})
	assert.Equal(t, "1080p", r.Resolution)
	assert.Equal(t, "hevc", r.Codec)
	assert.Equal(t, "WEB-DL", r.Source)
	assert.Greater(t, ScoreHint(r), 30)
}

func TestEnrichPreservesExisting(t *testing.T) {
	r := Enrich(plugin.SearchResult{
		Title:      "anything",
		Resolution: "720p",
		Codec:      "h264",
		Source:     "HDTV",
	})
	assert.Equal(t, "720p", r.Resolution)
	assert.Equal(t, "h264", r.Codec)
	assert.Equal(t, "HDTV", r.Source)
}

func TestScoreHintResolution(t *testing.T) {
	assert.Greater(t, ScoreHint(plugin.SearchResult{Resolution: "1080p"}), ScoreHint(plugin.SearchResult{Resolution: "720p"}))
}

func TestEnrichEmptyTitle(t *testing.T) {
	r := Enrich(plugin.SearchResult{})
	assert.Empty(t, r.Resolution)
}

func TestEnrich4KAndRemux(t *testing.T) {
	r := Enrich(plugin.SearchResult{Title: "Movie.2024.4K.REMUX.x265-GROUP"})
	assert.Equal(t, "4k", r.Resolution)
	assert.Equal(t, "hevc", r.Codec)
	assert.Equal(t, "REMUX", r.Source)
	assert.Greater(t, ScoreHint(r), 40)
}

func TestEnrichH264AndBluray(t *testing.T) {
	r := Enrich(plugin.SearchResult{Title: "Show.S01E01.720p.BluRay.x264-GROUP"})
	assert.Equal(t, "720p", r.Resolution)
	assert.Equal(t, "h264", r.Codec)
	assert.Equal(t, "BLURAY", r.Source)
}

func TestScoreHintSeedersCap(t *testing.T) {
	high := 100
	low := 5
	assert.Equal(t, ScoreHint(plugin.SearchResult{Seeders: &high}), ScoreHint(plugin.SearchResult{Seeders: &low})+45)
}
