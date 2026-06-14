package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/auth"
	"github.com/somralab/somra-media/internal/platform/db"
)

func TestLogin_DisabledUser(t *testing.T) {
	svc, d := newTestService(t)
	ctx := context.Background()
	user, err := svc.Register(ctx, "disabled", "ValidPass1", []string{"user"})
	require.NoError(t, err)
	repo := db.NewUserRepo(d.Querier())
	require.NoError(t, repo.SetDisabled(ctx, user.ID, true))

	_, _, err = svc.Login(ctx, "disabled", "ValidPass1", "", "1.2.3.4")
	require.ErrorIs(t, err, auth.ErrInvalidCredentials)
}

func TestRegister_WeakPassword(t *testing.T) {
	svc, _ := newTestService(t)
	_, _, err := svc.CreateAdmin(context.Background(), "admin", "AdminPass1")
	require.NoError(t, err)
	_, err = svc.Register(context.Background(), "u", "weak", []string{"user"})
	require.ErrorIs(t, err, auth.ErrWeakPassword)
}

func TestRefresh_DisabledUser(t *testing.T) {
	svc, d := newTestService(t)
	ctx := context.Background()
	user, pair, err := svc.CreateAdmin(ctx, "admin", "AdminPass1")
	require.NoError(t, err)
	repo := db.NewUserRepo(d.Querier())
	require.NoError(t, repo.SetDisabled(ctx, user.ID, true))
	_, err = svc.Refresh(ctx, pair.RefreshToken)
	require.ErrorIs(t, err, auth.ErrRevokedToken)
}

func TestLoginLockout_ResetOnSuccess(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()
	_, _, err := svc.CreateAdmin(ctx, "admin", "AdminPass1")
	require.NoError(t, err)
	for i := 0; i < 3; i++ {
		_, _, _ = svc.Login(ctx, "admin", "bad", "", "9.9.9.9")
	}
	_, _, err = svc.Login(ctx, "admin", "AdminPass1", "", "9.9.9.9")
	require.NoError(t, err)
}

func TestExpiredRefreshToken(t *testing.T) {
	d := openTestDB(t)
	secret := []byte("test-secret-key-at-least-32-bytes!!")
	jwtCfg := auth.DefaultJWTConfig(secret)
	jwtCfg.RefreshTTL = -time.Hour
	q := d.Querier()
	store := auth.NewSQLiteRefreshStore(db.NewSessionRepo(q), secret, jwtCfg.RefreshTTL)
	svc := auth.NewService(auth.ServiceConfig{
		Users: db.NewUserRepo(q), Sessions: db.NewSessionRepo(q), Profiles: db.NewProfileRepo(q),
		Tokens: auth.NewJWTService(jwtCfg), Refresh: store,
		Hasher:  auth.NewPasswordHasher(auth.DefaultPasswordPolicy()),
		Lockout: auth.NewLoginLockout(db.NewLoginAttemptRepo(q), auth.DefaultLockoutConfig()),
		JWT:     jwtCfg,
	})
	ctx := context.Background()
	_, pair, err := svc.CreateAdmin(ctx, "admin", "AdminPass1")
	require.NoError(t, err)
	_, err = svc.Refresh(ctx, pair.RefreshToken)
	require.ErrorIs(t, err, auth.ErrTokenNotFound)
}

func TestSetupRequired_ErrorPath(t *testing.T) {
	svc, _ := newTestService(t)
	required, err := svc.SetupRequired(context.Background())
	require.NoError(t, err)
	require.True(t, required)
}

func TestJWTIssue_WrongMethodRejected(t *testing.T) {
	secret := []byte("jwt-test-secret-key-32-chars!!")
	svc := auth.NewJWTService(auth.DefaultJWTConfig(secret))
	// tampered token
	_, err := svc.Validate(context.Background(), "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiJ1In0.")
	require.ErrorIs(t, err, auth.ErrInvalidToken)
}

func TestLoginLockout_UsernameOnly(t *testing.T) {
	d := openTestDB(t)
	secret := []byte("test-secret-key-at-least-32-bytes!!")
	q := d.Querier()
	lockout := auth.NewLoginLockout(db.NewLoginAttemptRepo(q), auth.LockoutConfig{
		MaxFailures: 2, LockDuration: time.Hour, TrackIP: false, TrackUsername: true,
	})
	svc := auth.NewService(auth.ServiceConfig{
		Users: db.NewUserRepo(q), Sessions: db.NewSessionRepo(q), Profiles: db.NewProfileRepo(q),
		Tokens:  auth.NewJWTService(auth.DefaultJWTConfig(secret)),
		Refresh: auth.NewSQLiteRefreshStore(db.NewSessionRepo(q), secret, time.Hour),
		Hasher:  auth.NewPasswordHasher(auth.DefaultPasswordPolicy()), Lockout: lockout,
		JWT: auth.DefaultJWTConfig(secret),
	})
	ctx := context.Background()
	_, _, _ = svc.CreateAdmin(ctx, "admin", "AdminPass1")
	_, _, _ = svc.Login(ctx, "admin", "bad", "", "1.1.1.1")
	_, _, err := svc.Login(ctx, "admin", "bad", "", "2.2.2.2")
	require.ErrorIs(t, err, auth.ErrInvalidCredentials)
	_, _, err = svc.Login(ctx, "admin", "bad", "", "3.3.3.3")
	require.ErrorIs(t, err, auth.ErrAccountLocked)
}
