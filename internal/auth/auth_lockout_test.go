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

func TestLoginLockout_RecordFailureAndReset(t *testing.T) {
	d := openTestDB(t)
	repo := db.NewLoginAttemptRepo(d.Querier())
	lockout := auth.NewLoginLockout(repo, auth.LockoutConfig{
		MaxFailures:   2,
		LockDuration:  time.Minute,
		TrackIP:       true,
		TrackUsername: true,
	})
	ctx := context.Background()

	require.NoError(t, lockout.RecordFailure(ctx, "10.0.0.1", "Alice"))
	locked, _, err := lockout.IsLocked(ctx, "10.0.0.1", db.LoginAttemptIP)
	require.NoError(t, err)
	assert.False(t, locked)

	require.NoError(t, lockout.RecordFailure(ctx, "10.0.0.1", "Alice"))
	locked, _, err = lockout.IsLocked(ctx, "10.0.0.1", db.LoginAttemptIP)
	require.NoError(t, err)
	assert.True(t, locked)

	require.NoError(t, lockout.Reset(ctx, "10.0.0.1", "Alice"))
	locked, _, err = lockout.IsLocked(ctx, "10.0.0.1", db.LoginAttemptIP)
	require.NoError(t, err)
	assert.False(t, locked)
}

func TestService_LoginDisabledUser(t *testing.T) {
	svc, d := newTestService(t)
	ctx := context.Background()
	user, err := svc.Register(ctx, "disabled", "ValidPass1", []string{auth.RoleUser})
	require.NoError(t, err)

	require.NoError(t, db.NewUserRepo(d.Querier()).SetDisabled(ctx, user.ID, true))

	_, _, err = svc.Login(ctx, "disabled", "ValidPass1", "", "127.0.0.1")
	require.ErrorIs(t, err, auth.ErrInvalidCredentials)
}

func TestFromContext_Nil(t *testing.T) {
	t.Parallel()
	var ctx context.Context
	_, ok := auth.FromContext(ctx)
	assert.False(t, ok)
}
