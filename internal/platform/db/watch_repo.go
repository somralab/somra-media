package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// WatchState tracks playback progress for a user+item pair.
type WatchState struct {
	UserID      string
	MediaItemID int64
	PositionMs  int64
	Completed   bool
	UpdatedAt   time.Time
}

// WatchRepo persists watch state, favorites, and watchlist entries.
type WatchRepo struct {
	q Querier
}

// NewWatchRepo returns a repository bound to q.
func NewWatchRepo(q Querier) *WatchRepo {
	return &WatchRepo{q: q}
}

// UpsertWatchState inserts or updates playback progress.
func (r *WatchRepo) UpsertWatchState(ctx context.Context, ws WatchState) error {
	_, err := r.q.ExecContext(ctx, `
		INSERT INTO watch_state (user_id, media_item_id, position_ms, completed, updated_at)
		VALUES (?, ?, ?, ?, datetime('now'))
		ON CONFLICT(user_id, media_item_id) DO UPDATE SET
			position_ms = excluded.position_ms,
			completed = excluded.completed,
			updated_at = datetime('now')
	`, ws.UserID, ws.MediaItemID, ws.PositionMs, boolToInt(ws.Completed))
	if err != nil {
		return fmt.Errorf("db watch upsert: %w", err)
	}
	return nil
}

// GetWatchState returns progress for a user+item pair.
func (r *WatchRepo) GetWatchState(ctx context.Context, userID string, mediaItemID int64) (WatchState, error) {
	var ws WatchState
	var completed int
	var updated string
	err := r.q.QueryRowContext(ctx, `
		SELECT user_id, media_item_id, position_ms, completed, updated_at
		FROM watch_state WHERE user_id = ? AND media_item_id = ?
	`, userID, mediaItemID).Scan(&ws.UserID, &ws.MediaItemID, &ws.PositionMs, &completed, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return WatchState{}, ErrNotFound("watch_state")
	}
	if err != nil {
		return WatchState{}, fmt.Errorf("db watch get: %w", err)
	}
	ws.Completed = completed != 0
	ws.UpdatedAt = parseSQLiteTime(updated)
	return ws, nil
}

// ListWatchStates returns all watch states for a user.
func (r *WatchRepo) ListWatchStates(ctx context.Context, userID string) ([]WatchState, error) {
	rows, err := r.q.QueryContext(ctx, `
		SELECT user_id, media_item_id, position_ms, completed, updated_at
		FROM watch_state WHERE user_id = ? ORDER BY updated_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("db watch list: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []WatchState
	for rows.Next() {
		var ws WatchState
		var completed int
		var updated string
		if err := rows.Scan(&ws.UserID, &ws.MediaItemID, &ws.PositionMs, &completed, &updated); err != nil {
			return nil, fmt.Errorf("db watch list scan: %w", err)
		}
		ws.Completed = completed != 0
		ws.UpdatedAt = parseSQLiteTime(updated)
		out = append(out, ws)
	}
	return out, rows.Err()
}

// AddFavorite inserts a favorite if not present.
func (r *WatchRepo) AddFavorite(ctx context.Context, userID string, mediaItemID int64) error {
	_, err := r.q.ExecContext(ctx, `
		INSERT OR IGNORE INTO favorite (user_id, media_item_id, created_at)
		VALUES (?, ?, datetime('now'))
	`, userID, mediaItemID)
	if err != nil {
		return fmt.Errorf("db favorite add: %w", err)
	}
	return nil
}

// RemoveFavorite deletes a favorite.
func (r *WatchRepo) RemoveFavorite(ctx context.Context, userID string, mediaItemID int64) error {
	_, err := r.q.ExecContext(ctx, `
		DELETE FROM favorite WHERE user_id = ? AND media_item_id = ?
	`, userID, mediaItemID)
	if err != nil {
		return fmt.Errorf("db favorite remove: %w", err)
	}
	return nil
}

// ListFavorites returns media item ids favorited by the user.
func (r *WatchRepo) ListFavorites(ctx context.Context, userID string) ([]int64, error) {
	rows, err := r.q.QueryContext(ctx, `
		SELECT media_item_id FROM favorite WHERE user_id = ? ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("db favorite list: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("db favorite list scan: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// AddWatchlist inserts a watchlist entry if not present.
func (r *WatchRepo) AddWatchlist(ctx context.Context, userID string, mediaItemID int64) error {
	_, err := r.q.ExecContext(ctx, `
		INSERT OR IGNORE INTO watchlist (user_id, media_item_id, created_at)
		VALUES (?, ?, datetime('now'))
	`, userID, mediaItemID)
	if err != nil {
		return fmt.Errorf("db watchlist add: %w", err)
	}
	return nil
}

// RemoveWatchlist deletes a watchlist entry.
func (r *WatchRepo) RemoveWatchlist(ctx context.Context, userID string, mediaItemID int64) error {
	_, err := r.q.ExecContext(ctx, `
		DELETE FROM watchlist WHERE user_id = ? AND media_item_id = ?
	`, userID, mediaItemID)
	if err != nil {
		return fmt.Errorf("db watchlist remove: %w", err)
	}
	return nil
}

// ListWatchlist returns media item ids on the user's watchlist.
func (r *WatchRepo) ListWatchlist(ctx context.Context, userID string) ([]int64, error) {
	rows, err := r.q.QueryContext(ctx, `
		SELECT media_item_id FROM watchlist WHERE user_id = ? ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("db watchlist list: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("db watchlist list scan: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// ErrNotFound is a generic not-found marker for watch repo lookups.
type notFoundErr string

func (e notFoundErr) Error() string { return string(e) }

// ErrNotFound returns a typed not-found error.
func ErrNotFound(entity string) error {
	return notFoundErr("db " + entity + ": not found")
}
