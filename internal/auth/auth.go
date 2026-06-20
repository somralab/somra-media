// Package auth declares the identity contract used by every protected API
// handler. It contains interfaces and value types only; no JWT signing,
// password hashing, or refresh-token persistence lives here.
//
// Implementation lives in Sprint 03 (see .plan/sprint-03-auth-users/). By
// pinning the contract early we let downstream packets depend on stable
// types without forcing premature security decisions.
//
// Security policy reference: .plan/sprint-03-auth-users/04-security-tasks.md.
package auth

import (
	"context"
	"errors"
	"time"
)

// ErrInvalidToken signals an access token that failed validation. Higher
// layers map this to a 401 response.
var ErrInvalidToken = errors.New("auth: invalid token")

// ErrRevokedToken signals a refresh token that was found but had been
// previously revoked. Maps to a 401 response.
var ErrRevokedToken = errors.New("auth: refresh token revoked")

// ErrTokenNotFound signals an unknown refresh token (never issued, expired
// from store, or pruned). Maps to a 401 response.
var ErrTokenNotFound = errors.New("auth: refresh token not found")

// Subject identifies a user across access and refresh tokens.
type Subject struct {
	UserID   string
	Username string
	Roles    []string
}

// Claims are the subset of token payload fields the API gateway cares about.
// Implementations may carry additional private claims; only these are part of
// the cross-module contract.
type Claims struct {
	Subject
	// IssuedAt is when the token was minted.
	IssuedAt time.Time
	// ExpiresAt is the absolute expiry deadline.
	ExpiresAt time.Time
	// SessionID groups access and refresh tokens for revocation.
	SessionID string
}

// TokenService validates and mints short-lived access tokens. Refresh tokens
// are handled separately via RefreshTokenStore so the two concerns can evolve
// independently (signing key rotation vs. revocation list size).
type TokenService interface {
	// Validate parses and verifies an access token. Implementations must
	// return ErrInvalidToken (possibly wrapped) on any failure to keep the
	// API surface authorisation-leak free.
	Validate(ctx context.Context, raw string) (Claims, error)

	// Issue mints a new access token for sub and returns the encoded token
	// and the resolved claims (including ExpiresAt) so callers can set
	// cookie or response metadata consistently.
	Issue(ctx context.Context, sub Subject, sessionID string) (raw string, claims Claims, err error)
}

// RefreshToken describes a stored, opaque refresh token. The raw secret is
// not held in this struct after issuance; only its hash, expiry, and revocation
// metadata are tracked.
type RefreshToken struct {
	ID        string
	SessionID string
	Subject   Subject
	IssuedAt  time.Time
	ExpiresAt time.Time
	RevokedAt *time.Time
}

// RefreshTokenStore persists refresh-token metadata so individual tokens can
// be revoked server-side (logout, password change, suspected leak).
type RefreshTokenStore interface {
	// Issue mints and stores a new refresh token for sub. Returns the
	// opaque secret to hand back to the client plus the stored record.
	Issue(ctx context.Context, sub Subject, sessionID string, ttl time.Duration) (secret string, token RefreshToken, err error)

	// Lookup resolves a presented refresh secret to its stored record.
	// Returns ErrTokenNotFound if no match, ErrRevokedToken if revoked.
	Lookup(ctx context.Context, secret string) (RefreshToken, error)

	// Revoke marks the token identified by id as revoked. Idempotent.
	Revoke(ctx context.Context, id string) error

	// RevokeSession revokes every refresh token belonging to sessionID.
	// Used by logout-everywhere flows.
	RevokeSession(ctx context.Context, sessionID string) error
}
