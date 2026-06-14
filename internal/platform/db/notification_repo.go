package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// NotificationChannelType identifies a delivery backend.
type NotificationChannelType string

const (
	NotificationChannelWebhook NotificationChannelType = "webhook"
	NotificationChannelDiscord NotificationChannelType = "discord"
	NotificationChannelEmail   NotificationChannelType = "email"
)

// NotificationChannel is a configured outbound notification destination.
type NotificationChannel struct {
	ID          int64                   `json:"id"`
	ChannelType NotificationChannelType `json:"channelType"`
	Name        string                  `json:"name"`
	Config      string                  `json:"config"`
	Enabled     bool                    `json:"enabled"`
	CreatedAt   time.Time               `json:"createdAt"`
	UpdatedAt   time.Time               `json:"updatedAt"`
}

// NotificationPreference binds a user/event pair to a channel.
type NotificationPreference struct {
	ID              int64  `json:"id"`
	UserID          string `json:"userId"`
	EventType       string `json:"eventType"`
	ChannelID       int64  `json:"channelId"`
	Enabled         bool   `json:"enabled"`
	DebounceSeconds int    `json:"debounceSeconds"`
}

// NotificationChannelRepo persists notification channels.
type NotificationChannelRepo struct {
	q Querier
}

// NewNotificationChannelRepo returns a repository bound to q.
func NewNotificationChannelRepo(q Querier) *NotificationChannelRepo {
	return &NotificationChannelRepo{q: q}
}

var (
	ErrNotificationChannelNotFound    = errors.New("db notification channel: not found")
	ErrNotificationPreferenceNotFound = errors.New("db notification preference: not found")
)

// Create inserts a notification channel.
func (r *NotificationChannelRepo) Create(ctx context.Context, ch NotificationChannel) (int64, error) {
	if ch.ChannelType == "" {
		return 0, fmt.Errorf("db notification channel create: channel type is required")
	}
	config := ch.Config
	if config == "" {
		config = "{}"
	}
	res, err := r.q.ExecContext(ctx, `
		INSERT INTO notification_channels (channel_type, name, config, enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))
	`, ch.ChannelType, nullStr(ch.Name), config, boolToInt(ch.Enabled))
	if err != nil {
		return 0, fmt.Errorf("db notification channel create: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("db notification channel create id: %w", err)
	}
	return id, nil
}

// GetByID returns a channel by primary key.
func (r *NotificationChannelRepo) GetByID(ctx context.Context, id int64) (NotificationChannel, error) {
	var ch NotificationChannel
	var enabled int
	var created, updated string
	err := r.q.QueryRowContext(ctx, `
		SELECT id, channel_type, name, config, enabled, created_at, updated_at
		FROM notification_channels WHERE id = ?
	`, id).Scan(&ch.ID, &ch.ChannelType, &ch.Name, &ch.Config, &enabled, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return NotificationChannel{}, ErrNotificationChannelNotFound
	}
	if err != nil {
		return NotificationChannel{}, fmt.Errorf("db notification channel get: %w", err)
	}
	ch.Enabled = enabled != 0
	ch.CreatedAt = parseSQLiteTime(created)
	ch.UpdatedAt = parseSQLiteTime(updated)
	return ch, nil
}

// List returns all channels ordered by id.
func (r *NotificationChannelRepo) List(ctx context.Context) ([]NotificationChannel, error) {
	rows, err := r.q.QueryContext(ctx, `
		SELECT id, channel_type, name, config, enabled, created_at, updated_at
		FROM notification_channels ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("db notification channel list: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []NotificationChannel
	for rows.Next() {
		var ch NotificationChannel
		var enabled int
		var created, updated string
		if err := rows.Scan(&ch.ID, &ch.ChannelType, &ch.Name, &ch.Config, &enabled, &created, &updated); err != nil {
			return nil, fmt.Errorf("db notification channel scan: %w", err)
		}
		ch.Enabled = enabled != 0
		ch.CreatedAt = parseSQLiteTime(created)
		ch.UpdatedAt = parseSQLiteTime(updated)
		out = append(out, ch)
	}
	return out, rows.Err()
}

// SetEnabled toggles a channel.
func (r *NotificationChannelRepo) SetEnabled(ctx context.Context, id int64, enabled bool) error {
	res, err := r.q.ExecContext(ctx, `
		UPDATE notification_channels SET enabled = ?, updated_at = datetime('now') WHERE id = ?
	`, boolToInt(enabled), id)
	if err != nil {
		return fmt.Errorf("db notification channel set enabled: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("db notification channel set enabled rows: %w", err)
	}
	if n == 0 {
		return ErrNotificationChannelNotFound
	}
	return nil
}

// NotificationPreferenceRepo persists per-user notification subscriptions.
type NotificationPreferenceRepo struct {
	q Querier
}

// NewNotificationPreferenceRepo returns a repository bound to q.
func NewNotificationPreferenceRepo(q Querier) *NotificationPreferenceRepo {
	return &NotificationPreferenceRepo{q: q}
}

// ListByUser returns preferences for userID.
func (r *NotificationPreferenceRepo) ListByUser(ctx context.Context, userID string) ([]NotificationPreference, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("db notification preference list: user id is required")
	}
	rows, err := r.q.QueryContext(ctx, `
		SELECT id, user_id, event_type, channel_id, enabled, debounce_seconds
		FROM notification_preferences WHERE user_id = ? ORDER BY id
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("db notification preference list: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []NotificationPreference
	for rows.Next() {
		var p NotificationPreference
		var enabled int
		if err := rows.Scan(&p.ID, &p.UserID, &p.EventType, &p.ChannelID, &enabled, &p.DebounceSeconds); err != nil {
			return nil, fmt.Errorf("db notification preference scan: %w", err)
		}
		p.Enabled = enabled != 0
		out = append(out, p)
	}
	return out, rows.Err()
}

// Upsert inserts or updates a preference row.
func (r *NotificationPreferenceRepo) Upsert(ctx context.Context, p NotificationPreference) (int64, error) {
	if strings.TrimSpace(p.UserID) == "" || strings.TrimSpace(p.EventType) == "" || p.ChannelID == 0 {
		return 0, fmt.Errorf("db notification preference upsert: user, event type, and channel id are required")
	}
	if p.ID > 0 {
		res, err := r.q.ExecContext(ctx, `
			UPDATE notification_preferences SET
				event_type = ?,
				channel_id = ?,
				enabled = ?,
				debounce_seconds = ?
			WHERE id = ? AND user_id = ?
		`, p.EventType, p.ChannelID, boolToInt(p.Enabled), p.DebounceSeconds, p.ID, p.UserID)
		if err != nil {
			return 0, fmt.Errorf("db notification preference update: %w", err)
		}
		n, err := res.RowsAffected()
		if err != nil {
			return 0, fmt.Errorf("db notification preference update rows: %w", err)
		}
		if n == 0 {
			return 0, ErrNotificationPreferenceNotFound
		}
		return p.ID, nil
	}

	res, err := r.q.ExecContext(ctx, `
		INSERT INTO notification_preferences (user_id, event_type, channel_id, enabled, debounce_seconds)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(user_id, event_type, channel_id) DO UPDATE SET
			enabled = excluded.enabled,
			debounce_seconds = excluded.debounce_seconds
	`, p.UserID, p.EventType, p.ChannelID, boolToInt(p.Enabled), p.DebounceSeconds)
	if err != nil {
		return 0, fmt.Errorf("db notification preference upsert: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("db notification preference upsert id: %w", err)
	}
	if id == 0 {
		err = r.q.QueryRowContext(ctx, `
			SELECT id FROM notification_preferences
			WHERE user_id = ? AND event_type = ? AND channel_id = ?
		`, p.UserID, p.EventType, p.ChannelID).Scan(&id)
		if err != nil {
			return 0, fmt.Errorf("db notification preference lookup id: %w", err)
		}
	}
	return id, nil
}
