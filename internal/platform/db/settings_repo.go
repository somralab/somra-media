package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// ErrSettingNotFound signals that the requested key has no row in the
// settings table. Callers should use errors.Is for matching.
var ErrSettingNotFound = errors.New("db settings: key not found")

// SettingsRepo is a minimal CRUD wrapper around the settings table. It is
// intentionally free of business logic; higher-level packages compose it
// with domain-specific behaviour (e.g. defaults, validation, caching).
type SettingsRepo struct {
	q Querier
}

// NewSettingsRepo returns a repository bound to q. The Querier may be either
// the connection pool (via DB.Querier) or an in-flight transaction.
func NewSettingsRepo(q Querier) *SettingsRepo {
	return &SettingsRepo{q: q}
}

// Get returns the value stored under key, or ErrSettingNotFound if no row
// exists. An empty key is treated as a validation error.
func (r *SettingsRepo) Get(ctx context.Context, key string) (string, error) {
	if strings.TrimSpace(key) == "" {
		return "", fmt.Errorf("db settings get: key must not be empty")
	}
	var value sql.NullString
	err := r.q.QueryRowContext(ctx, `SELECT value FROM settings WHERE key = ?`, key).Scan(&value)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return "", fmt.Errorf("db settings get %q: %w", key, ErrSettingNotFound)
	case err != nil:
		return "", fmt.Errorf("db settings get %q: %w", key, err)
	}
	if !value.Valid {
		return "", nil
	}
	return value.String, nil
}

// Set inserts or updates the value for key, refreshing updated_at on every
// write so consumers can detect changes via the timestamp.
func (r *SettingsRepo) Set(ctx context.Context, key, value string) error {
	if strings.TrimSpace(key) == "" {
		return fmt.Errorf("db settings set: key must not be empty")
	}
	const q = `
		INSERT INTO settings (key, value, updated_at)
		VALUES (?, ?, datetime('now'))
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = datetime('now')
	`
	if _, err := r.q.ExecContext(ctx, q, key, value); err != nil {
		return fmt.Errorf("db settings set %q: %w", key, err)
	}
	return nil
}

// Insert performs a strict insert. It returns the underlying constraint
// error (typically a UNIQUE violation) without translating it, so callers
// can decide how to react.
func (r *SettingsRepo) Insert(ctx context.Context, key, value string) error {
	if strings.TrimSpace(key) == "" {
		return fmt.Errorf("db settings insert: key must not be empty")
	}
	const q = `INSERT INTO settings (key, value, updated_at) VALUES (?, ?, datetime('now'))`
	if _, err := r.q.ExecContext(ctx, q, key, value); err != nil {
		return fmt.Errorf("db settings insert %q: %w", key, err)
	}
	return nil
}

// Delete removes the row stored under key. Deleting a missing key is a
// no-op so callers can treat the operation as idempotent.
func (r *SettingsRepo) Delete(ctx context.Context, key string) error {
	if strings.TrimSpace(key) == "" {
		return fmt.Errorf("db settings delete: key must not be empty")
	}
	if _, err := r.q.ExecContext(ctx, `DELETE FROM settings WHERE key = ?`, key); err != nil {
		return fmt.Errorf("db settings delete %q: %w", key, err)
	}
	return nil
}
