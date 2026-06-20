package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/somralab/somra-media/internal/platform/db"
)

// Permission names used by RBAC middleware.
const (
	PermLibraryRead         = "library:read"
	PermLibraryWrite        = "library:write"
	PermUsersManage         = "users:manage"
	PermProfileEdit         = "profile:edit"
	PermRequestsCreate      = "requests:create"
	PermRequestsRead        = "requests:read"
	PermRequestsManage      = "requests:manage"
	PermNotificationsManage = "notifications:manage"
	PermPluginsManage       = "plugins:manage"
)

// Role names.
const (
	RoleAdmin = "admin"
	RoleUser  = "user"
	RoleChild = "child"
)

// ContentRating ordering for parental filtering (lower = more restrictive).
var contentRatingOrder = map[string]int{
	"G":     0,
	"PG":    1,
	"PG-13": 2,
	"R":     3,
	"NC-17": 4,
}

// RatingAllowed returns true when itemRating is visible under maxRating.
// nil maxRating means no restriction; nil itemRating is treated as unrated (visible).
func RatingAllowed(maxRating *string, itemRating *string) bool {
	if maxRating == nil || *maxRating == "" {
		return true
	}
	if itemRating == nil || *itemRating == "" {
		return true
	}
	maxOrd, okMax := contentRatingOrder[*maxRating]
	itemOrd, okItem := contentRatingOrder[*itemRating]
	if !okMax || !okItem {
		return true
	}
	return itemOrd <= maxOrd
}

// AuthContext carries authenticated user data on request context.
type AuthContext struct {
	Claims      Claims
	Permissions []string
	Profile     db.UserProfile
}

type authCtxKey struct{}

// WithAuthContext attaches auth data to ctx.
func WithAuthContext(ctx context.Context, ac AuthContext) context.Context {
	return context.WithValue(ctx, authCtxKey{}, ac)
}

// FromContext returns auth data when present.
func FromContext(ctx context.Context) (AuthContext, bool) {
	if ctx == nil {
		return AuthContext{}, false
	}
	v, ok := ctx.Value(authCtxKey{}).(AuthContext)
	return v, ok
}

// HasPermission reports whether ac includes perm.
func HasPermission(ac AuthContext, perm string) bool {
	for _, p := range ac.Permissions {
		if p == perm {
			return true
		}
	}
	return false
}

// HasRole reports whether the subject carries role.
func HasRole(ac AuthContext, role string) bool {
	for _, r := range ac.Claims.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// LockoutConfig configures brute-force protection.
type LockoutConfig struct {
	MaxFailures   int
	LockDuration  time.Duration
	TrackIP       bool
	TrackUsername bool
}

// DefaultLockoutConfig returns Somra defaults.
func DefaultLockoutConfig() LockoutConfig {
	return LockoutConfig{
		MaxFailures:   5,
		LockDuration:  15 * time.Minute,
		TrackIP:       true,
		TrackUsername: true,
	}
}

// LoginLockout wraps login attempt persistence.
type LoginLockout struct {
	repo *db.LoginAttemptRepo
	cfg  LockoutConfig
}

// NewLoginLockout returns a lockout helper.
func NewLoginLockout(repo *db.LoginAttemptRepo, cfg LockoutConfig) *LoginLockout {
	return &LoginLockout{repo: repo, cfg: cfg}
}

// IsLocked reports whether identifier is currently locked out.
func (l *LoginLockout) IsLocked(ctx context.Context, identifier string, kind db.LoginAttemptKind) (bool, time.Time, error) {
	la, err := l.repo.Get(ctx, identifier, kind)
	if err != nil {
		return false, time.Time{}, err
	}
	if la.LockedUntil != nil && time.Now().Before(*la.LockedUntil) {
		return true, *la.LockedUntil, nil
	}
	return false, time.Time{}, nil
}

// RecordFailure increments counters for IP and/or username.
func (l *LoginLockout) RecordFailure(ctx context.Context, ip, username string) error {
	if l.cfg.TrackIP && ip != "" {
		if _, err := l.repo.RecordFailure(ctx, ip, db.LoginAttemptIP, l.cfg.MaxFailures, l.cfg.LockDuration); err != nil {
			return err
		}
	}
	if l.cfg.TrackUsername && username != "" {
		if _, err := l.repo.RecordFailure(ctx, strings.ToLower(username), db.LoginAttemptUsername, l.cfg.MaxFailures, l.cfg.LockDuration); err != nil {
			return err
		}
	}
	return nil
}

// Reset clears counters after successful login.
func (l *LoginLockout) Reset(ctx context.Context, ip, username string) error {
	if l.cfg.TrackIP && ip != "" {
		if err := l.repo.Reset(ctx, ip, db.LoginAttemptIP); err != nil {
			return err
		}
	}
	if l.cfg.TrackUsername && username != "" {
		if err := l.repo.Reset(ctx, strings.ToLower(username), db.LoginAttemptUsername); err != nil {
			return err
		}
	}
	return nil
}

// ClientIP extracts a best-effort client IP from the request.
func ClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return strings.TrimSpace(xrip)
	}
	host := r.RemoteAddr
	if i := strings.LastIndex(host, ":"); i >= 0 {
		return host[:i]
	}
	return host
}

// Service orchestrates login, refresh, logout, and setup flows.
type Service struct {
	users    *db.UserRepo
	sessions *db.SessionRepo
	profiles *db.ProfileRepo
	tokens   TokenService
	refresh  RefreshTokenStore
	hasher   *PasswordHasher
	lockout  *LoginLockout
	jwtCfg   JWTConfig
}

// ServiceConfig wires auth dependencies.
type ServiceConfig struct {
	Users    *db.UserRepo
	Sessions *db.SessionRepo
	Profiles *db.ProfileRepo
	Tokens   TokenService
	Refresh  RefreshTokenStore
	Hasher   *PasswordHasher
	Lockout  *LoginLockout
	JWT      JWTConfig
}

// NewService returns an auth service.
func NewService(cfg ServiceConfig) *Service {
	return &Service{
		users:    cfg.Users,
		sessions: cfg.Sessions,
		profiles: cfg.Profiles,
		tokens:   cfg.Tokens,
		refresh:  cfg.Refresh,
		hasher:   cfg.Hasher,
		lockout:  cfg.Lockout,
		jwtCfg:   cfg.JWT,
	}
}

// TokenPair holds access and refresh credentials.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	SessionID    string
}

// SetupRequired reports whether first-admin setup is needed.
func (s *Service) SetupRequired(ctx context.Context) (bool, error) {
	n, err := s.users.Count(ctx)
	if err != nil {
		return false, err
	}
	return n == 0, nil
}

// CreateAdmin creates the first admin user during setup.
func (s *Service) CreateAdmin(ctx context.Context, username, password string) (UserAccount, TokenPair, error) {
	required, err := s.SetupRequired(ctx)
	if err != nil {
		return UserAccount{}, TokenPair{}, err
	}
	if !required {
		return UserAccount{}, TokenPair{}, ErrSetupComplete
	}
	return s.register(ctx, username, password, []string{RoleAdmin}, "")
}

// Register creates a new user (admin-only in handlers).
func (s *Service) Register(ctx context.Context, username, password string, roles []string) (UserAccount, error) {
	hash, err := s.hasher.Hash(password)
	if err != nil {
		return UserAccount{}, err
	}
	id := uuid.NewString()
	if _, err := s.users.Create(ctx, id, username, hash, roles); err != nil {
		return UserAccount{}, err
	}
	user, err := s.users.GetByID(ctx, id)
	if err != nil {
		return UserAccount{}, err
	}
	return toUserAccount(user), nil
}

// HashPassword validates and hashes a password for admin updates.
func (s *Service) HashPassword(password string) (string, error) {
	return s.hasher.Hash(password)
}

var ErrSetupComplete = errors.New("auth: setup already completed")

// UserAccount is the public user view without password hash.
type UserAccount struct {
	ID       string
	Username string
	Roles    []string
	Disabled bool
}

func (s *Service) register(ctx context.Context, username, password string, roles []string, deviceLabel string) (UserAccount, TokenPair, error) {
	hash, err := s.hasher.Hash(password)
	if err != nil {
		return UserAccount{}, TokenPair{}, err
	}
	id := uuid.NewString()
	if _, err := s.users.Create(ctx, id, username, hash, roles); err != nil {
		return UserAccount{}, TokenPair{}, err
	}
	user, err := s.users.GetByID(ctx, id)
	if err != nil {
		return UserAccount{}, TokenPair{}, err
	}
	pair, err := s.issueTokens(ctx, user, deviceLabel)
	if err != nil {
		return UserAccount{}, TokenPair{}, err
	}
	return toUserAccount(user), pair, nil
}

// Login validates credentials and issues tokens.
func (s *Service) Login(ctx context.Context, username, password, deviceLabel, ip string) (UserAccount, TokenPair, error) {
	if s.lockout != nil {
		if locked, _, err := s.lockout.IsLocked(ctx, ip, db.LoginAttemptIP); err != nil {
			return UserAccount{}, TokenPair{}, err
		} else if locked {
			return UserAccount{}, TokenPair{}, ErrAccountLocked
		}
		if locked, _, err := s.lockout.IsLocked(ctx, strings.ToLower(username), db.LoginAttemptUsername); err != nil {
			return UserAccount{}, TokenPair{}, err
		} else if locked {
			return UserAccount{}, TokenPair{}, ErrAccountLocked
		}
	}

	user, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, db.ErrUserNotFound) {
			s.recordLoginFailure(ctx, ip, username)
			return UserAccount{}, TokenPair{}, ErrInvalidCredentials
		}
		return UserAccount{}, TokenPair{}, err
	}
	if user.Disabled {
		return UserAccount{}, TokenPair{}, ErrInvalidCredentials
	}
	ok, err := s.hasher.Verify(password, user.PasswordHash)
	if err != nil || !ok {
		s.recordLoginFailure(ctx, ip, username)
		return UserAccount{}, TokenPair{}, ErrInvalidCredentials
	}
	if s.lockout != nil {
		_ = s.lockout.Reset(ctx, ip, username)
	}
	pair, err := s.issueTokens(ctx, user, deviceLabel)
	if err != nil {
		return UserAccount{}, TokenPair{}, err
	}
	return toUserAccount(user), pair, nil
}

var ErrAccountLocked = errors.New("auth: account locked")

func (s *Service) recordLoginFailure(ctx context.Context, ip, username string) {
	if s.lockout != nil {
		_ = s.lockout.RecordFailure(ctx, ip, username)
	}
}

func (s *Service) issueTokens(ctx context.Context, user db.UserAccount, deviceLabel string) (TokenPair, error) {
	sessionID := NewSessionID()
	sub := Subject{UserID: user.ID, Username: user.Username, Roles: user.Roles}
	access, claims, err := s.tokens.Issue(ctx, sub, sessionID)
	if err != nil {
		return TokenPair{}, err
	}
	refreshSecret, err := s.refreshIssue(ctx, sub, sessionID, deviceLabel)
	if err != nil {
		return TokenPair{}, err
	}
	return TokenPair{
		AccessToken:  access,
		RefreshToken: refreshSecret,
		ExpiresAt:    claims.ExpiresAt,
		SessionID:    sessionID,
	}, nil
}

func (s *Service) refreshIssue(ctx context.Context, sub Subject, sessionID string, _ string) (string, error) {
	secret, _, err := s.refresh.Issue(ctx, sub, sessionID, s.jwtCfg.RefreshTTL)
	return secret, err
}

// Refresh rotates tokens using a refresh secret.
func (s *Service) Refresh(ctx context.Context, refreshSecret string) (UserAccount, TokenPair, error) {
	rec, err := s.refresh.Lookup(ctx, refreshSecret)
	if err != nil {
		return UserAccount{}, TokenPair{}, err
	}
	user, err := s.users.GetByID(ctx, rec.Subject.UserID)
	if err != nil {
		return UserAccount{}, TokenPair{}, err
	}
	if user.Disabled {
		return UserAccount{}, TokenPair{}, ErrRevokedToken
	}
	_ = s.refresh.Revoke(ctx, rec.ID)
	sub := Subject{UserID: user.ID, Username: user.Username, Roles: user.Roles}
	sessionID := NewSessionID()
	access, claims, err := s.tokens.Issue(ctx, sub, sessionID)
	if err != nil {
		return UserAccount{}, TokenPair{}, err
	}
	newRefresh, err := s.refreshIssue(ctx, sub, sessionID, "")
	if err != nil {
		return UserAccount{}, TokenPair{}, err
	}
	return toUserAccount(user), TokenPair{
		AccessToken:  access,
		RefreshToken: newRefresh,
		ExpiresAt:    claims.ExpiresAt,
		SessionID:    sessionID,
	}, nil
}

// Logout revokes the refresh token session.
func (s *Service) Logout(ctx context.Context, refreshSecret string) error {
	rec, err := s.refresh.Lookup(ctx, refreshSecret)
	if err != nil {
		if errors.Is(err, ErrTokenNotFound) || errors.Is(err, ErrRevokedToken) {
			return nil
		}
		return err
	}
	return s.refresh.RevokeSession(ctx, rec.SessionID)
}

// RevokeSession revokes a session by id for the owning user.
func (s *Service) RevokeSession(ctx context.Context, userID, sessionID string) error {
	rec, err := s.sessions.GetByID(ctx, sessionID)
	if err != nil {
		return err
	}
	if rec.UserID != userID {
		return ErrForbidden
	}
	return s.refresh.RevokeSession(ctx, sessionID)
}

var ErrForbidden = errors.New("auth: forbidden")

// ListSessions returns sessions for a user (without token hashes).
func (s *Service) ListSessions(ctx context.Context, userID string) ([]SessionInfo, error) {
	recs, err := s.sessions.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]SessionInfo, 0, len(recs))
	for _, rec := range recs {
		out = append(out, SessionInfo{
			ID:          rec.ID,
			DeviceLabel: rec.DeviceLabel,
			CreatedAt:   rec.CreatedAt,
			LastUsedAt:  rec.LastUsedAt,
			RevokedAt:   rec.RevokedAt,
			ExpiresAt:   rec.ExpiresAt,
		})
	}
	return out, nil
}

// SessionInfo is the public session view.
type SessionInfo struct {
	ID          string
	DeviceLabel string
	CreatedAt   time.Time
	LastUsedAt  *time.Time
	RevokedAt   *time.Time
	ExpiresAt   time.Time
}

// ResolveAuth loads permissions and profile for validated claims.
func (s *Service) ResolveAuth(ctx context.Context, claims Claims) (AuthContext, error) {
	perms, err := s.users.PermissionsForUser(ctx, claims.UserID)
	if err != nil {
		return AuthContext{}, fmt.Errorf("auth resolve permissions: %w", err)
	}
	profile, err := s.profiles.Get(ctx, claims.UserID)
	if err != nil && !errors.Is(err, db.ErrProfileNotFound) {
		return AuthContext{}, fmt.Errorf("auth resolve profile: %w", err)
	}
	return AuthContext{Claims: claims, Permissions: perms, Profile: profile}, nil
}

func toUserAccount(u db.UserAccount) UserAccount {
	return UserAccount{ID: u.ID, Username: u.Username, Roles: u.Roles, Disabled: u.Disabled}
}

// TokenService returns the configured access-token service.
func (s *Service) TokenService() TokenService {
	return s.tokens
}
