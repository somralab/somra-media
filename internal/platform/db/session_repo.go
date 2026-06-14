package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// SessionRecord is a persisted refresh-token session.
type SessionRecord struct {
	ID          string
	UserID      string
	DeviceLabel string
	TokenHash   string
	ExpiresAt   time.Time
	RevokedAt   *time.Time
	CreatedAt   time.Time
	LastUsedAt  *time.Time
}

// SessionRepo persists refresh-token sessions.
type SessionRepo struct {
	q Querier
}

// NewSessionRepo returns a repository bound to q.
func NewSessionRepo(q Querier) *SessionRepo {
	return &SessionRepo{q: q}
}

var ErrSessionNotFound = errors.New("db session: not found")

// Create inserts a new session row.
func (r *SessionRepo) Create(ctx context.Context, rec SessionRecord) error {
	if rec.ID == "" || rec.UserID == "" || rec.TokenHash == "" {
		return fmt.Errorf("db session create: id, user_id and token_hash are required")
	}
	_, err := r.q.ExecContext(ctx, `
		INSERT INTO session (id, user_id, device_label, token_hash, expires_at, created_at, last_used_at)
		VALUES (?, ?, ?, ?, ?, datetime('now'), ?)
	`, rec.ID, rec.UserID, nullString(rec.DeviceLabel), rec.TokenHash,
		rec.ExpiresAt.UTC().Format("2006-01-02 15:04:05"), nullTime(rec.LastUsedAt))
	if err != nil {
		return fmt.Errorf("db session create: %w", err)
	}
	return nil
}

// GetByTokenHash returns a session by hashed refresh secret.
func (r *SessionRepo) GetByTokenHash(ctx context.Context, tokenHash string) (SessionRecord, error) {
	var rec SessionRecord
	var deviceLabel sql.NullString
	var revokedAt, lastUsed sql.NullString
	var expires, created string
	err := r.q.QueryRowContext(ctx, `
		SELECT id, user_id, device_label, token_hash, expires_at, revoked_at, created_at, last_used_at
		FROM session WHERE token_hash = ?
	`, tokenHash).Scan(
		&rec.ID, &rec.UserID, &deviceLabel, &rec.TokenHash,
		&expires, &revokedAt, &created, &lastUsed,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return SessionRecord{}, ErrSessionNotFound
	}
	if err != nil {
		return SessionRecord{}, fmt.Errorf("db session get by hash: %w", err)
	}
	if deviceLabel.Valid {
		rec.DeviceLabel = deviceLabel.String
	}
	rec.ExpiresAt = parseSQLiteTime(expires)
	rec.CreatedAt = parseSQLiteTime(created)
	if revokedAt.Valid {
		t := parseSQLiteTime(revokedAt.String)
		rec.RevokedAt = &t
	}
	if lastUsed.Valid {
		t := parseSQLiteTime(lastUsed.String)
		rec.LastUsedAt = &t
	}
	return rec, nil
}

// GetByID returns a session by id.
func (r *SessionRepo) GetByID(ctx context.Context, id string) (SessionRecord, error) {
	var rec SessionRecord
	var deviceLabel sql.NullString
	var revokedAt, lastUsed sql.NullString
	var expires, created string
	err := r.q.QueryRowContext(ctx, `
		SELECT id, user_id, device_label, token_hash, expires_at, revoked_at, created_at, last_used_at
		FROM session WHERE id = ?
	`, id).Scan(
		&rec.ID, &rec.UserID, &deviceLabel, &rec.TokenHash,
		&expires, &revokedAt, &created, &lastUsed,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return SessionRecord{}, ErrSessionNotFound
	}
	if err != nil {
		return SessionRecord{}, fmt.Errorf("db session get by id: %w", err)
	}
	if deviceLabel.Valid {
		rec.DeviceLabel = deviceLabel.String
	}
	rec.ExpiresAt = parseSQLiteTime(expires)
	rec.CreatedAt = parseSQLiteTime(created)
	if revokedAt.Valid {
		t := parseSQLiteTime(revokedAt.String)
		rec.RevokedAt = &t
	}
	if lastUsed.Valid {
		t := parseSQLiteTime(lastUsed.String)
		rec.LastUsedAt = &t
	}
	return rec, nil
}

// ListByUser returns active and revoked sessions for a user, newest first.
func (r *SessionRepo) ListByUser(ctx context.Context, userID string) ([]SessionRecord, error) {
	rows, err := r.q.QueryContext(ctx, `
		SELECT id, user_id, device_label, token_hash, expires_at, revoked_at, created_at, last_used_at
		FROM session WHERE user_id = ?
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("db session list: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []SessionRecord
	for rows.Next() {
		var rec SessionRecord
		var deviceLabel sql.NullString
		var revokedAt, lastUsed sql.NullString
		var expires, created string
		if err := rows.Scan(
			&rec.ID, &rec.UserID, &deviceLabel, &rec.TokenHash,
			&expires, &revokedAt, &created, &lastUsed,
		); err != nil {
			return nil, fmt.Errorf("db session list scan: %w", err)
		}
		if deviceLabel.Valid {
			rec.DeviceLabel = deviceLabel.String
		}
		rec.ExpiresAt = parseSQLiteTime(expires)
		rec.CreatedAt = parseSQLiteTime(created)
		if revokedAt.Valid {
			t := parseSQLiteTime(revokedAt.String)
			rec.RevokedAt = &t
		}
		if lastUsed.Valid {
			t := parseSQLiteTime(lastUsed.String)
			rec.LastUsedAt = &t
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}

// RevokeByID marks a session revoked by id.
func (r *SessionRepo) RevokeByID(ctx context.Context, id string) error {
	res, err := r.q.ExecContext(ctx, `
		UPDATE session SET revoked_at = datetime('now') WHERE id = ? AND revoked_at IS NULL
	`, id)
	if err != nil {
		return fmt.Errorf("db session revoke: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("db session revoke rows: %w", err)
	}
	if n == 0 {
		return ErrSessionNotFound
	}
	return nil
}

// RevokeSession revokes every token in a session group (same session id).
func (r *SessionRepo) RevokeSession(ctx context.Context, sessionID string) error {
	_, err := r.q.ExecContext(ctx, `
		UPDATE session SET revoked_at = datetime('now') WHERE id = ? AND revoked_at IS NULL
	`, sessionID)
	if err != nil {
		return fmt.Errorf("db session revoke session: %w", err)
	}
	return nil
}

// TouchLastUsed updates last_used_at for a session.
func (r *SessionRepo) TouchLastUsed(ctx context.Context, id string) error {
	_, err := r.q.ExecContext(ctx, `
		UPDATE session SET last_used_at = datetime('now') WHERE id = ?
	`, id)
	if err != nil {
		return fmt.Errorf("db session touch: %w", err)
	}
	return nil
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func nullTime(t *time.Time) sql.NullString {
	if t == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: t.UTC().Format("2006-01-02 15:04:05"), Valid: true}
}
