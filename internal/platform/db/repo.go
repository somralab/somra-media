package db

import (
	"context"
	"database/sql"
)

// Querier abstracts the subset of *sql.DB / *sql.Tx that repositories need.
//
// Accepting Querier in repository methods lets the same code run inside a
// transaction (via WithTx) and outside of one against the connection pool,
// without duplicating logic.
type Querier interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

// Compile-time assertions that the standard library types satisfy Querier.
// They prevent accidental breakage if a method signature ever changes.
var (
	_ Querier = (*sql.DB)(nil)
	_ Querier = (*sql.Tx)(nil)
)

// Querier returns the underlying *sql.DB as a Querier so callers that do not
// need an explicit transaction can still depend on the interface.
func (d *DB) Querier() Querier { return d.sqlDB }
