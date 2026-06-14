package db

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseSQLiteTime(t *testing.T) {
	t.Parallel()
	assert.True(t, parseSQLiteTime("").IsZero())

	rfc := time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC).Format(time.RFC3339)
	assert.Equal(t, 2026, parseSQLiteTime(rfc).Year())

	sqlite := "2026-06-14 12:00:00"
	got := parseSQLiteTime(sqlite)
	assert.Equal(t, 2026, got.Year())
	assert.Equal(t, time.June, got.Month())

	assert.True(t, parseSQLiteTime("not-a-date").IsZero())
}
