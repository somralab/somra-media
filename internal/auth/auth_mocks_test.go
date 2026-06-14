package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/auth"
	"github.com/somralab/somra-media/internal/platform/db"
)

type failTokenService struct{}

func (failTokenService) Validate(context.Context, string) (auth.Claims, error) {
	return auth.Claims{}, auth.ErrInvalidToken
}

func (failTokenService) Issue(context.Context, auth.Subject, string) (string, auth.Claims, error) {
	return "", auth.Claims{}, errors.New("issue failed")
}

func TestIssueTokens_TokenFailure(t *testing.T) {
	d := openTestDB(t)
	secret := []byte("test-secret-key-at-least-32-bytes!!")
	jwtCfg := auth.DefaultJWTConfig(secret)
	q := d.Querier()
	svc := auth.NewService(auth.ServiceConfig{
		Users: db.NewUserRepo(q), Sessions: db.NewSessionRepo(q), Profiles: db.NewProfileRepo(q),
		Tokens:  failTokenService{},
		Refresh: auth.NewSQLiteRefreshStore(db.NewSessionRepo(q), secret, jwtCfg.RefreshTTL),
		Hasher:  auth.NewPasswordHasher(auth.DefaultPasswordPolicy()),
		Lockout: nil, JWT: jwtCfg,
	})
	_, _, err := svc.CreateAdmin(context.Background(), "admin", "AdminPass1")
	require.Error(t, err)
}

type failRefreshStore struct {
	auth.RefreshTokenStore
}

func (failRefreshStore) Issue(context.Context, auth.Subject, string, time.Duration) (string, auth.RefreshToken, error) {
	return "", auth.RefreshToken{}, errors.New("refresh issue failed")
}

func (failRefreshStore) Lookup(context.Context, string) (auth.RefreshToken, error) {
	return auth.RefreshToken{}, auth.ErrTokenNotFound
}

func (failRefreshStore) Revoke(context.Context, string) error        { return nil }
func (failRefreshStore) RevokeSession(context.Context, string) error { return nil }

func TestIssueTokens_RefreshFailure(t *testing.T) {
	d := openTestDB(t)
	secret := []byte("test-secret-key-at-least-32-bytes!!")
	jwtCfg := auth.DefaultJWTConfig(secret)
	q := d.Querier()
	base := auth.NewSQLiteRefreshStore(db.NewSessionRepo(q), secret, jwtCfg.RefreshTTL)
	svc := auth.NewService(auth.ServiceConfig{
		Users: db.NewUserRepo(q), Sessions: db.NewSessionRepo(q), Profiles: db.NewProfileRepo(q),
		Tokens: auth.NewJWTService(jwtCfg), Refresh: failRefreshStore{RefreshTokenStore: base},
		Hasher:  auth.NewPasswordHasher(auth.DefaultPasswordPolicy()),
		Lockout: nil, JWT: jwtCfg,
	})
	_, _, err := svc.CreateAdmin(context.Background(), "admin", "AdminPass1")
	require.Error(t, err)
}

func TestLogout_PropagatesError(t *testing.T) {
	d := openTestDB(t)
	secret := []byte("test-secret-key-at-least-32-bytes!!")
	jwtCfg := auth.DefaultJWTConfig(secret)
	q := d.Querier()
	svc := auth.NewService(auth.ServiceConfig{
		Users: db.NewUserRepo(q), Sessions: db.NewSessionRepo(q), Profiles: db.NewProfileRepo(q),
		Tokens:  auth.NewJWTService(jwtCfg),
		Refresh: brokenLookupStore{},
		Hasher:  auth.NewPasswordHasher(auth.DefaultPasswordPolicy()),
		Lockout: nil, JWT: jwtCfg,
	})
	err := svc.Logout(context.Background(), "token")
	require.Error(t, err)
}

type brokenLookupStore struct{}

func (brokenLookupStore) Issue(context.Context, auth.Subject, string, time.Duration) (string, auth.RefreshToken, error) {
	return "", auth.RefreshToken{}, nil
}
func (brokenLookupStore) Lookup(context.Context, string) (auth.RefreshToken, error) {
	return auth.RefreshToken{}, errors.New("lookup boom")
}
func (brokenLookupStore) Revoke(context.Context, string) error        { return nil }
func (brokenLookupStore) RevokeSession(context.Context, string) error { return nil }

func TestJWTValidate_UnsupportedAlgorithm(t *testing.T) {
	secret := []byte("jwt-test-secret-key-32-chars!!")
	svc := auth.NewJWTService(auth.DefaultJWTConfig(secret))
	// alg none style token
	_, err := svc.Validate(context.Background(), "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1In0.sig")
	require.ErrorIs(t, err, auth.ErrInvalidToken)
}

func TestRatingAllowed_MaxEmpty(t *testing.T) {
	empty := ""
	require.True(t, auth.RatingAllowed(&empty, strPtr("R")))
}

type countingTokenService struct {
	auth.TokenService
	failAfter int
	calls     int
}

func (c *countingTokenService) Issue(ctx context.Context, sub auth.Subject, sid string) (string, auth.Claims, error) {
	c.calls++
	if c.calls > c.failAfter {
		return "", auth.Claims{}, errors.New("issue failed late")
	}
	return c.TokenService.Issue(ctx, sub, sid)
}

func TestRefresh_TokenIssueFailure(t *testing.T) {
	d := openTestDB(t)
	secret := []byte("test-secret-key-at-least-32-bytes!!")
	jwtCfg := auth.DefaultJWTConfig(secret)
	q := d.Querier()
	base := auth.NewJWTService(jwtCfg)
	counter := &countingTokenService{TokenService: base, failAfter: 1}
	svc := auth.NewService(auth.ServiceConfig{
		Users: db.NewUserRepo(q), Sessions: db.NewSessionRepo(q), Profiles: db.NewProfileRepo(q),
		Tokens:  counter,
		Refresh: auth.NewSQLiteRefreshStore(db.NewSessionRepo(q), secret, jwtCfg.RefreshTTL),
		Hasher:  auth.NewPasswordHasher(auth.DefaultPasswordPolicy()),
		Lockout: nil, JWT: jwtCfg,
	})
	ctx := context.Background()
	_, pair, err := svc.CreateAdmin(ctx, "admin", "AdminPass1")
	require.NoError(t, err)
	_, err = svc.Refresh(ctx, pair.RefreshToken)
	require.Error(t, err)
}

func TestRegister_CreateFails(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()
	_, _, err := svc.CreateAdmin(ctx, "admin", "AdminPass1")
	require.NoError(t, err)
	_, err = svc.Register(ctx, "admin", "AdminPass1", []string{"user"})
	require.Error(t, err)
}

type countingRefreshStore struct {
	inner *auth.SQLiteRefreshStore
	limit int
	n     int
}

func (c *countingRefreshStore) Issue(ctx context.Context, sub auth.Subject, sid string, ttl time.Duration) (string, auth.RefreshToken, error) {
	c.n++
	if c.n > c.limit {
		return "", auth.RefreshToken{}, errors.New("refresh capped")
	}
	return c.inner.Issue(ctx, sub, sid, ttl)
}
func (c *countingRefreshStore) Lookup(ctx context.Context, secret string) (auth.RefreshToken, error) {
	return c.inner.Lookup(ctx, secret)
}
func (c *countingRefreshStore) Revoke(ctx context.Context, id string) error {
	return c.inner.Revoke(ctx, id)
}
func (c *countingRefreshStore) RevokeSession(ctx context.Context, sid string) error {
	return c.inner.RevokeSession(ctx, sid)
}

func TestRefresh_RefreshIssueFailure(t *testing.T) {
	d := openTestDB(t)
	secret := []byte("test-secret-key-at-least-32-bytes!!")
	jwtCfg := auth.DefaultJWTConfig(secret)
	q := d.Querier()
	inner := auth.NewSQLiteRefreshStore(db.NewSessionRepo(q), secret, jwtCfg.RefreshTTL)
	cr := &countingRefreshStore{inner: inner, limit: 1}
	svc := auth.NewService(auth.ServiceConfig{
		Users: db.NewUserRepo(q), Sessions: db.NewSessionRepo(q), Profiles: db.NewProfileRepo(q),
		Tokens: auth.NewJWTService(jwtCfg), Refresh: cr,
		Hasher:  auth.NewPasswordHasher(auth.DefaultPasswordPolicy()),
		Lockout: nil, JWT: jwtCfg,
	})
	ctx := context.Background()
	_, pair, err := svc.CreateAdmin(ctx, "admin", "AdminPass1")
	require.NoError(t, err)
	_, err = svc.Refresh(ctx, pair.RefreshToken)
	require.Error(t, err)
}

func TestLogin_ContextCanceled(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()
	_, _, err := svc.CreateAdmin(ctx, "admin", "AdminPass1")
	require.NoError(t, err)
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	_, _, err = svc.Login(canceled, "admin", "AdminPass1", "", "1.1.1.1")
	require.Error(t, err)
}

func TestListSessions_DBError(t *testing.T) {
	svc, d := newTestService(t)
	ctx := context.Background()
	user, _, err := svc.CreateAdmin(ctx, "admin", "AdminPass1")
	require.NoError(t, err)
	require.NoError(t, d.Close())
	_, err = svc.ListSessions(ctx, user.ID)
	require.Error(t, err)
}

func TestCreateAdmin_SetupCheckError(t *testing.T) {
	svc, d := newTestService(t)
	require.NoError(t, d.Close())
	_, _, err := svc.CreateAdmin(context.Background(), "admin", "AdminPass1")
	require.Error(t, err)
}
