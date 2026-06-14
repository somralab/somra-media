package notifications_test

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/auth"
	"github.com/somralab/somra-media/internal/notifications"
	"github.com/somralab/somra-media/internal/platform/db"
)

func TestDBPreferenceStore_ChannelTypes(t *testing.T) {
	ctx := context.Background()
	d := openNotificationsTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	users := db.NewUserRepo(d.Querier())
	userID := uuid.NewString()
	_, err := users.Create(ctx, userID, "multi-ch", "hash", []string{auth.RoleUser})
	require.NoError(t, err)

	chRepo := db.NewNotificationChannelRepo(d.Querier())
	prefRepo := db.NewNotificationPreferenceRepo(d.Querier())

	discordID, err := chRepo.Create(ctx, db.NotificationChannel{
		ChannelType: db.NotificationChannelDiscord,
		Name:        "discord",
		Config:      `{"webhookUrl":"https://discord.com/api/webhooks/x"}`,
		Enabled:     true,
	})
	require.NoError(t, err)
	emailID, err := chRepo.Create(ctx, db.NotificationChannel{
		ChannelType: db.NotificationChannelEmail,
		Name:        "email",
		Config:      `{"host":"smtp.example.com","from":"a@b.c"}`,
		Enabled:     true,
	})
	require.NoError(t, err)

	_, err = prefRepo.Upsert(ctx, db.NotificationPreference{
		UserID: userID, EventType: "request.approved", ChannelID: discordID, Enabled: true,
	})
	require.NoError(t, err)
	_, err = prefRepo.Upsert(ctx, db.NotificationPreference{
		UserID: userID, EventType: "request.completed", ChannelID: emailID, Enabled: true,
	})
	require.NoError(t, err)

	store := &notifications.DBPreferenceStore{Prefs: prefRepo, Channels: chRepo, Users: users}
	prefs, err := store.GetUserPreferences(ctx, userID)
	require.NoError(t, err)
	assert.True(t, prefs.EnabledChannels[notifications.ChannelDiscord])
	assert.True(t, prefs.EnabledChannels[notifications.ChannelSMTP])
}

func TestLoadChannelConfig_ReaderError(t *testing.T) {
	t.Parallel()
	_, err := notifications.LoadChannelConfig(context.Background(), mapSettings{
		notifications.KeyWebhookEnabled: "not-bool",
	})
	require.NoError(t, err)

	failReader := mapSettings{"notifications.webhook.enabled": "true"}
	failReader["notifications.webhook.url"] = ""
	_, err = notifications.LoadChannelConfig(context.Background(), failingReader{err: errors.New("boom")})
	require.Error(t, err)
}

type failingReader struct {
	err error
}

func (f failingReader) GetString(_ context.Context, _, _ string) (string, error) {
	return "", f.err
}

func TestNewSMTPChannelValidation(t *testing.T) {
	t.Parallel()
	_, err := notifications.NewSMTPChannel(notifications.SMTPConfig{})
	require.Error(t, err)
	_, err = notifications.NewSMTPChannel(notifications.SMTPConfig{Host: "h", Port: 0, From: "a@b.c"})
	require.Error(t, err)
	_, err = notifications.NewSMTPChannel(notifications.SMTPConfig{Host: "h", Port: 25, From: ""})
	require.Error(t, err)
}

func TestSMTPChannelSendWithAuth(t *testing.T) {
	t.Parallel()
	ch, err := notifications.NewSMTPChannel(notifications.SMTPConfig{
		Host:     "127.0.0.1",
		Port:     25,
		From:     "somra@test.local",
		Username: "user",
		Password: "pass",
		UseTLS:   false,
		Dial: func(_, _ string) (net.Conn, error) {
			return nil, errors.New("dial blocked in test")
		},
	})
	require.NoError(t, err)
	err = ch.Send(context.Background(), notifications.Notification{
		Subject: "x",
		Body:    "y",
		Data:    map[string]any{"email": "to@test.local"},
	})
	require.Error(t, err)
}
