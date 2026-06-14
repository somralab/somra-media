package library

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFileName_MovieWithYear(t *testing.T) {
	p := ParseFileName("The Matrix (1999).mkv")
	assert.Equal(t, "The Matrix", p.Title)
	assert.NotNil(t, p.Year)
	assert.Equal(t, 1999, *p.Year)
}

func TestParseFileName_TVSeasonEpisode(t *testing.T) {
	p := ParseFileName("Breaking Bad S01E02.mkv")
	assert.Equal(t, "Breaking Bad", p.Title)
	assert.NotNil(t, p.Season)
	assert.Equal(t, 1, *p.Season)
	assert.NotNil(t, p.Episode)
	assert.Equal(t, 2, *p.Episode)
}

func TestParseFileName_TVAltPattern(t *testing.T) {
	p := ParseFileName("Show Name 2x05.mkv")
	assert.Equal(t, "Show Name", p.Title)
	assert.NotNil(t, p.Season)
	assert.Equal(t, 2, *p.Season)
}

func TestParseFileName_DotSeparated(t *testing.T) {
	p := ParseFileName("Inception.2010.1080p.mkv")
	assert.Contains(t, p.Title, "Inception")
	assert.NotNil(t, p.Year)
	assert.Equal(t, 2010, *p.Year)
}
