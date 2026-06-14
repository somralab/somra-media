package notifications

import (
	"sync"
	"time"
)

// Debouncer suppresses duplicate deliveries within a sliding window.
type Debouncer struct {
	mu     sync.Mutex
	window time.Duration
	last   map[string]time.Time
}

// NewDebouncer returns a debouncer with the given window. Zero window
// disables debouncing (Allow always returns true).
func NewDebouncer(window time.Duration) *Debouncer {
	return &Debouncer{
		window: window,
		last:   make(map[string]time.Time),
	}
}

// Allow reports whether key may proceed. When allowed, the key timestamp
// is updated; otherwise the call is suppressed until the window elapses.
func (d *Debouncer) Allow(key string) bool {
	return d.AllowWithWindow(key, d.window)
}

// AllowWithWindow applies a per-call debounce window. Zero window disables
// debouncing for that check.
func (d *Debouncer) AllowWithWindow(key string, window time.Duration) bool {
	if d == nil || window <= 0 || key == "" {
		return true
	}
	now := time.Now()
	d.mu.Lock()
	defer d.mu.Unlock()
	if last, ok := d.last[key]; ok && now.Sub(last) < window {
		return false
	}
	d.last[key] = now
	return true
}

// DebounceKey builds a stable suppression key.
func DebounceKey(userID string, eventType EventType, channelID ChannelID) string {
	return userID + "|" + string(eventType) + "|" + string(channelID)
}

// SetWindow updates the debounce window at runtime (e.g. after settings reload).
func (d *Debouncer) SetWindow(window time.Duration) {
	if d == nil {
		return
	}
	d.mu.Lock()
	d.window = window
	d.mu.Unlock()
}

// windowForDefault returns the configured default window.
func (d *Debouncer) windowForDefault() time.Duration {
	if d == nil {
		return 0
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.window
}

// Reset clears all tracked keys (used in tests).
func (d *Debouncer) Reset() {
	if d == nil {
		return
	}
	d.mu.Lock()
	d.last = make(map[string]time.Time)
	d.mu.Unlock()
}
