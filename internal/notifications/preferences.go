package notifications

import (
	"context"
	"fmt"
)

// UserPreferences describes per-user notification opt-in state.
type UserPreferences struct {
	UserID          string
	Locale          string
	Subscribed      map[EventType]bool
	EnabledChannels map[ChannelID]bool
	DebounceSeconds *int
	Email           string
	IsAdmin         bool
}

// PreferenceStore loads notification preferences. Implementations may
// be backed by DB tables from Sprint 08 Wave 1A; until then callers
// supply test doubles or in-memory stores.
type PreferenceStore interface {
	GetUserPreferences(ctx context.Context, userID string) (UserPreferences, error)
}

// PreferenceFilter applies subscription and channel rules before send.
type PreferenceFilter struct {
	store PreferenceStore
}

// NewPreferenceFilter returns a filter backed by store.
func NewPreferenceFilter(store PreferenceStore) *PreferenceFilter {
	return &PreferenceFilter{store: store}
}

// ShouldNotify reports whether userID wants eventType notifications.
func (f *PreferenceFilter) ShouldNotify(ctx context.Context, userID string, eventType EventType) (bool, error) {
	if f == nil || f.store == nil {
		return true, nil
	}
	prefs, err := f.store.GetUserPreferences(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("notifications: load preferences for %q: %w", userID, err)
	}
	if len(prefs.Subscribed) == 0 {
		return defaultSubscription(eventType, prefs.IsAdmin), nil
	}
	subscribed, ok := prefs.Subscribed[eventType]
	if !ok {
		return defaultSubscription(eventType, prefs.IsAdmin), nil
	}
	return subscribed, nil
}

// AllowedChannels returns channel IDs enabled for userID and eventType.
func (f *PreferenceFilter) AllowedChannels(ctx context.Context, userID string, eventType EventType) ([]ChannelID, error) {
	ok, err := f.ShouldNotify(ctx, userID, eventType)
	if err != nil || !ok {
		return nil, err
	}
	if f == nil || f.store == nil {
		return allChannels(), nil
	}
	prefs, err := f.store.GetUserPreferences(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("notifications: load preferences for %q: %w", userID, err)
	}
	if len(prefs.EnabledChannels) == 0 {
		return allChannels(), nil
	}
	out := make([]ChannelID, 0, len(prefs.EnabledChannels))
	for id, enabled := range prefs.EnabledChannels {
		if enabled {
			out = append(out, id)
		}
	}
	return out, nil
}

// Locale returns the recipient locale or en-US when unset.
func (f *PreferenceFilter) Locale(ctx context.Context, userID string) (string, error) {
	if f == nil || f.store == nil {
		return fallbackLocale, nil
	}
	prefs, err := f.store.GetUserPreferences(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("notifications: load preferences for %q: %w", userID, err)
	}
	if prefs.Locale == "" {
		return fallbackLocale, nil
	}
	return prefs.Locale, nil
}

// DebounceSeconds returns per-user debounce override when set.
func (f *PreferenceFilter) DebounceSeconds(ctx context.Context, userID string) (*int, error) {
	if f == nil || f.store == nil {
		return nil, nil
	}
	prefs, err := f.store.GetUserPreferences(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("notifications: load preferences for %q: %w", userID, err)
	}
	return prefs.DebounceSeconds, nil
}

func defaultSubscription(eventType EventType, isAdmin bool) bool {
	switch eventType {
	case EventRequestCreated, EventRequestApproved, EventRequestRejected, EventRequestCompleted:
		return true
	case EventSystemError:
		return isAdmin
	default:
		return false
	}
}

func allChannels() []ChannelID {
	return []ChannelID{ChannelWebhook, ChannelDiscord, ChannelSMTP}
}
