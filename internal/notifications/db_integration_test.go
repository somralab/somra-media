package notifications_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/auth"
	"github.com/somralab/somra-media/internal/notifications"
	"github.com/somralab/somra-media/internal/platform/db"
)

func openNotificationsTestDB(t *testing.T) *db.DB {
	t.Helper()
	cfg := db.Default()
	cfg.DataDir = t.TempDir()
	ctx := context.Background()
	d, err := db.Initialize(ctx, cfg, nil)
	require.NoError(t, err)
	return d
}

func TestChannelFromDB_AllTypes(t *testing.T) {
	t.Parallel()

	webhook, err := notifications.ChannelFromDB(db.NotificationChannel{
		ChannelType: db.NotificationChannelWebhook,
		Config:      `{"url":"https://example.com/hook"}`,
	})
	require.NoError(t, err)
	assert.Equal(t, notifications.ChannelWebhook, webhook.ID())

	discord, err := notifications.ChannelFromDB(db.NotificationChannel{
		ChannelType: db.NotificationChannelDiscord,
		Config:      `{"webhookUrl":"https://discord.com/api/webhooks/x"}`,
	})
	require.NoError(t, err)
	assert.Equal(t, notifications.ChannelDiscord, discord.ID())

	email, err := notifications.ChannelFromDB(db.NotificationChannel{
		ChannelType: db.NotificationChannelEmail,
		Config:      `{"host":"smtp.example.com","from":"somra@example.com"}`,
	})
	require.NoError(t, err)
	assert.Equal(t, notifications.ChannelSMTP, email.ID())

	_, err = notifications.ChannelFromDB(db.NotificationChannel{
		ChannelType: "unknown",
		Config:      `{}`,
	})
	require.Error(t, err)

	_, err = notifications.ChannelFromDB(db.NotificationChannel{
		ChannelType: db.NotificationChannelWebhook,
		Config:      `{`,
	})
	require.Error(t, err)
}

func TestChannelFromDB_DecodeErrors(t *testing.T) {
	t.Parallel()
	_, err := notifications.ChannelFromDB(db.NotificationChannel{
		ChannelType: db.NotificationChannelDiscord,
		Config:      `{`,
	})
	require.Error(t, err)
	_, err = notifications.ChannelFromDB(db.NotificationChannel{
		ChannelType: db.NotificationChannelEmail,
		Config:      `{`,
	})
	require.Error(t, err)
}

func TestDBPreferenceStore_NilStore(t *testing.T) {
	t.Parallel()
	var store *notifications.DBPreferenceStore
	_, err := store.GetUserPreferences(context.Background(), "u")
	require.Error(t, err)
}

func TestDBPreferenceStore_LoadsUserState(t *testing.T) {
	ctx := context.Background()
	d := openNotificationsTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	users := db.NewUserRepo(d.Querier())
	profiles := db.NewProfileRepo(d.Querier())
	userID := uuid.NewString()
	_, err := users.Create(ctx, userID, "notify-user", "hash", []string{auth.RoleUser})
	require.NoError(t, err)
	require.NoError(t, profiles.Update(ctx, db.UserProfile{UserID: userID, Locale: "tr-TR", Theme: "cinematic"}))

	chRepo := db.NewNotificationChannelRepo(d.Querier())
	prefRepo := db.NewNotificationPreferenceRepo(d.Querier())
	chID, err := chRepo.Create(ctx, db.NotificationChannel{
		ChannelType: db.NotificationChannelWebhook,
		Name:        "hook",
		Config:      `{"url":"https://example.com/hook"}`,
		Enabled:     true,
	})
	require.NoError(t, err)

	_, err = prefRepo.Upsert(ctx, db.NotificationPreference{
		UserID:          userID,
		EventType:       "request.created",
		ChannelID:       chID,
		Enabled:         false,
		DebounceSeconds: 15,
	})
	require.NoError(t, err)

	store := &notifications.DBPreferenceStore{
		Prefs:    prefRepo,
		Channels: chRepo,
		Profiles: profiles,
		Users:    users,
	}
	prefs, err := store.GetUserPreferences(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, "tr-TR", prefs.Locale)
	assert.False(t, prefs.IsAdmin)
	assert.False(t, prefs.Subscribed[notifications.EventRequestCreated])
	assert.False(t, prefs.EnabledChannels[notifications.ChannelWebhook])
	require.NotNil(t, prefs.DebounceSeconds)
	assert.Equal(t, 15, *prefs.DebounceSeconds)
}

func TestChannelFromDB_EmailWithTLS(t *testing.T) {
	t.Parallel()
	ch, err := notifications.ChannelFromDB(db.NotificationChannel{
		ChannelType: db.NotificationChannelEmail,
		Config:      `{"host":"smtp.example.com","port":465,"username":"u","password":"p","from":"a@b.c","useTls":true}`,
	})
	require.NoError(t, err)
	assert.Equal(t, notifications.ChannelSMTP, ch.ID())
}

func TestDBPreferenceStore_SkipsMissingChannel(t *testing.T) {
	ctx := context.Background()
	d := openNotificationsTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	users := db.NewUserRepo(d.Querier())
	userID := uuid.NewString()
	_, err := users.Create(ctx, userID, "orphan-pref", "hash", []string{auth.RoleUser})
	require.NoError(t, err)

	chRepo := db.NewNotificationChannelRepo(d.Querier())
	prefRepo := db.NewNotificationPreferenceRepo(d.Querier())
	chID, err := chRepo.Create(ctx, db.NotificationChannel{
		ChannelType: db.NotificationChannelWebhook,
		Name:        "temp",
		Config:      `{"url":"https://example.com"}`,
		Enabled:     true,
	})
	require.NoError(t, err)
	_, err = prefRepo.Upsert(ctx, db.NotificationPreference{
		UserID: userID, EventType: "request.created", ChannelID: chID, Enabled: true,
	})
	require.NoError(t, err)
	_, err = d.Querier().ExecContext(ctx, `DELETE FROM notification_channels WHERE id = ?`, chID)
	require.NoError(t, err)

	store := &notifications.DBPreferenceStore{Prefs: prefRepo, Channels: chRepo, Users: users}
	prefs, err := store.GetUserPreferences(ctx, userID)
	require.NoError(t, err)
	assert.Empty(t, prefs.Subscribed)
}

func TestDBPreferenceStore_AdminFlag(t *testing.T) {
	ctx := context.Background()
	d := openNotificationsTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	users := db.NewUserRepo(d.Querier())
	adminID := uuid.NewString()
	_, err := users.Create(ctx, adminID, "notify-admin", "hash", []string{auth.RoleAdmin})
	require.NoError(t, err)

	store := &notifications.DBPreferenceStore{Users: users}
	prefs, err := store.GetUserPreferences(ctx, adminID)
	require.NoError(t, err)
	assert.True(t, prefs.IsAdmin)
}
