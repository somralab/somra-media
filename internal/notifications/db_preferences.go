package notifications

import (
	"context"
	"fmt"

	"github.com/somralab/somra-media/internal/auth"
	"github.com/somralab/somra-media/internal/platform/db"
)

// DBPreferenceStore loads notification preferences from SQLite tables.
type DBPreferenceStore struct {
	Prefs    *db.NotificationPreferenceRepo
	Channels *db.NotificationChannelRepo
	Profiles *db.ProfileRepo
	Users    *db.UserRepo
}

// GetUserPreferences implements PreferenceStore.
func (s *DBPreferenceStore) GetUserPreferences(ctx context.Context, userID string) (UserPreferences, error) {
	if s == nil {
		return UserPreferences{}, fmt.Errorf("notifications: preference store is nil")
	}
	out := UserPreferences{
		UserID:          userID,
		Locale:          fallbackLocale,
		Subscribed:      make(map[EventType]bool),
		EnabledChannels: make(map[ChannelID]bool),
	}
	if s.Profiles != nil {
		if profile, err := s.Profiles.Get(ctx, userID); err == nil && profile.Locale != "" {
			out.Locale = profile.Locale
		}
	}
	if s.Users != nil {
		if user, err := s.Users.GetByID(ctx, userID); err == nil {
			for _, role := range user.Roles {
				if role == auth.RoleAdmin {
					out.IsAdmin = true
					break
				}
			}
		}
	}
	if s.Prefs == nil || s.Channels == nil {
		return out, nil
	}
	rows, err := s.Prefs.ListByUser(ctx, userID)
	if err != nil {
		return UserPreferences{}, fmt.Errorf("notifications: list preferences: %w", err)
	}
	for _, row := range rows {
		ch, err := s.Channels.GetByID(ctx, row.ChannelID)
		if err != nil {
			continue
		}
		chID, ok := channelIDForDBType(ch.ChannelType)
		if !ok {
			continue
		}
		ev := EventType(row.EventType)
		if row.Enabled {
			out.Subscribed[ev] = true
			out.EnabledChannels[chID] = true
		} else {
			if _, seen := out.Subscribed[ev]; !seen {
				out.Subscribed[ev] = false
			}
			out.EnabledChannels[chID] = false
		}
		if row.DebounceSeconds > 0 {
			sec := row.DebounceSeconds
			out.DebounceSeconds = &sec
		}
	}
	return out, nil
}

func channelIDForDBType(t db.NotificationChannelType) (ChannelID, bool) {
	switch t {
	case db.NotificationChannelWebhook:
		return ChannelWebhook, true
	case db.NotificationChannelDiscord:
		return ChannelDiscord, true
	case db.NotificationChannelEmail:
		return ChannelSMTP, true
	default:
		return "", false
	}
}
