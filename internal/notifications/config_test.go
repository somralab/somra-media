package notifications_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/notifications"
)

type mapSettings map[string]string

func (m mapSettings) GetString(_ context.Context, key, defaultVal string) (string, error) {
	if v, ok := m[key]; ok {
		return v, nil
	}
	return defaultVal, nil
}

func TestLoadChannelConfig(t *testing.T) {
	t.Parallel()
	cfg, err := notifications.LoadChannelConfig(context.Background(), mapSettings{
		notifications.KeyWebhookEnabled:     "true",
		notifications.KeyWebhookURL:         "https://example.com/hook",
		notifications.KeyDiscordEnabled:     "true",
		notifications.KeyDiscordWebhookURL:  "https://discord.com/api/webhooks/x",
		notifications.KeySMTPEnabled:        "true",
		notifications.KeySMTPHost:           "smtp.example.com",
		notifications.KeySMTPPort:           "465",
		notifications.KeySMTPFrom:           "somra@example.com",
		notifications.KeyDebounceSeconds:    "30",
		notifications.KeyDefaultAdminEmails: `["admin@example.com"]`,
	})
	require.NoError(t, err)
	assert.True(t, cfg.WebhookEnabled)
	assert.Equal(t, "https://example.com/hook", cfg.WebhookURL)
	assert.True(t, cfg.DiscordEnabled)
	assert.True(t, cfg.SMTPEnabled)
	assert.Equal(t, 465, cfg.SMTPPort)
	assert.Equal(t, 30, cfg.DebounceSeconds)
	assert.Equal(t, []string{"admin@example.com"}, cfg.AdminEmails)
}

func TestBuildChannelsFromSettings(t *testing.T) {
	t.Parallel()
	channels, err := notifications.BuildChannels(context.Background(), mapSettings{
		notifications.KeyWebhookEnabled: "true",
		notifications.KeyWebhookURL:     "https://example.com/hook",
		notifications.KeyDiscordEnabled: "false",
		notifications.KeySMTPEnabled:    "false",
	}, nil)
	require.NoError(t, err)
	require.Len(t, channels, 1)
	assert.Equal(t, notifications.ChannelWebhook, channels[0].ID())
}

func TestBuildChannels_AllEnabled(t *testing.T) {
	t.Parallel()
	channels, err := notifications.BuildChannels(context.Background(), mapSettings{
		notifications.KeyWebhookEnabled:    "true",
		notifications.KeyWebhookURL:          "https://example.com/hook",
		notifications.KeyDiscordEnabled:      "true",
		notifications.KeyDiscordWebhookURL:   "https://discord.com/api/webhooks/x",
		notifications.KeySMTPEnabled:         "true",
		notifications.KeySMTPHost:            "smtp.example.com",
		notifications.KeySMTPFrom:            "somra@example.com",
	}, nil)
	require.NoError(t, err)
	require.Len(t, channels, 3)
}

func TestLoadChannelConfig_ReaderErrorsOnKeys(t *testing.T) {
	t.Parallel()
	keys := []string{
		notifications.KeyWebhookEnabled,
		notifications.KeyWebhookURL,
		notifications.KeyDiscordEnabled,
		notifications.KeyDiscordWebhookURL,
		notifications.KeySMTPEnabled,
		notifications.KeySMTPHost,
		notifications.KeySMTPPort,
		notifications.KeySMTPUsername,
		notifications.KeySMTPPassword,
		notifications.KeySMTPFrom,
		notifications.KeySMTPUseTLS,
		notifications.KeyDebounceSeconds,
		notifications.KeyDefaultAdminEmails,
	}
	for _, key := range keys {
		_, err := notifications.LoadChannelConfig(context.Background(), singleKeyFailReader{key: key})
		require.Error(t, err, "key %s", key)
	}
}

type singleKeyFailReader struct{ key string }

func (s singleKeyFailReader) GetString(_ context.Context, key, defaultVal string) (string, error) {
	if key == s.key {
		return "", errors.New("reader failed")
	}
	switch key {
	case notifications.KeyWebhookEnabled, notifications.KeyDiscordEnabled, notifications.KeySMTPEnabled:
		return "false", nil
	case notifications.KeySMTPUseTLS:
		return "true", nil
	}
	return defaultVal, nil
}

func TestChannelConfig_DebounceWindow(t *testing.T) {
	t.Parallel()
	assert.Equal(t, time.Duration(0), notifications.ChannelConfig{DebounceSeconds: 0}.DebounceWindow())
	assert.Equal(t, 30*time.Second, notifications.ChannelConfig{DebounceSeconds: 30}.DebounceWindow())
}

func TestLoadChannelConfig_InvalidPort(t *testing.T) {
	t.Parallel()
	_, err := notifications.LoadChannelConfig(context.Background(), mapSettings{
		notifications.KeySMTPEnabled: "true",
		notifications.KeySMTPPort:    "99999",
	})
	require.Error(t, err)
}

func TestParseEmailList_InvalidJSON(t *testing.T) {
	t.Parallel()
	cfg, err := notifications.LoadChannelConfig(context.Background(), mapSettings{
		notifications.KeyDefaultAdminEmails: "not-json",
	})
	require.NoError(t, err)
	assert.Nil(t, cfg.AdminEmails)
}

func TestLoadChannelConfig_EmptyEmailList(t *testing.T) {
	t.Parallel()
	cfg, err := notifications.LoadChannelConfig(context.Background(), mapSettings{
		notifications.KeyDefaultAdminEmails: "[]",
	})
	require.NoError(t, err)
	assert.Empty(t, cfg.AdminEmails)

	cfg, err = notifications.LoadChannelConfig(context.Background(), mapSettings{
		notifications.KeyDefaultAdminEmails: "",
	})
	require.NoError(t, err)
	assert.Nil(t, cfg.AdminEmails)
}

func TestLoadChannelConfig_InvalidDebounce(t *testing.T) {
	t.Parallel()
	_, err := notifications.LoadChannelConfig(context.Background(), mapSettings{
		notifications.KeyDebounceSeconds: "bad",
	})
	require.Error(t, err)
}

func TestLoadChannelConfig_InvalidSMTPPortWhenEnabled(t *testing.T) {
	t.Parallel()
	_, err := notifications.LoadChannelConfig(context.Background(), mapSettings{
		notifications.KeySMTPEnabled: "true",
		notifications.KeySMTPPort:    "70000",
	})
	require.Error(t, err)
}

func TestDiscordChannelApprovedAndCompletedColors(t *testing.T) {
	t.Parallel()
	var colors []int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Embeds []struct{ Color int `json:"color"` } `json:"embeds"`
		}
		_ = json.NewDecoder(r.Body).Decode(&payload)
		if len(payload.Embeds) > 0 {
			colors = append(colors, payload.Embeds[0].Color)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	ch, err := notifications.NewDiscordChannel(notifications.DiscordConfig{WebhookURL: srv.URL, HTTPClient: srv.Client()})
	require.NoError(t, err)
	now := time.Now()
	require.NoError(t, ch.Send(context.Background(), notifications.Notification{
		EventType: notifications.EventRequestApproved, Subject: "A", Body: "B", SentAt: now,
	}))
	require.NoError(t, ch.Send(context.Background(), notifications.Notification{
		EventType: notifications.EventSystemError, Subject: "E", Body: "B", SentAt: now,
	}))
	require.Len(t, colors, 2)
	assert.Equal(t, 0x57F287, colors[0])
	assert.Equal(t, 0xED4245, colors[1])
}

func TestPreferenceFilter_StoreError(t *testing.T) {
	t.Parallel()
	f := notifications.NewPreferenceFilter(errPrefStore{err: errors.New("db down")})
	_, err := f.ShouldNotify(context.Background(), "u", notifications.EventRequestCreated)
	require.Error(t, err)
}

type errPrefStore struct{ err error }

func (e errPrefStore) GetUserPreferences(_ context.Context, _ string) (notifications.UserPreferences, error) {
	return notifications.UserPreferences{}, e.err
}
