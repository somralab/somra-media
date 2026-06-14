package auth_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/auth"
	"github.com/somralab/somra-media/internal/platform/db"
)

func TestService_NilLockout(t *testing.T) {
	d := openTestDB(t)
	secret := []byte("test-secret-key-at-least-32-bytes!!")
	jwtCfg := auth.DefaultJWTConfig(secret)
	q := d.Querier()
	svc := auth.NewService(auth.ServiceConfig{
		Users: db.NewUserRepo(q), Sessions: db.NewSessionRepo(q), Profiles: db.NewProfileRepo(q),
		Tokens:  auth.NewJWTService(jwtCfg),
		Refresh: auth.NewSQLiteRefreshStore(db.NewSessionRepo(q), secret, jwtCfg.RefreshTTL),
		Hasher:  auth.NewPasswordHasher(auth.DefaultPasswordPolicy()), Lockout: nil,
		JWT: jwtCfg,
	})
	ctx := context.Background()
	_, pair, err := svc.CreateAdmin(ctx, "admin", "AdminPass1")
	require.NoError(t, err)
	_, pair2, err := svc.Login(ctx, "admin", "AdminPass1", "ipad", "10.0.0.1")
	require.NoError(t, err)
	require.NotEqual(t, pair.SessionID, pair2.SessionID)
}

func TestService_ResolveAuthPermissions(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()
	_, pair, err := svc.CreateAdmin(ctx, "admin", "AdminPass1")
	require.NoError(t, err)
	claims, err := svc.TokenService().Validate(ctx, pair.AccessToken)
	require.NoError(t, err)
	ac, err := svc.ResolveAuth(ctx, claims)
	require.NoError(t, err)
	require.NotEmpty(t, ac.Permissions)
	require.Equal(t, "en-US", ac.Profile.Locale)
}

func TestCreateAdmin_WeakPassword(t *testing.T) {
	svc, _ := newTestService(t)
	_, _, err := svc.CreateAdmin(context.Background(), "admin", "weak")
	require.ErrorIs(t, err, auth.ErrWeakPassword)
}

func TestLoginLockout_BothTracks(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()
	_, _, _ = svc.CreateAdmin(ctx, "admin", "AdminPass1")
	for i := 0; i < 5; i++ {
		_, _, _ = svc.Login(ctx, "admin", "nope", "dev", "8.8.8.8")
	}
	_, _, err := svc.Login(ctx, "admin", "AdminPass1", "dev", "8.8.4.4")
	require.ErrorIs(t, err, auth.ErrAccountLocked)
}

func TestListSessions_Multiple(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()
	user, _, err := svc.CreateAdmin(ctx, "admin", "AdminPass1")
	require.NoError(t, err)
	for i := 0; i < 3; i++ {
		_, _, err = svc.Login(ctx, "admin", "AdminPass1", "device", "1.2.3.4")
		require.NoError(t, err)
	}
	sessions, err := svc.ListSessions(ctx, user.ID)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(sessions), 3)
}

func TestRefreshStore_LookupDBError(t *testing.T) {
	d := openTestDB(t)
	store := auth.NewSQLiteRefreshStore(db.NewSessionRepo(d.Querier()), []byte("pepper123456789012"), 0)
	_, err := store.Lookup(context.Background(), "missing")
	require.ErrorIs(t, err, auth.ErrTokenNotFound)
}
