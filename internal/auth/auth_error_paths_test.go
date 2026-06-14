package auth_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/auth"
	"github.com/somralab/somra-media/internal/platform/db"
)

func TestFromContext_Empty(t *testing.T) {
	_, ok := auth.FromContext(context.Background())
	require.False(t, ok)
}

func TestRefresh_UserDeleted(t *testing.T) {
	svc, d := newTestService(t)
	ctx := context.Background()
	user, pair, err := svc.CreateAdmin(ctx, "admin", "AdminPass1")
	require.NoError(t, err)
	_, err = d.SQL().ExecContext(ctx, `DELETE FROM user_account WHERE id = ?`, user.ID)
	require.NoError(t, err)
	_, err = svc.Refresh(ctx, pair.RefreshToken)
	require.Error(t, err)
}

func TestRevokeSession_NotFound(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()
	user, _, err := svc.CreateAdmin(ctx, "admin", "AdminPass1")
	require.NoError(t, err)
	err = svc.RevokeSession(ctx, user.ID, "missing-session")
	require.ErrorIs(t, err, db.ErrSessionNotFound)
}

func TestSetupRequired_DBClosed(t *testing.T) {
	svc, d := newTestService(t)
	require.NoError(t, d.Close())
	_, err := svc.SetupRequired(context.Background())
	require.Error(t, err)
}

func TestResolveAuth_UnknownUserPermissions(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()
	claims := auth.Claims{Subject: auth.Subject{UserID: "does-not-exist"}}
	_, err := svc.ResolveAuth(ctx, claims)
	require.NoError(t, err)
}
