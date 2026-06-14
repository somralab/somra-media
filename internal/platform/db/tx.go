package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// ErrTxRollback is returned (joined with the original error) when fn passed
// to WithTx returns an error, so callers can disambiguate rollback failures
// from the underlying business error with errors.Is.
var ErrTxRollback = errors.New("db: transaction rolled back")

// WithTx runs fn inside a single SQL transaction. If fn returns an error the
// transaction is rolled back and the returned error wraps both ErrTxRollback
// and fn's error; otherwise the transaction is committed.
//
// Commit failures are wrapped with %w so callers can inspect them.
// Rollback failures are joined with the original error to avoid silently
// dropping information.
func WithTx(ctx context.Context, d *DB, fn func(Querier) error) error {
	if d == nil || d.sqlDB == nil {
		return errors.New("db tx: nil database handle")
	}
	if fn == nil {
		return errors.New("db tx: nil callback")
	}

	tx, err := d.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("db tx: begin: %w", err)
	}

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
			return errors.Join(ErrTxRollback, err, fmt.Errorf("db tx: rollback: %w", rbErr))
		}
		return errors.Join(ErrTxRollback, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("db tx: commit: %w", err)
	}
	return nil
}
