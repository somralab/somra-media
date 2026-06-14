package db

import (
	"context"
	"database/sql"
	"fmt"

	// modernc.org/sqlite registers itself with database/sql on init; the
	// blank import is the documented way to wire the pure-Go driver.
	_ "modernc.org/sqlite"
)

// driverName is the database/sql driver name registered by modernc.org/sqlite.
// Keeping it as a package constant avoids string literal drift across the
// package and makes it easy to swap with an alternative pure-Go driver in
// the future.
const driverName = "sqlite"

// pragmas lists the SQLite PRAGMA statements applied to every new
// connection. They are evaluated in order on Open so a failure to apply any
// of them surfaces immediately as a startup error.
//
// Rationale per statement:
//   - journal_mode=WAL: enables concurrent readers alongside a single writer
//     which matches Somra's read-heavy workload (browsing, metadata reads).
//   - synchronous=NORMAL: pairs with WAL to balance durability and write
//     throughput on consumer hardware. FULL is overkill for a home server.
//   - foreign_keys=ON: enforces relational integrity; SQLite ships with this
//     disabled by default which is a footgun.
//   - busy_timeout=5000: blocks instead of returning SQLITE_BUSY immediately
//     when another connection holds a lock, smoothing over short writer
//     contention without the caller having to retry.
//   - temp_store=MEMORY: keeps transient sort/index data in RAM; the data
//     dir is the only thing we want persisted to disk.
//   - cache_size=-65536: negative values are interpreted as KiB; 64 MiB is
//     a safe default for the small-server profile.
var pragmas = []string{
	"PRAGMA journal_mode=WAL",
	"PRAGMA synchronous=NORMAL",
	"PRAGMA foreign_keys=ON",
	"PRAGMA busy_timeout=5000",
	"PRAGMA temp_store=MEMORY",
	"PRAGMA cache_size=-65536",
}

// DB wraps a *sql.DB so the rest of the application can depend on this
// package's higher-level interface (Querier, transactions, migrations)
// without leaking the underlying driver type.
type DB struct {
	sqlDB *sql.DB
}

// Open dials the SQLite database described by cfg, applies the required
// PRAGMAs and connection pool limits, and returns a ready handle. The
// PRAGMA queries themselves exercise the connection so a separate ping is
// not needed.
//
// SQLite serialises all writes through a single writer; multiple Go
// goroutines opening write transactions in parallel will queue, not run
// concurrently. We therefore cap MaxOpenConns at a small number that is
// still high enough to keep readers responsive under WAL mode. Connection
// lifetime is intentionally unbounded (0) because SQLite connections are
// cheap and have no remote keepalive concerns; recycling them would only
// throw away cached statements.
func Open(ctx context.Context, cfg Config) (*DB, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("db open: %w", err)
	}
	if err := cfg.EnsureDataDir(); err != nil {
		return nil, fmt.Errorf("db open: %w", err)
	}
	dsn, err := cfg.DSN()
	if err != nil {
		return nil, fmt.Errorf("db open: %w", err)
	}

	sqlDB, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("db open: %w", err)
	}

	sqlDB.SetMaxOpenConns(8)
	sqlDB.SetMaxIdleConns(4)
	sqlDB.SetConnMaxLifetime(0)

	if err := applyPragmas(ctx, sqlDB); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("db open: %w", err)
	}

	return &DB{sqlDB: sqlDB}, nil
}

// SQL exposes the underlying *sql.DB. Higher layers should prefer the
// Querier interface and transaction helpers in this package, but the raw
// handle is occasionally needed for things like exposing health stats or
// passing to libraries that require database/sql directly.
func (d *DB) SQL() *sql.DB { return d.sqlDB }

// Close releases the connection pool. It is safe to call on a nil or
// uninitialised receiver and may be called more than once; subsequent
// calls return the error from database/sql, if any.
func (d *DB) Close() error {
	if d == nil || d.sqlDB == nil {
		return nil
	}
	return d.sqlDB.Close()
}

// Ping verifies that the database is reachable.
func (d *DB) Ping(ctx context.Context) error {
	if err := d.sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("db ping: %w", err)
	}
	return nil
}

// JournalMode returns the active journal_mode reported by SQLite. It is
// primarily useful for diagnostics and tests that need to assert WAL is on.
func (d *DB) JournalMode(ctx context.Context) (string, error) {
	var mode string
	if err := d.sqlDB.QueryRowContext(ctx, "PRAGMA journal_mode").Scan(&mode); err != nil {
		return "", fmt.Errorf("db journal_mode: %w", err)
	}
	return mode, nil
}

// applyPragmas runs every entry in pragmas against the connection. Because
// some PRAGMAs (e.g. journal_mode) return a row, we use QueryContext and
// drain the rows rather than ExecContext.
func applyPragmas(ctx context.Context, sqlDB *sql.DB) error {
	for _, p := range pragmas {
		rows, err := sqlDB.QueryContext(ctx, p)
		if err != nil {
			return fmt.Errorf("apply pragma %q: %w", p, err)
		}
		for rows.Next() {
			// drain rows so the connection is returned to the pool
		}
		err = rows.Err()
		if cerr := rows.Close(); cerr != nil && err == nil {
			err = cerr
		}
		if err != nil {
			return fmt.Errorf("apply pragma %q: %w", p, err)
		}
	}
	return nil
}
