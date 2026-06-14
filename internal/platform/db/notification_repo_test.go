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
		ID:              prefID,
		UserID:          userID,
		EventType:       "request.created",
		ChannelID:       chID,
		Enabled:         false,
		DebounceSeconds: 45,
	})
	require.NoError(t, err)
	assert.Equal(t, prefID, prefID2)

	prefID2, err = prefRepo.Upsert(ctx, NotificationPreference{
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

func TestNotificationChannelRepo_ListEmpty(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })
	repo := NewNotificationChannelRepo(d.Querier())
	channels, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Empty(t, channels)
}

func TestNotificationPreferenceRepo_UpsertConflictLookup(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	users := NewUserRepo(d.Querier())
	userID := uuid.NewString()
	_, err := users.Create(ctx, userID, "conflict-user", "hash", []string{"user"})
	require.NoError(t, err)

	chRepo := NewNotificationChannelRepo(d.Querier())
	chID, err := chRepo.Create(ctx, NotificationChannel{
		ChannelType: NotificationChannelWebhook,
		Name:        "hook",
		Config:      `{"url":"https://example.com"}`,
		Enabled:     true,
	})
	require.NoError(t, err)

	prefRepo := NewNotificationPreferenceRepo(d.Querier())
	id1, err := prefRepo.Upsert(ctx, NotificationPreference{
		UserID: userID, EventType: "request.approved", ChannelID: chID, Enabled: true,
	})
	require.NoError(t, err)
	id2, err := prefRepo.Upsert(ctx, NotificationPreference{
		UserID: userID, EventType: "request.approved", ChannelID: chID, Enabled: false, DebounceSeconds: 5,
	})
	require.NoError(t, err)
	assert.Equal(t, id1, id2)
}

func TestNotificationChannelRepo_DefaultConfig(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })
	repo := NewNotificationChannelRepo(d.Querier())

	id, err := repo.Create(ctx, NotificationChannel{
		ChannelType: NotificationChannelWebhook,
		Name:        "Default cfg",
	})
	require.NoError(t, err)
	ch, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, "{}", ch.Config)
}

func TestNotificationPreferenceRepo_Validation(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })
	repo := NewNotificationPreferenceRepo(d.Querier())

	_, err := repo.ListByUser(ctx, "")
	require.Error(t, err)

	_, err = repo.Upsert(ctx, NotificationPreference{})
	require.Error(t, err)
}

func TestNotificationChannelRepo_Validation(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })
	repo := NewNotificationChannelRepo(d.Querier())

	_, err := repo.Create(ctx, NotificationChannel{})
	require.Error(t, err)

	_, err = repo.GetByID(ctx, 99999)
	require.ErrorIs(t, err, ErrNotificationChannelNotFound)

	require.Error(t, repo.SetEnabled(ctx, 99999, true))
}

func TestNotificationPreferenceRepo_ListAfterClose(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	repo := NewNotificationPreferenceRepo(d.Querier())
	require.NoError(t, d.Close())
	_, err := repo.ListByUser(ctx, "user")
	require.Error(t, err)
}
