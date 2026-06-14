package metadata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLocaleToTMDB(t *testing.T) {
	assert.Equal(t, "tr-TR", localeToTMDB("tr-TR"))
	assert.Equal(t, "en-US", localeToTMDB("en-US"))
}

func TestParseYear(t *testing.T) {
	y := parseYear("2020-01-01", "")
	requireInt(t, 2020, y)
}

func TestNewTMDBProvider(t *testing.T) {
	p := NewTMDBProvider("key", nil)
	assert.Equal(t, "tmdb", p.Name())
}

func requireInt(t *testing.T, want int, got *int) {
	t.Helper()
	if got == nil || *got != want {
		t.Fatalf("want %d got %v", want, got)
	}
}
