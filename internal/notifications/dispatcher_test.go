package notifications_test

import (
	"context"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/jobs"
	"github.com/somralab/somra-media/internal/notifications"
	"github.com/somralab/somra-media/internal/platform/i18n"
)

type memoryPrefs struct {
	byUser map[string]notifications.UserPreferences
}

func (m *memoryPrefs) GetUserPreferences(_ context.Context, userID string) (notifications.UserPreferences, error) {
	if p, ok := m.byUser[userID]; ok {
		return p, nil
	}
	return notifications.UserPreferences{UserID: userID, Locale: "en-US"}, nil
}

type recordingChannel struct {
	id      notifications.ChannelID
	mu      sync.Mutex
	sent    []notifications.Notification
	sendErr error
}

func (c *recordingChannel) ID() notifications.ChannelID { return c.id }

func (c *recordingChannel) Send(_ context.Context, n notifications.Notification) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sent = append(c.sent, n)
	return c.sendErr
}

func (c *recordingChannel) count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.sent)
}

func testBundle(t *testing.T) *i18n.Bundle {
	t.Helper()
	b, err := i18n.NewBundle()
	require.NoError(t, err)
	return b
}

func testDispatcher(t *testing.T, prefs notifications.PreferenceStore, debounce *notifications.Debouncer) (*notifications.Dispatcher, *recordingChannel) {
	t.Helper()
	q := jobs.NewMemoryQueue(jobs.MemoryQueueConfig{Workers: 2, Buffer: 8, Logger: slog.Default()})
	t.Cleanup(q.Close)

	ch := &recordingChannel{id: notifications.ChannelWebhook}
	d := notifications.NewDispatcher(notifications.DispatcherConfig{
		Renderer: notifications.NewTemplateRenderer(testBundle(t)),
		Filter:   notifications.NewPreferenceFilter(prefs),
		Debounce: debounce,
		Queue:    q,
	})
	d.RegisterChannel(ch)
	return d, ch
}

func TestDispatcherDispatchesAsync(t *testing.T) {
	t.Parallel()
	d, ch := testDispatcher(t, &memoryPrefs{byUser: map[string]notifications.UserPreferences{
		"user-1": {UserID: "user-1", Locale: "en-US", EnabledChannels: map[notifications.ChannelID]bool{
			notifications.ChannelWebhook: true,
		}},
	}}, notifications.NewDebouncer(0))

	err := d.HandleRequestCreated(context.Background(), notifications.Event{
		UserID: "user-1",
		Title:  "Dune",
		Detail: "4K please",
	})
	require.NoError(t, err)

	require.Eventually(t, func() bool { return ch.count() == 1 }, time.Second, 5*time.Millisecond)
	ch.mu.Lock()
	defer ch.mu.Unlock()
	require.Len(t, ch.sent, 1)
	assert.Equal(t, notifications.EventRequestCreated, ch.sent[0].EventType)
	assert.Contains(t, ch.sent[0].Subject, "Dune")
}

func TestDispatcherRespectsUnsubscribed(t *testing.T) {
	t.Parallel()
	d, ch := testDispatcher(t, &memoryPrefs{byUser: map[string]notifications.UserPreferences{
		"user-1": {
			UserID: "user-1",
			Subscribed: map[notifications.EventType]bool{
				notifications.EventRequestCreated: false,
			},
			EnabledChannels: map[notifications.ChannelID]bool{notifications.ChannelWebhook: true},
		},
	}}, notifications.NewDebouncer(0))

	err := d.HandleRequestCreated(context.Background(), notifications.Event{UserID: "user-1", Title: "X"})
	require.NoError(t, err)
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 0, ch.count())
}

func TestDispatcherDebouncesDuplicate(t *testing.T) {
	t.Parallel()
	d, ch := testDispatcher(t, &memoryPrefs{byUser: map[string]notifications.UserPreferences{
		"user-1": {UserID: "user-1", EnabledChannels: map[notifications.ChannelID]bool{
			notifications.ChannelWebhook: true,
		}},
	}}, notifications.NewDebouncer(time.Minute))

	ev := notifications.Event{UserID: "user-1", Title: "A"}
	require.NoError(t, d.HandleRequestCreated(context.Background(), ev))
	require.NoError(t, d.HandleRequestCreated(context.Background(), ev))

	require.Eventually(t, func() bool { return ch.count() == 1 }, time.Second, 5*time.Millisecond)
}

func TestDispatcherLocalizedTemplate(t *testing.T) {
	t.Parallel()
	d, ch := testDispatcher(t, &memoryPrefs{byUser: map[string]notifications.UserPreferences{
		"user-tr": {UserID: "user-tr", Locale: "tr-TR", EnabledChannels: map[notifications.ChannelID]bool{
			notifications.ChannelWebhook: true,
		}},
	}}, notifications.NewDebouncer(0))

	require.NoError(t, d.HandleRequestApproved(context.Background(), notifications.Event{
		UserID: "user-tr",
		Title:  "Breaking Bad",
	}))
	require.Eventually(t, func() bool { return ch.count() == 1 }, time.Second, 5*time.Millisecond)
	ch.mu.Lock()
	defer ch.mu.Unlock()
	assert.Contains(t, ch.sent[0].Subject, "onaylandı")
	assert.Equal(t, "tr-TR", ch.sent[0].Locale)
}

func TestDispatcherSystemErrorAdminOnly(t *testing.T) {
	t.Parallel()
	d, ch := testDispatcher(t, &memoryPrefs{byUser: map[string]notifications.UserPreferences{
		"admin-1": {UserID: "admin-1", IsAdmin: true, EnabledChannels: map[notifications.ChannelID]bool{
			notifications.ChannelWebhook: true,
		}},
		"user-1": {UserID: "user-1", IsAdmin: false, EnabledChannels: map[notifications.ChannelID]bool{
			notifications.ChannelWebhook: true,
		}},
	}}, notifications.NewDebouncer(0))

	require.NoError(t, d.HandleSystemError(context.Background(), notifications.Event{
		AdminIDs: []string{"admin-1"},
		UserID:   "user-1",
		ErrorMsg: "disk full",
	}))
	require.Eventually(t, func() bool { return ch.count() == 1 }, time.Second, 5*time.Millisecond)
}

func TestDispatcherHandleRejectedAndCompleted(t *testing.T) {
	t.Parallel()
	d, ch := testDispatcher(t, &memoryPrefs{byUser: map[string]notifications.UserPreferences{
		"user-1": {UserID: "user-1", EnabledChannels: map[notifications.ChannelID]bool{
			notifications.ChannelWebhook: true,
		}},
	}}, notifications.NewDebouncer(0))

	require.NoError(t, d.HandleRequestRejected(context.Background(), notifications.Event{
		UserID: "user-1",
		Title:  "Denied",
	}))
	require.NoError(t, d.HandleRequestCompleted(context.Background(), notifications.Event{
		UserID: "user-1",
		Title:  "Done",
	}))
	require.Eventually(t, func() bool { return ch.count() == 2 }, time.Second, 5*time.Millisecond)
}

func TestDispatcherNilReceiver(t *testing.T) {
	t.Parallel()
	var d *notifications.Dispatcher
	err := d.Dispatch(context.Background(), notifications.Event{UserID: "u"})
	require.Error(t, err)
}

func TestDispatcherMissingRenderer(t *testing.T) {
	t.Parallel()
	q := jobs.NewMemoryQueue(jobs.MemoryQueueConfig{Workers: 1, Buffer: 2})
	t.Cleanup(q.Close)
	d := notifications.NewDispatcher(notifications.DispatcherConfig{Queue: q})
	err := d.Dispatch(context.Background(), notifications.Event{UserID: "u", Type: notifications.EventRequestCreated})
	require.Error(t, err)
}

func TestPreferenceFilterDebounceSeconds(t *testing.T) {
	t.Parallel()
	sec := 45
	f := notifications.NewPreferenceFilter(&memoryPrefs{byUser: map[string]notifications.UserPreferences{
		"u1": {UserID: "u1", DebounceSeconds: &sec},
	}})
	got, err := f.DebounceSeconds(context.Background(), "u1")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, 45, *got)
}
