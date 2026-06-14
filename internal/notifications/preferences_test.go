package notifications_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/notifications"
)

func TestPreferenceFilterDefaultSubscription(t *testing.T) {
	t.Parallel()
	f := notifications.NewPreferenceFilter(&memoryPrefs{byUser: map[string]notifications.UserPreferences{
		"u1": {UserID: "u1"},
	}})

	ok, err := f.ShouldNotify(context.Background(), "u1", notifications.EventRequestApproved)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = f.ShouldNotify(context.Background(), "u1", notifications.EventSystemError)
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestPreferenceFilterExplicitUnsubscribe(t *testing.T) {
	t.Parallel()
	f := notifications.NewPreferenceFilter(&memoryPrefs{byUser: map[string]notifications.UserPreferences{
		"u1": {
			UserID: "u1",
			Subscribed: map[notifications.EventType]bool{
				notifications.EventRequestCreated: false,
			},
		},
	}})

	ok, err := f.ShouldNotify(context.Background(), "u1", notifications.EventRequestCreated)
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestPreferenceFilterAdminSystemErrors(t *testing.T) {
	t.Parallel()
	f := notifications.NewPreferenceFilter(&memoryPrefs{byUser: map[string]notifications.UserPreferences{
		"admin": {UserID: "admin", IsAdmin: true},
	}})

	ok, err := f.ShouldNotify(context.Background(), "admin", notifications.EventSystemError)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestPreferenceFilterAllowedChannelsSubset(t *testing.T) {
	t.Parallel()
	f := notifications.NewPreferenceFilter(&memoryPrefs{byUser: map[string]notifications.UserPreferences{
		"u1": {
			UserID: "u1",
			EnabledChannels: map[notifications.ChannelID]bool{
				notifications.ChannelWebhook: true,
				notifications.ChannelDiscord: false,
			},
		},
	}})

	channels, err := f.AllowedChannels(context.Background(), "u1", notifications.EventRequestCreated)
	require.NoError(t, err)
	require.Len(t, channels, 1)
	assert.Equal(t, notifications.ChannelWebhook, channels[0])
}

func TestPreferenceFilterLocaleFallback(t *testing.T) {
	t.Parallel()
	f := notifications.NewPreferenceFilter(&memoryPrefs{byUser: map[string]notifications.UserPreferences{
		"u1": {UserID: "u1", Locale: "tr-TR"},
	}})

	locale, err := f.Locale(context.Background(), "u1")
	require.NoError(t, err)
	assert.Equal(t, "tr-TR", locale)

	fEmpty := notifications.NewPreferenceFilter(&memoryPrefs{byUser: map[string]notifications.UserPreferences{
		"u2": {UserID: "u2"},
	}})
	locale, err = fEmpty.Locale(context.Background(), "u2")
	require.NoError(t, err)
	assert.Equal(t, "en-US", locale)
}

func TestEventRecipients(t *testing.T) {
	t.Parallel()
	ev := notifications.Event{
		Type:     notifications.EventRequestCreated,
		UserID:   "requester",
		AdminIDs: []string{"admin-1", "admin-2", "requester"},
	}
	recipients := ev.Recipients()
	assert.ElementsMatch(t, []string{"requester", "admin-1", "admin-2"}, recipients)

	sys := notifications.Event{Type: notifications.EventSystemError, UserID: "user", AdminIDs: []string{"admin"}}
	assert.Equal(t, []string{"admin"}, sys.Recipients())
}

func TestPreferenceFilter_AllChannelsWhenNoStore(t *testing.T) {
	t.Parallel()
	f := notifications.NewPreferenceFilter(nil)
	channels, err := f.AllowedChannels(context.Background(), "u1", notifications.EventRequestCreated)
	require.NoError(t, err)
	require.Len(t, channels, 3)
}

func TestPreferenceFilter_AllChannelsWhenUnsubscribed(t *testing.T) {
	t.Parallel()
	f := notifications.NewPreferenceFilter(&memoryPrefs{byUser: map[string]notifications.UserPreferences{
		"u1": {
			UserID: "u1",
			Subscribed: map[notifications.EventType]bool{
				notifications.EventRequestCreated: false,
			},
		},
	}})
	channels, err := f.AllowedChannels(context.Background(), "u1", notifications.EventRequestCreated)
	require.NoError(t, err)
	assert.Empty(t, channels)
}

func TestPreferenceFilter_UnknownEvent(t *testing.T) {
	t.Parallel()
	f := notifications.NewPreferenceFilter(&memoryPrefs{byUser: map[string]notifications.UserPreferences{
		"u1": {UserID: "u1"},
	}})
	ok, err := f.ShouldNotify(context.Background(), "u1", notifications.EventType("unknown.event"))
	require.NoError(t, err)
	assert.False(t, ok)
}
