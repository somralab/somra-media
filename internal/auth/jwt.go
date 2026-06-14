package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTConfig configures access-token signing.
type JWTConfig struct {
	Secret     []byte
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

// DefaultJWTConfig returns safe defaults (15m access, 7d refresh).
func DefaultJWTConfig(secret []byte) JWTConfig {
	return JWTConfig{
		Secret:     secret,
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 7 * 24 * time.Hour,
	}
}

// JWTService implements TokenService with HS256 JWTs.
type JWTService struct {
	cfg JWTConfig
}

// NewJWTService returns a TokenService backed by HS256.
func NewJWTService(cfg JWTConfig) *JWTService {
	return &JWTService{cfg: cfg}
}

type accessClaims struct {
	jwt.RegisteredClaims
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	Session  string   `json:"sid"`
}

// Validate parses and verifies an access token.
func (s *JWTService) Validate(_ context.Context, raw string) (Claims, error) {
	token, err := jwt.ParseWithClaims(raw, &accessClaims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("auth validate: unexpected signing method")
		}
		return s.cfg.Secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return Claims{}, fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}
	claims, ok := token.Claims.(*accessClaims)
	if !ok || !token.Valid {
		return Claims{}, ErrInvalidToken
	}
	sub := Subject{
		UserID:   claims.Subject,
		Username: claims.Username,
		Roles:    claims.Roles,
	}
	return Claims{
		Subject:   sub,
		IssuedAt:  claims.IssuedAt.Time,
		ExpiresAt: claims.ExpiresAt.Time,
		SessionID: claims.Session,
	}, nil
}

// Issue mints a new access token.
func (s *JWTService) Issue(_ context.Context, sub Subject, sessionID string) (string, Claims, error) {
	now := time.Now()
	exp := now.Add(s.cfg.AccessTTL)
	claims := accessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   sub.UserID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
			ID:        uuid.NewString(),
		},
		Username: sub.Username,
		Roles:    sub.Roles,
		Session:  sessionID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	raw, err := token.SignedString(s.cfg.Secret)
	if err != nil {
		return "", Claims{}, fmt.Errorf("auth issue: %w", err)
	}
	return raw, Claims{
		Subject:   sub,
		IssuedAt:  now,
		ExpiresAt: exp,
		SessionID: sessionID,
	}, nil
}

// NewSessionID returns a random session identifier.
func NewSessionID() string {
	return uuid.NewString()
}

// HashRefreshSecret returns a keyed HMAC-SHA256 hex digest of a refresh secret.
func HashRefreshSecret(pepper []byte, secret string) string {
	mac := hmac.New(sha256.New, pepper)
	_, _ = mac.Write([]byte(secret))
	return hex.EncodeToString(mac.Sum(nil))
}

// NewRefreshSecret returns a URL-safe opaque refresh token.
func NewRefreshSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("auth refresh secret: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// BearerToken extracts the token from an Authorization header value.
func BearerToken(header string) (string, bool) {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}
	tok := strings.TrimSpace(header[len(prefix):])
	if tok == "" {
		return "", false
	}
	return tok, true
}
