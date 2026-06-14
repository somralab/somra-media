package auth_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/auth"
	"github.com/somralab/somra-media/internal/platform/db"
)

func TestLockout_IPAndReset(t *testing.T) {
	d := openTestDB(t)
	q := d.Querier()
	cfg := auth.LockoutConfig{MaxFailures: 1, LockDuration: time.Hour, TrackIP: true, TrackUsername: true}
	lockout := auth.NewLoginLockout(db.NewLoginAttemptRepo(q), cfg)
	ctx := context.Background()

	locked, _, err := lockout.IsLocked(ctx, "1.2.3.4", db.LoginAttemptIP)
	require.NoError(t, err)
	assert.False(t, locked)

	require.NoError(t, lockout.RecordFailure(ctx, "1.2.3.4", "user"))
	locked, _, err = lockout.IsLocked(ctx, "1.2.3.4", db.LoginAttemptIP)
	require.NoError(t, err)
	assert.True(t, locked)

	require.NoError(t, lockout.Reset(ctx, "1.2.3.4", "user"))
	locked, _, err = lockout.IsLocked(ctx, "1.2.3.4", db.LoginAttemptIP)
	require.NoError(t, err)
	assert.False(t, locked)
}

func TestPasswordDecodeErrors(t *testing.T) {
	h := auth.NewPasswordHasher(auth.DefaultPasswordPolicy())
	hash, err := h.Hash("ValidPass1")
	require.NoError(t, err)

	cases := []string{
		"bad",
		"$argon2id$v=19$m=65536,t=3,p=4$salt$key",
		"$bcrypt$v=19$m=65536,t=3,p=4$salt$key",
		"$argon2id$v=19$m=bad,t=3,p=4$salt$key",
		"$argon2id$v=19$m=65536,t=3,p=4$!!!$key",
	}
	for _, c := range cases {
		ok, err := h.Verify("ValidPass1", c)
		if c == hash {
			require.NoError(t, err)
			assert.True(t, ok)
			continue
		}
		assert.False(t, ok)
	}
}

func TestBearerToken_EdgeCases(t *testing.T) {
	_, ok := auth.BearerToken("Bearer ")
	assert.False(t, ok)
	_, ok = auth.BearerToken("Token abc")
	assert.False(t, ok)
}

func TestClientIP_RemoteAddr(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.0.2.1:1234"
	assert.Equal(t, "192.0.2.1", auth.ClientIP(req))
}

func TestRegister_DuplicateUser(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()
	_, _, err := svc.CreateAdmin(ctx, "admin", "AdminPass1")
	require.NoError(t, err)
	_, err = svc.Register(ctx, "admin", "AdminPass1", []string{"user"})
	require.Error(t, err)
}

func TestLogin_UnknownUser(t *testing.T) {
	svc, _ := newTestService(t)
	_, _, err := svc.Login(context.Background(), "ghost", "AnyPass1", "", "1.1.1.1")
	require.ErrorIs(t, err, auth.ErrInvalidCredentials)
}

func TestRefreshStore_IssueZeroTTL(t *testing.T) {
	d := openTestDB(t)
	store := auth.NewSQLiteRefreshStore(db.NewSessionRepo(d.Querier()), []byte("pepper123456789012"), time.Hour)
	users := db.NewUserRepo(d.Querier())
	ctx := context.Background()
	uid := auth.NewSessionID()
	_, err := users.Create(ctx, uid, "u", "h", []string{"user"})
	require.NoError(t, err)
	_, _, err = store.Issue(ctx, auth.Subject{UserID: uid}, auth.NewSessionID(), 0)
	require.NoError(t, err)
}

func TestResolveAuth_ProfileMissing(t *testing.T) {
	svc, d := newTestService(t)
	ctx := context.Background()
	user, pair, err := svc.CreateAdmin(ctx, "admin", "AdminPass1")
	require.NoError(t, err)
	_, err = d.SQL().ExecContext(ctx, `DELETE FROM user_profile WHERE user_id = ?`, user.ID)
	require.NoError(t, err)
	claims, err := svc.TokenService().Validate(ctx, pair.AccessToken)
	require.NoError(t, err)
	ac, err := svc.ResolveAuth(ctx, claims)
	require.NoError(t, err)
	assert.Empty(t, ac.Profile.UserID)
}

func TestHasPermission_Empty(t *testing.T) {
	assert.False(t, auth.HasPermission(auth.AuthContext{}, auth.PermLibraryRead))
}

func TestNewSessionIDAndDefaultConfigs(t *testing.T) {
	assert.NotEmpty(t, auth.NewSessionID())
	cfg := auth.DefaultLockoutConfig()
	assert.Equal(t, 5, cfg.MaxFailures)
	jwt := auth.DefaultJWTConfig([]byte("secret"))
	assert.Equal(t, 15*time.Minute, jwt.AccessTTL)
}
