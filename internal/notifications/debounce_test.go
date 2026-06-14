package notifications_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/somralab/somra-media/internal/notifications"
)

func TestDebouncerSetWindowAndReset(t *testing.T) {
	t.Parallel()
	d := notifications.NewDebouncer(time.Minute)
	d.SetWindow(time.Second)
	assert.True(t, d.Allow("k"))
	d.Reset()
	assert.True(t, d.Allow("k"))
}

func TestDebouncerSuppressesWithinWindow(t *testing.T) {
	t.Parallel()
	d := notifications.NewDebouncer(100 * time.Millisecond)
	key := notifications.DebounceKey("u1", notifications.EventRequestCreated, notifications.ChannelWebhook)

	assert.True(t, d.Allow(key))
	assert.False(t, d.Allow(key))
}

func TestDebouncerAllowsAfterWindow(t *testing.T) {
	t.Parallel()
	d := notifications.NewDebouncer(20 * time.Millisecond)
	key := "k"

	assert.True(t, d.Allow(key))
	time.Sleep(25 * time.Millisecond)
	assert.True(t, d.Allow(key))
}

func TestDebouncerZeroWindowAlwaysAllows(t *testing.T) {
	t.Parallel()
	d := notifications.NewDebouncer(0)
	assert.True(t, d.Allow("any"))
	assert.True(t, d.Allow("any"))
	assert.True(t, d.AllowWithWindow("any", 0))
}

func TestDebounceKeyFormat(t *testing.T) {
	t.Parallel()
	key := notifications.DebounceKey("user", notifications.EventSystemError, notifications.ChannelDiscord)
	assert.Equal(t, "user|system.error|discord", key)
}
