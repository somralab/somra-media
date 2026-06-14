package db

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotificationRepos_CRUD(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	users := NewUserRepo(d.Querier())
	userID := uuid.NewString()
	_, err := users.Create(ctx, userID, "notify-user", "hash", []string{"user"})
	require.NoError(t, err)

	chRepo := NewNotificationChannelRepo(d.Querier())
	prefRepo := NewNotificationPreferenceRepo(d.Querier())

	chID, err := chRepo.Create(ctx, NotificationChannel{
		ChannelType: NotificationChannelWebhook,
		Name:        "Ops webhook",
		Config:      `{"url":"https://example.com/hook"}`,
		Enabled:     true,
	})
	require.NoError(t, err)
	require.Greater(t, chID, int64(0))

	ch, err := chRepo.GetByID(ctx, chID)
	require.NoError(t, err)
	assert.Equal(t, NotificationChannelWebhook, ch.ChannelType)
	assert.True(t, ch.Enabled)
	assert.False(t, ch.CreatedAt.IsZero())

	channels, err := chRepo.List(ctx)
	require.NoError(t, err)
	require.Len(t, channels, 1)

	require.NoError(t, chRepo.SetEnabled(ctx, chID, false))
	ch, err = chRepo.GetByID(ctx, chID)
	require.NoError(t, err)
	assert.False(t, ch.Enabled)

	prefID, err := prefRepo.Upsert(ctx, NotificationPreference{
		UserID:          userID,
		EventType:       "request.created",
		ChannelID:       chID,
		Enabled:         true,
		DebounceSeconds: 30,
	})
	require.NoError(t, err)
	require.Greater(t, prefID, int64(0))

	prefs, err := prefRepo.ListByUser(ctx, userID)
	require.NoError(t, err)
	require.Len(t, prefs, 1)
	assert.Equal(t, "request.created", prefs[0].EventType)
	assert.Equal(t, 30, prefs[0].DebounceSeconds)

	prefID2, err := prefRepo.Upsert(ctx, NotificationPreference{
		UserID:          userID,
		EventType:       "request.created",
		ChannelID:       chID,
		Enabled:         false,
		DebounceSeconds: 60,
	})
	require.NoError(t, err)
	assert.Equal(t, prefID, prefID2)

	prefs, err = prefRepo.ListByUser(ctx, userID)
	require.NoError(t, err)
	require.Len(t, prefs, 1)
	assert.False(t, prefs[0].Enabled)
	assert.Equal(t, 60, prefs[0].DebounceSeconds)

	_, err = chRepo.GetByID(ctx, 99999)
	require.ErrorIs(t, err, ErrNotificationChannelNotFound)

	_, err = prefRepo.Upsert(ctx, NotificationPreference{
		ID:        99999,
		UserID:    userID,
		EventType: "request.created",
		ChannelID: chID,
	})
	require.ErrorIs(t, err, ErrNotificationPreferenceNotFound)
}

func TestNotificationPreferenceRepo_ListAfterClose(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	repo := NewNotificationPreferenceRepo(d.Querier())
	require.NoError(t, d.Close())
	_, err := repo.ListByUser(ctx, "user")
	require.Error(t, err)
}
