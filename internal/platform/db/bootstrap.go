package db

import (
	"context"
	"fmt"
	"log/slog"
)

// Initialize opens the SQLite database described by cfg and applies all
// pending migrations in a single call. It is the integration point that
// other packets (notably Paket 1 / cmd/somra/main.go) wire into the binary's
// startup sequence so the rest of the application sees a fully-migrated
// database from the first request onwards.
//
// The caller owns the returned *DB and must call Close on shutdown.
func Initialize(ctx context.Context, cfg Config, logger *slog.Logger) (*DB, error) {
	if logger == nil {
		logger = slog.Default()
	}

	d, err := Open(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("db initialize: %w", err)
	}

	if err := MigrateUp(ctx, d, logger); err != nil {
		_ = d.Close()
		return nil, fmt.Errorf("db initialize: %w", err)
	}

	return d, nil
}
