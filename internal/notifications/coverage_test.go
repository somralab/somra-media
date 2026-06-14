package notifications_test

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/jobs"
	"github.com/somralab/somra-media/internal/notifications"
	"github.com/somralab/somra-media/internal/platform/db"
)

func TestLoadChannelConfig_NilReader(t *testing.T) {
	t.Parallel()
	_, err := notifications.LoadChannelConfig(context.Background(), nil)
	require.Error(t, err)
}

func TestBuildChannels_InvalidDiscord(t *testing.T) {
	t.Parallel()
	channels, err := notifications.BuildChannels(context.Background(), mapSettings{
		notifications.KeyDiscordEnabled:    "true",
		notifications.KeyDiscordWebhookURL: "",
	}, nil)
	require.NoError(t, err)
	assert.Empty(t, channels)
}

func TestEventTemplateDataMergesFields(t *testing.T) {
	t.Parallel()
	ev := notifications.Event{
		Title: "T",
		Data:  map[string]any{"requestId": 42},
	}
	data := ev.TemplateData()
	assert.Equal(t, "T", data["Title"])
	assert.Equal(t, 42, data["requestId"])
}

func TestDispatcherSkipsUnregisteredChannel(t *testing.T) {
	t.Parallel()
	q := jobs.NewMemoryQueue(jobs.MemoryQueueConfig{Workers: 1, Buffer: 2})
	t.Cleanup(q.Close)
	d := notifications.NewDispatcher(notifications.DispatcherConfig{
		Renderer: notifications.NewTemplateRenderer(testBundle(t)),
		Filter:   notifications.NewPreferenceFilter(nil),
		Debounce: notifications.NewDebouncer(0),
		Queue:    q,
	})
	require.NoError(t, d.Dispatch(context.Background(), notifications.Event{
		Type:   notifications.EventRequestCreated,
		UserID: "user-1",
		Title:  "No Channel",
	}))
}

func TestDispatcherDebouncesRepeat(t *testing.T) {
	t.Parallel()
	d, ch := testDispatcher(t, &memoryPrefs{byUser: map[string]notifications.UserPreferences{
		"user-1": {UserID: "user-1", EnabledChannels: map[notifications.ChannelID]bool{
			notifications.ChannelWebhook: true,
		}},
	}}, notifications.NewDebouncer(time.Minute))

	ev := notifications.Event{UserID: "user-1", Title: "Once"}
	require.NoError(t, d.HandleRequestCreated(context.Background(), ev))
	require.NoError(t, d.HandleRequestCreated(context.Background(), ev))
	require.Eventually(t, func() bool { return ch.count() == 1 }, time.Second, 5*time.Millisecond)
}

func TestDispatcherDebounceOverride(t *testing.T) {
	t.Parallel()
	sec := 120
	d, ch := testDispatcher(t, &memoryPrefs{byUser: map[string]notifications.UserPreferences{
		"user-1": {
			UserID:          "user-1",
			DebounceSeconds: &sec,
			EnabledChannels: map[notifications.ChannelID]bool{notifications.ChannelWebhook: true},
		},
	}}, notifications.NewDebouncer(time.Minute))

	ev := notifications.Event{UserID: "user-1", Title: "Debounced"}
	require.NoError(t, d.HandleRequestCreated(context.Background(), ev))
	require.NoError(t, d.HandleRequestCreated(context.Background(), ev))
	require.Eventually(t, func() bool { return ch.count() == 1 }, time.Second, 5*time.Millisecond)
}

func TestDispatcherEnqueueSendFailure(t *testing.T) {
	t.Parallel()
	q := jobs.NewMemoryQueue(jobs.MemoryQueueConfig{Workers: 0, Buffer: 1})
	q.Close()
	d := notifications.NewDispatcher(notifications.DispatcherConfig{
		Renderer: notifications.NewTemplateRenderer(testBundle(t)),
		Filter: notifications.NewPreferenceFilter(&memoryPrefs{byUser: map[string]notifications.UserPreferences{
			"u": {UserID: "u", EnabledChannels: map[notifications.ChannelID]bool{
				notifications.ChannelWebhook: true,
			}},
		}}),
		Debounce: notifications.NewDebouncer(0),
		Queue:    q,
	})
	ch := &recordingChannel{id: notifications.ChannelWebhook}
	d.RegisterChannel(ch)
	err := d.Dispatch(context.Background(), notifications.Event{
		Type:   notifications.EventRequestCreated,
		UserID: "u",
		Title:  "fail",
	})
	require.Error(t, err)
}

func TestDiscordChannelColorsAndErrors(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("bad"))
	}))
	t.Cleanup(srv.Close)

	ch, err := notifications.NewDiscordChannel(notifications.DiscordConfig{WebhookURL: srv.URL, HTTPClient: srv.Client()})
	require.NoError(t, err)
	err = ch.Send(context.Background(), notifications.Notification{
		EventType: notifications.EventRequestRejected,
		Subject:   "Rejected",
		Body:      "No",
		SentAt:    time.Now(),
	})
	require.Error(t, err)
}

func TestChannelFromDB_DiscordURLFallback(t *testing.T) {
	t.Parallel()
	ch, err := notifications.ChannelFromDB(db.NotificationChannel{
		ChannelType: db.NotificationChannelDiscord,
		Config:      `{"url":"https://discord.com/api/webhooks/fallback"}`,
	})
	require.NoError(t, err)
	assert.Equal(t, notifications.ChannelDiscord, ch.ID())
}

func TestDispatcher_RegisterChannelNilSafe(t *testing.T) {
	t.Parallel()
	var d *notifications.Dispatcher
	d.RegisterChannel(nil)
	q := jobs.NewMemoryQueue(jobs.MemoryQueueConfig{Workers: 1, Buffer: 2})
	t.Cleanup(q.Close)
	d2 := notifications.NewDispatcher(notifications.DispatcherConfig{Queue: q})
	d2.RegisterChannel(nil)
	ch := &recordingChannel{id: notifications.ChannelWebhook}
	d2.RegisterChannel(ch)
	got, ok := d2.Channel(notifications.ChannelWebhook)
	require.True(t, ok)
	assert.Equal(t, ch, got)
}

func TestDispatcher_SendFailureLogged(t *testing.T) {
	t.Parallel()
	q := jobs.NewMemoryQueue(jobs.MemoryQueueConfig{Workers: 1, Buffer: 4, Logger: slog.Default()})
	t.Cleanup(q.Close)
	ch := &recordingChannel{id: notifications.ChannelWebhook, sendErr: errors.New("delivery failed")}
	d := notifications.NewDispatcher(notifications.DispatcherConfig{
		Renderer: notifications.NewTemplateRenderer(testBundle(t)),
		Filter: notifications.NewPreferenceFilter(&memoryPrefs{byUser: map[string]notifications.UserPreferences{
			"u": {UserID: "u", EnabledChannels: map[notifications.ChannelID]bool{notifications.ChannelWebhook: true}},
		}}),
		Debounce: notifications.NewDebouncer(0),
		Queue:    q,
		Logger:   slog.Default(),
	})
	d.RegisterChannel(ch)
	require.NoError(t, d.Dispatch(context.Background(), notifications.Event{
		Type:   notifications.EventRequestCreated,
		UserID: "u",
		Title:  "Fail send",
	}))
	require.Eventually(t, func() bool { return ch.count() == 1 }, time.Second, 5*time.Millisecond)
}

func TestDispatcherEmptyRecipients(t *testing.T) {
	t.Parallel()
	d, _ := testDispatcher(t, &memoryPrefs{byUser: map[string]notifications.UserPreferences{}}, notifications.NewDebouncer(0))
	require.NoError(t, d.Dispatch(context.Background(), notifications.Event{
		Type: notifications.EventRequestCreated,
	}))
}

func TestWebhookChannelNilSend(t *testing.T) {
	t.Parallel()
	var ch *notifications.WebhookChannel
	err := ch.Send(context.Background(), notifications.Notification{Subject: "x", Body: "y"})
	require.Error(t, err)
}

func TestRecordingChannelSendError(t *testing.T) {
	t.Parallel()
	ch := &recordingChannel{id: notifications.ChannelWebhook, sendErr: errors.New("send failed")}
	err := ch.Send(context.Background(), notifications.Notification{Subject: "x", Body: "y"})
	require.Error(t, err)
}
