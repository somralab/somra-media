package notifications_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/notifications"
)

func TestWebhookChannelSend(t *testing.T) {
	t.Parallel()
	var got struct {
		EventType string `json:"eventType"`
		Subject   string `json:"subject"`
		Body      string `json:"body"`
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	ch, err := notifications.NewWebhookChannel(notifications.WebhookConfig{URL: srv.URL, HTTPClient: srv.Client()})
	require.NoError(t, err)

	err = ch.Send(context.Background(), notifications.Notification{
		EventType: notifications.EventRequestCreated,
		Subject:   "Hello",
		Body:      "World",
		Locale:    "en-US",
		UserID:    "u1",
		SentAt:    time.Now(),
	})
	require.NoError(t, err)
	assert.Equal(t, "request.created", got.EventType)
	assert.Equal(t, "Hello", got.Subject)
}

func TestWebhookChannelHTTPError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = io.WriteString(w, "fail")
	}))
	t.Cleanup(srv.Close)

	ch, err := notifications.NewWebhookChannel(notifications.WebhookConfig{URL: srv.URL, HTTPClient: srv.Client()})
	require.NoError(t, err)
	err = ch.Send(context.Background(), notifications.Notification{Subject: "x", Body: "y", SentAt: time.Now()})
	require.Error(t, err)
}

func TestWebhookChannelRequiresURL(t *testing.T) {
	t.Parallel()
	_, err := notifications.NewWebhookChannel(notifications.WebhookConfig{})
	require.Error(t, err)
}

func TestWebhookChannelID(t *testing.T) {
	t.Parallel()
	ch, err := notifications.NewWebhookChannel(notifications.WebhookConfig{URL: "https://example.com/hook"})
	require.NoError(t, err)
	assert.Equal(t, notifications.ChannelWebhook, ch.ID())
}

func TestDiscordChannelSend(t *testing.T) {
	t.Parallel()
	var got struct {
		Embeds []struct {
			Title       string `json:"title"`
			Description string `json:"description"`
		} `json:"embeds"`
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	ch, err := notifications.NewDiscordChannel(notifications.DiscordConfig{WebhookURL: srv.URL, HTTPClient: srv.Client()})
	require.NoError(t, err)

	err = ch.Send(context.Background(), notifications.Notification{
		EventType: notifications.EventRequestApproved,
		Subject:   "Approved",
		Body:      "Done",
		SentAt:    time.Now(),
	})
	require.NoError(t, err)
	require.Len(t, got.Embeds, 1)
	assert.Equal(t, "Approved", got.Embeds[0].Title)
}

func TestDiscordChannelRequiresURL(t *testing.T) {
	t.Parallel()
	_, err := notifications.NewDiscordChannel(notifications.DiscordConfig{})
	require.Error(t, err)
}

func TestDiscordChannelID(t *testing.T) {
	t.Parallel()
	ch, err := notifications.NewDiscordChannel(notifications.DiscordConfig{WebhookURL: "https://discord.com/api/webhooks/x"})
	require.NoError(t, err)
	assert.Equal(t, notifications.ChannelDiscord, ch.ID())
}
