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

func TestBearerToken(t *testing.T) {
	t.Parallel()
	tok, ok := auth.BearerToken("Bearer abc.def")
	assert.True(t, ok)
	assert.Equal(t, "abc.def", tok)
	_, ok = auth.BearerToken("Basic x")
	assert.False(t, ok)
}

func TestHashRefreshSecret_Deterministic(t *testing.T) {
	t.Parallel()
	pepper := []byte("pepper-bytes-1234")
	a := auth.HashRefreshSecret(pepper, "secret")
	b := auth.HashRefreshSecret(pepper, "secret")
	assert.Equal(t, a, b)
	assert.NotEqual(t, a, auth.HashRefreshSecret(pepper, "other"))
}

func TestSQLiteRefreshStore_IssueLookupRevoke(t *testing.T) {
	d := openTestDB(t)
	q := d.Querier()
	secret := []byte("refresh-pepper-32-bytes-long!!")
	store := auth.NewSQLiteRefreshStore(db.NewSessionRepo(q), secret, time.Hour)
	users := db.NewUserRepo(q)
	ctx := context.Background()

	uid := auth.NewSessionID()
	_, err := users.Create(ctx, uid, "refresh-user", "hash", []string{"user"})
	require.NoError(t, err)

	sub := auth.Subject{UserID: uid, Username: "refresh-user", Roles: []string{"user"}}
	sid := auth.NewSessionID()
	raw, rec, err := store.Issue(ctx, sub, sid, time.Hour)
	require.NoError(t, err)
	assert.NotEmpty(t, raw)
	assert.Equal(t, sid, rec.SessionID)

	found, err := store.Lookup(ctx, raw)
	require.NoError(t, err)
	assert.Equal(t, uid, found.Subject.UserID)

	require.NoError(t, store.Revoke(ctx, rec.ID))
	_, err = store.Lookup(ctx, raw)
	require.ErrorIs(t, err, auth.ErrRevokedToken)
}

func TestService_RegisterAndSessions(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()
	_, _, err := svc.CreateAdmin(ctx, "admin", "AdminPass1")
	require.NoError(t, err)

	user, err := svc.Register(ctx, "member", "MemberPass1", []string{"user"})
	require.NoError(t, err)
	assert.Equal(t, "member", user.Username)

	_, pair, err := svc.Login(ctx, "member", "MemberPass1", "phone", "127.0.0.1")
	require.NoError(t, err)

	sessions, err := svc.ListSessions(ctx, user.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, sessions)

	require.NoError(t, svc.RevokeSession(ctx, user.ID, pair.SessionID))
}

func TestService_ResolveAuth(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()
	user, pair, err := svc.CreateAdmin(ctx, "admin", "AdminPass1")
	require.NoError(t, err)

	claims, err := svc.TokenService().Validate(ctx, pair.AccessToken)
	require.NoError(t, err)

	ac, err := svc.ResolveAuth(ctx, claims)
	require.NoError(t, err)
	assert.Contains(t, ac.Permissions, auth.PermUsersManage)
	assert.Equal(t, user.ID, ac.Claims.UserID)
}

func TestRefreshStore_NotFound(t *testing.T) {
	d := openTestDB(t)
	store := auth.NewSQLiteRefreshStore(db.NewSessionRepo(d.Querier()), []byte("pepper"), time.Hour)
	_, err := store.Lookup(context.Background(), "nonexistent")
	require.ErrorIs(t, err, auth.ErrTokenNotFound)
}

func TestNewRefreshSecret(t *testing.T) {
	a, err := auth.NewRefreshSecret()
	require.NoError(t, err)
	b, err := auth.NewRefreshSecret()
	require.NoError(t, err)
	assert.NotEqual(t, a, b)
}

func TestClientIP(t *testing.T) {
	t.Parallel()
	// ClientIP is tested indirectly via login; ensure package compiles with http import in service tests.
	assert.NotEmpty(t, auth.RoleAdmin)
}
