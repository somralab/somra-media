package requests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCanTransition_AllowedPaths(t *testing.T) {
	t.Parallel()
	assert.True(t, CanTransition(StatusPending, StatusApproved))
	assert.True(t, CanTransition(StatusPending, StatusRejected))
	assert.True(t, CanTransition(StatusPending, StatusCancelled))
	assert.True(t, CanTransition(StatusApproved, StatusCompleted))
	assert.True(t, CanTransition(StatusApproved, StatusCancelled))
}

func TestCanTransition_DisallowedPaths(t *testing.T) {
	t.Parallel()
	disallowed := [][2]Status{
		{StatusPending, StatusCompleted},
		{StatusPending, StatusPending},
		{StatusApproved, StatusPending},
		{StatusApproved, StatusRejected},
		{StatusApproved, StatusApproved},
		{StatusRejected, StatusApproved},
		{StatusRejected, StatusCancelled},
		{StatusCompleted, StatusCancelled},
		{StatusCompleted, StatusApproved},
		{StatusCancelled, StatusApproved},
		{StatusCancelled, StatusPending},
	}
	for _, pair := range disallowed {
		assert.False(t, CanTransition(pair[0], pair[1]), "%s -> %s", pair[0], pair[1])
	}
}

func TestIsTerminal(t *testing.T) {
	t.Parallel()
	assert.False(t, IsTerminal(StatusPending))
	assert.False(t, IsTerminal(StatusApproved))
	assert.True(t, IsTerminal(StatusRejected))
	assert.True(t, IsTerminal(StatusCompleted))
	assert.True(t, IsTerminal(StatusCancelled))
}

func TestTransitionTo_Success(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC)
	req := &Request{Status: StatusPending, CreatedAt: now, UpdatedAt: now}

	require.NoError(t, TransitionTo(req, StatusApproved, now))
	assert.Equal(t, StatusApproved, req.Status)
	assert.Equal(t, now, req.UpdatedAt)

	later := now.Add(time.Hour)
	require.NoError(t, TransitionTo(req, StatusCompleted, later))
	assert.Equal(t, StatusCompleted, req.Status)
	assert.Equal(t, later, req.UpdatedAt)
}

func TestTransitionTo_Invalid(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	req := &Request{Status: StatusRejected}
	err := TransitionTo(req, StatusApproved, now)
	require.ErrorIs(t, err, ErrInvalidTransition)

	err = TransitionTo(nil, StatusApproved, now)
	require.ErrorIs(t, err, ErrInvalidTransition)
}

func TestTransitionTo_RejectAndCancel(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()

	pending := &Request{Status: StatusPending}
	require.NoError(t, TransitionTo(pending, StatusRejected, now))
	assert.Equal(t, StatusRejected, pending.Status)

	approved := &Request{Status: StatusApproved}
	require.NoError(t, TransitionTo(approved, StatusCancelled, now))
	assert.Equal(t, StatusCancelled, approved.Status)
}

func TestAllowedTargets(t *testing.T) {
	t.Parallel()
	targets := AllowedTargets(StatusPending)
	assert.Len(t, targets, 3)
	assert.Empty(t, AllowedTargets(StatusCompleted))
}
