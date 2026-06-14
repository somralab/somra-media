package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// LoginAttemptKind identifies whether the lockout key is IP or username.
type LoginAttemptKind string

const (
	LoginAttemptIP       LoginAttemptKind = "ip"
	LoginAttemptUsername LoginAttemptKind = "username"
)

// LoginAttempt tracks failed login counts and lockout windows.
type LoginAttempt struct {
	Identifier  string
	Kind        LoginAttemptKind
	FailedCount int
	LockedUntil *time.Time
	UpdatedAt   time.Time
}

// LoginAttemptRepo persists brute-force counters.
type LoginAttemptRepo struct {
	q Querier
}

// NewLoginAttemptRepo returns a repository bound to q.
func NewLoginAttemptRepo(q Querier) *LoginAttemptRepo {
	return &LoginAttemptRepo{q: q}
}

// Get returns the attempt record for identifier+kind, or zero values if absent.
func (r *LoginAttemptRepo) Get(ctx context.Context, identifier string, kind LoginAttemptKind) (LoginAttempt, error) {
	var la LoginAttempt
	var locked sql.NullString
	var updated string
	err := r.q.QueryRowContext(ctx, `
		SELECT identifier, kind, failed_count, locked_until, updated_at
		FROM login_attempt WHERE identifier = ? AND kind = ?
	`, identifier, kind).Scan(&la.Identifier, &la.Kind, &la.FailedCount, &locked, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return LoginAttempt{Identifier: identifier, Kind: kind}, nil
	}
	if err != nil {
		return LoginAttempt{}, fmt.Errorf("db login attempt get: %w", err)
	}
	if locked.Valid {
		t := parseSQLiteTime(locked.String)
		la.LockedUntil = &t
	}
	la.UpdatedAt = parseSQLiteTime(updated)
	return la, nil
}

// RecordFailure increments the failure counter and optionally sets lockout.
func (r *LoginAttemptRepo) RecordFailure(ctx context.Context, identifier string, kind LoginAttemptKind, lockAfter int, lockDuration time.Duration) (LoginAttempt, error) {
	la, err := r.Get(ctx, identifier, kind)
	if err != nil {
		return LoginAttempt{}, err
	}
	la.FailedCount++
	var lockedUntil *time.Time
	if lockAfter > 0 && la.FailedCount >= lockAfter {
		t := time.Now().Add(lockDuration)
		lockedUntil = &t
	}
	_, err = r.q.ExecContext(ctx, `
		INSERT INTO login_attempt (identifier, kind, failed_count, locked_until, updated_at)
		VALUES (?, ?, ?, ?, datetime('now'))
		ON CONFLICT(identifier, kind) DO UPDATE SET
			failed_count = excluded.failed_count,
			locked_until = excluded.locked_until,
			updated_at = datetime('now')
	`, identifier, kind, la.FailedCount, nullTime(lockedUntil))
	if err != nil {
		return LoginAttempt{}, fmt.Errorf("db login attempt record failure: %w", err)
	}
	la.LockedUntil = lockedUntil
	la.UpdatedAt = time.Now()
	return la, nil
}

// Reset clears failure state after a successful login.
func (r *LoginAttemptRepo) Reset(ctx context.Context, identifier string, kind LoginAttemptKind) error {
	_, err := r.q.ExecContext(ctx, `
		DELETE FROM login_attempt WHERE identifier = ? AND kind = ?
	`, identifier, kind)
	if err != nil {
		return fmt.Errorf("db login attempt reset: %w", err)
	}
	return nil
}
