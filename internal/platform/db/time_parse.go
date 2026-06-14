package db

import "time"

// parseSQLiteTime parses timestamps stored via SQLite datetime('now').
func parseSQLiteTime(raw string) time.Time {
	if raw == "" {
		return time.Time{}
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t
	}
	if t, err := time.Parse("2006-01-02 15:04:05", raw); err == nil {
		return t
	}
	return time.Time{}
}
