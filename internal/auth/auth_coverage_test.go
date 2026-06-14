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

func TestAuthContextHelpers(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	_, ok := auth.FromContext(ctx)
	assert.False(t, ok)

	ac := auth.AuthContext{Claims: auth.Claims{Subject: auth.Subject{UserID: "u1"}}}
	ctx = auth.WithAuthContext(ctx, ac)
	got, ok := auth.FromContext(ctx)
	require.True(t, ok)
	assert.Equal(t, "u1", got.Claims.UserID)
}

func TestPasswordValidate_AllRules(t *testing.T) {
	h := auth.NewPasswordHasher(auth.PasswordPolicy{
		MinLength: 8, RequireUpper: true, RequireLower: true, RequireDigit: true, RequireSpecial: true,
	})
	require.Error(t, h.Validate("NoSpecial1"))
	require.Error(t, h.Validate("alllower1!"))
	require.NoError(t, h.Validate("GoodPass1!"))
}

func TestPasswordVerify_MalformedHash(t *testing.T) {
	h := auth.NewPasswordHasher(auth.DefaultPasswordPolicy())
	ok, err := h.Verify("x", "not-a-hash")
	require.Error(t, err)
	assert.False(t, ok)
}

func TestJWTValidate_BadToken(t *testing.T) {
	svc := auth.NewJWTService(auth.DefaultJWTConfig([]byte("jwt-test-secret-key-32-chars!!")))
	_, err := svc.Validate(context.Background(), "not.a.jwt")
	require.ErrorIs(t, err, auth.ErrInvalidToken)
}

func TestClientIP_HeaderParsing(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.1, 198.51.100.2")
	assert.Equal(t, "203.0.113.1", auth.ClientIP(req))

	req.Header.Set("X-Real-IP", "10.0.0.5")
	req.Header.Del("X-Forwarded-For")
	assert.Equal(t, "10.0.0.5", auth.ClientIP(req))
}

func TestService_HashPassword(t *testing.T) {
	svc, _ := newTestService(t)
	hash, err := svc.HashPassword("ValidPass1")
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
}

func TestService_RevokeSessionForbidden(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()
	admin, pair, err := svc.CreateAdmin(ctx, "admin", "AdminPass1")
	require.NoError(t, err)

	err = svc.RevokeSession(ctx, "other-user", pair.SessionID)
	require.ErrorIs(t, err, auth.ErrForbidden)

	_, userPair, err := svc.Login(ctx, "admin", "AdminPass1", "", "127.0.0.1")
	require.NoError(t, err)
	require.NoError(t, svc.RevokeSession(ctx, admin.ID, userPair.SessionID))
}

func TestService_LogoutIdempotent(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()
	require.NoError(t, svc.Logout(ctx, "invalid"))
}

func TestRatingAllowed_UnknownRatings(t *testing.T) {
	t.Parallel()
	x := "X"
	assert.True(t, auth.RatingAllowed(&x, strPtr("Y")))
}

func TestHasRole_Missing(t *testing.T) {
	t.Parallel()
	ac := auth.AuthContext{Claims: auth.Claims{Subject: auth.Subject{Roles: []string{"user"}}}}
	assert.False(t, auth.HasRole(ac, auth.RoleAdmin))
}

func TestRefreshStore_RevokeSession(t *testing.T) {
	d := openTestDB(t)
	store := auth.NewSQLiteRefreshStore(db.NewSessionRepo(d.Querier()), []byte("pepper1234567890"), time.Hour)
	ctx := context.Background()
	sub := auth.Subject{UserID: "u", Username: "u", Roles: []string{"user"}}
	users := db.NewUserRepo(d.Querier())
	_, _ = users.Create(ctx, "u", "u", "h", []string{"user"})
	sid := auth.NewSessionID()
	raw, _, err := store.Issue(ctx, sub, sid, time.Hour)
	require.NoError(t, err)
	require.NoError(t, store.RevokeSession(ctx, sid))
	_, err = store.Lookup(ctx, raw)
	require.ErrorIs(t, err, auth.ErrRevokedToken)
}

func TestRefreshStore_RevokeUnknown(t *testing.T) {
	d := openTestDB(t)
	store := auth.NewSQLiteRefreshStore(db.NewSessionRepo(d.Querier()), []byte("pepper1234567890"), time.Hour)
	require.NoError(t, store.Revoke(context.Background(), "missing"))
}
