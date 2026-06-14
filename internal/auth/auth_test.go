package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/auth"
	"github.com/somralab/somra-media/internal/platform/db"
)

func openTestDB(t *testing.T) *db.DB {
	t.Helper()
	cfg := db.Default()
	cfg.DataDir = t.TempDir()
	d, err := db.Initialize(context.Background(), cfg, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = d.Close() })
	return d
}

func newTestService(t *testing.T) (*auth.Service, *db.DB) {
	t.Helper()
	d := openTestDB(t)
	secret := []byte("test-secret-key-at-least-32-bytes!!")
	jwtCfg := auth.DefaultJWTConfig(secret)
	q := d.Querier()
	svc := auth.NewService(auth.ServiceConfig{
		Users:    db.NewUserRepo(q),
		Sessions: db.NewSessionRepo(q),
		Profiles: db.NewProfileRepo(q),
		Tokens:   auth.NewJWTService(jwtCfg),
		Refresh:  auth.NewSQLiteRefreshStore(db.NewSessionRepo(q), secret, jwtCfg.RefreshTTL),
		Hasher:   auth.NewPasswordHasher(auth.DefaultPasswordPolicy()),
		Lockout:  auth.NewLoginLockout(db.NewLoginAttemptRepo(q), auth.DefaultLockoutConfig()),
		JWT:      jwtCfg,
	})
	return svc, d
}

func TestPasswordHasher_HashAndVerify(t *testing.T) {
	t.Parallel()
	h := auth.NewPasswordHasher(auth.DefaultPasswordPolicy())
	hash, err := h.Hash("SecurePass1")
	require.NoError(t, err)
	ok, err := h.Verify("SecurePass1", hash)
	require.NoError(t, err)
	assert.True(t, ok)
	ok, err = h.Verify("wrong", hash)
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestPasswordHasher_RejectsWeakPassword(t *testing.T) {
	t.Parallel()
	h := auth.NewPasswordHasher(auth.DefaultPasswordPolicy())
	_, err := h.Hash("short")
	require.ErrorIs(t, err, auth.ErrWeakPassword)
}

func TestJWTService_IssueAndValidate(t *testing.T) {
	t.Parallel()
	secret := []byte("jwt-test-secret-key-32-chars!!")
	svc := auth.NewJWTService(auth.DefaultJWTConfig(secret))
	sub := auth.Subject{UserID: "u1", Username: "alice", Roles: []string{"admin"}}
	raw, claims, err := svc.Issue(context.Background(), sub, "sess-1")
	require.NoError(t, err)
	assert.NotEmpty(t, raw)
	assert.Equal(t, "u1", claims.UserID)
	parsed, err := svc.Validate(context.Background(), raw)
	require.NoError(t, err)
	assert.Equal(t, "alice", parsed.Username)
}

func TestJWTService_ExpiredToken(t *testing.T) {
	secret := []byte("jwt-test-secret-key-32-chars!!")
	cfg := auth.DefaultJWTConfig(secret)
	cfg.AccessTTL = -time.Minute
	svc := auth.NewJWTService(cfg)
	raw, _, err := svc.Issue(context.Background(), auth.Subject{UserID: "u1", Username: "a"}, "s1")
	require.NoError(t, err)
	_, err = svc.Validate(context.Background(), raw)
	require.ErrorIs(t, err, auth.ErrInvalidToken)
}

func TestService_SetupAdminAndLogin(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()

	required, err := svc.SetupRequired(ctx)
	require.NoError(t, err)
	assert.True(t, required)

	user, pair, err := svc.CreateAdmin(ctx, "admin", "AdminPass1")
	require.NoError(t, err)
	assert.Equal(t, "admin", user.Username)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)

	required, err = svc.SetupRequired(ctx)
	require.NoError(t, err)
	assert.False(t, required)

	_, _, err = svc.CreateAdmin(ctx, "other", "AdminPass1")
	require.ErrorIs(t, err, auth.ErrSetupComplete)

	user2, pair2, err := svc.Login(ctx, "admin", "AdminPass1", "browser", "127.0.0.1")
	require.NoError(t, err)
	assert.Equal(t, user.ID, user2.ID)
	assert.NotEmpty(t, pair2.AccessToken)

	_, _, err = svc.Login(ctx, "admin", "wrong", "browser", "127.0.0.1")
	require.ErrorIs(t, err, auth.ErrInvalidCredentials)
}

func TestService_RefreshAndLogout(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()
	_, pair, err := svc.CreateAdmin(ctx, "admin", "AdminPass1")
	require.NoError(t, err)

	_, newPair, err := svc.Refresh(ctx, pair.RefreshToken)
	require.NoError(t, err)
	assert.NotEmpty(t, newPair.AccessToken)

	_, _, err = svc.Refresh(ctx, pair.RefreshToken)
	require.Error(t, err)

	require.NoError(t, svc.Logout(ctx, newPair.RefreshToken))
}

func TestRatingAllowed(t *testing.T) {
	t.Parallel()
	pg := "PG"
	r := "R"
	assert.True(t, auth.RatingAllowed(&pg, strPtr("G")))
	assert.False(t, auth.RatingAllowed(&pg, strPtr("R")))
	assert.True(t, auth.RatingAllowed(nil, &r))
}

func TestLoginLockout(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()
	_, _, _ = svc.CreateAdmin(ctx, "admin", "AdminPass1")

	for i := 0; i < 5; i++ {
		_, _, err := svc.Login(ctx, "admin", "bad", "", "10.0.0.1")
		require.ErrorIs(t, err, auth.ErrInvalidCredentials)
	}
	_, _, err := svc.Login(ctx, "admin", "AdminPass1", "", "10.0.0.1")
	require.ErrorIs(t, err, auth.ErrAccountLocked)
}

func TestRBACPermissions(t *testing.T) {
	t.Parallel()
	ac := auth.AuthContext{Permissions: []string{auth.PermLibraryRead}}
	assert.True(t, auth.HasPermission(ac, auth.PermLibraryRead))
	assert.False(t, auth.HasPermission(ac, auth.PermLibraryWrite))
	ac2 := auth.AuthContext{Claims: auth.Claims{Subject: auth.Subject{Roles: []string{auth.RoleAdmin}}}}
	assert.True(t, auth.HasRole(ac2, auth.RoleAdmin))
}

func strPtr(s string) *string { return &s }
