package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/somralab/somra-media/internal/platform/db"
)

// SQLiteRefreshStore implements RefreshTokenStore over session rows.
type SQLiteRefreshStore struct {
	sessions   *db.SessionRepo
	pepper     []byte
	refreshTTL time.Duration
}

// NewSQLiteRefreshStore returns a store backed by session repo.
func NewSQLiteRefreshStore(sessions *db.SessionRepo, pepper []byte, refreshTTL time.Duration) *SQLiteRefreshStore {
	return &SQLiteRefreshStore{sessions: sessions, pepper: pepper, refreshTTL: refreshTTL}
}

// Issue mints and stores a refresh token.
func (s *SQLiteRefreshStore) Issue(ctx context.Context, sub Subject, sessionID string, ttl time.Duration) (string, RefreshToken, error) {
	if ttl <= 0 {
		ttl = s.refreshTTL
	}
	secret, err := NewRefreshSecret()
	if err != nil {
		return "", RefreshToken{}, err
	}
	hash := HashRefreshSecret(s.pepper, secret)
	exp := time.Now().Add(ttl)
	rec := db.SessionRecord{
		ID:        sessionID,
		UserID:    sub.UserID,
		TokenHash: hash,
		ExpiresAt: exp,
	}
	if err := s.sessions.Create(ctx, rec); err != nil {
		return "", RefreshToken{}, fmt.Errorf("auth refresh issue: %w", err)
	}
	return secret, RefreshToken{
		ID:        sessionID,
		SessionID: sessionID,
		Subject:   sub,
		IssuedAt:  time.Now(),
		ExpiresAt: exp,
	}, nil
}

// Lookup resolves a refresh secret.
func (s *SQLiteRefreshStore) Lookup(ctx context.Context, secret string) (RefreshToken, error) {
	hash := HashRefreshSecret(s.pepper, secret)
	rec, err := s.sessions.GetByTokenHash(ctx, hash)
	if err != nil {
		if err == db.ErrSessionNotFound {
			return RefreshToken{}, ErrTokenNotFound
		}
		return RefreshToken{}, fmt.Errorf("auth refresh lookup: %w", err)
	}
	if rec.RevokedAt != nil {
		return RefreshToken{}, ErrRevokedToken
	}
	if time.Now().After(rec.ExpiresAt) {
		return RefreshToken{}, ErrTokenNotFound
	}
	return RefreshToken{
		ID:        rec.ID,
		SessionID: rec.ID,
		Subject:   Subject{UserID: rec.UserID},
		IssuedAt:  rec.CreatedAt,
		ExpiresAt: rec.ExpiresAt,
		RevokedAt: rec.RevokedAt,
	}, nil
}

// Revoke marks a token revoked by id.
func (s *SQLiteRefreshStore) Revoke(ctx context.Context, id string) error {
	if err := s.sessions.RevokeByID(ctx, id); err != nil {
		if err == db.ErrSessionNotFound {
			return nil
		}
		return fmt.Errorf("auth refresh revoke: %w", err)
	}
	return nil
}

// RevokeSession revokes every token in a session.
func (s *SQLiteRefreshStore) RevokeSession(ctx context.Context, sessionID string) error {
	if err := s.sessions.RevokeSession(ctx, sessionID); err != nil {
		return fmt.Errorf("auth refresh revoke session: %w", err)
	}
	return nil
}
