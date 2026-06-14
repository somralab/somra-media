package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// PlaybackMode describes how media is delivered to the client.
type PlaybackMode string

const (
	PlaybackDirectPlay   PlaybackMode = "direct_play"
	PlaybackDirectStream PlaybackMode = "direct_stream"
	PlaybackTranscode    PlaybackMode = "transcode"
)

// PlaybackStatus tracks session lifecycle.
type PlaybackStatus string

const (
	PlaybackPending PlaybackStatus = "pending"
	PlaybackQueued  PlaybackStatus = "queued"
	PlaybackActive  PlaybackStatus = "active"
	PlaybackStopped PlaybackStatus = "stopped"
	PlaybackFailed  PlaybackStatus = "failed"
	PlaybackExpired PlaybackStatus = "expired"
)

// PlaybackSession is a persisted streaming/transcode session.
type PlaybackSession struct {
	ID                  string
	UserID              string
	MediaItemID         int64
	MediaFileID         int64
	Mode                PlaybackMode
	Status              PlaybackStatus
	CachePath           string
	StartPositionMs     int64
	AudioStreamIndex    *int
	SubtitleStreamIndex *int
	ExpiresAt           time.Time
	LastAccessAt        time.Time
	CreatedAt           time.Time
	ErrorMessage        string
}

// PlaybackRepo persists playback sessions.
type PlaybackRepo struct {
	q Querier
}

// NewPlaybackRepo returns a repository bound to q.
func NewPlaybackRepo(q Querier) *PlaybackRepo {
	return &PlaybackRepo{q: q}
}

var ErrPlaybackSessionNotFound = errors.New("db playback session: not found")

// Create inserts a new playback session row.
func (r *PlaybackRepo) Create(ctx context.Context, s PlaybackSession) error {
	if s.ID == "" || s.UserID == "" {
		return fmt.Errorf("db playback create: id and user_id are required")
	}
	_, err := r.q.ExecContext(ctx, `
		INSERT INTO playback_session (
			id, user_id, media_item_id, media_file_id, mode, status, cache_path,
			start_position_ms, audio_stream_index, subtitle_stream_index,
			expires_at, last_access_at, created_at, error_message
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'), ?)
	`, s.ID, s.UserID, s.MediaItemID, s.MediaFileID, s.Mode, s.Status,
		nullStr(s.CachePath), s.StartPositionMs,
		nullInt(s.AudioStreamIndex), nullInt(s.SubtitleStreamIndex),
		s.ExpiresAt.UTC().Format("2006-01-02 15:04:05"),
		nullStr(s.ErrorMessage))
	if err != nil {
		return fmt.Errorf("db playback create %q: %w", s.ID, err)
	}
	return nil
}

// GetByID returns a session by id.
func (r *PlaybackRepo) GetByID(ctx context.Context, id string) (PlaybackSession, error) {
	return r.scanOne(r.q.QueryRowContext(ctx, `
		SELECT id, user_id, media_item_id, media_file_id, mode, status, cache_path,
		       start_position_ms, audio_stream_index, subtitle_stream_index,
		       expires_at, last_access_at, created_at, error_message
		FROM playback_session WHERE id = ?
	`, id))
}

// GetByIDForUser returns a session owned by userID.
func (r *PlaybackRepo) GetByIDForUser(ctx context.Context, id, userID string) (PlaybackSession, error) {
	s, err := r.GetByID(ctx, id)
	if err != nil {
		return PlaybackSession{}, err
	}
	if s.UserID != userID {
		return PlaybackSession{}, ErrPlaybackSessionNotFound
	}
	return s, nil
}

// ListActiveByUser returns non-terminal sessions for a user.
func (r *PlaybackRepo) ListActiveByUser(ctx context.Context, userID string) ([]PlaybackSession, error) {
	rows, err := r.q.QueryContext(ctx, `
		SELECT id, user_id, media_item_id, media_file_id, mode, status, cache_path,
		       start_position_ms, audio_stream_index, subtitle_stream_index,
		       expires_at, last_access_at, created_at, error_message
		FROM playback_session
		WHERE user_id = ? AND status IN ('pending', 'queued', 'active')
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("db playback list active: %w", err)
	}
	defer func() { _ = rows.Close() }()
	return r.scanRows(rows)
}

// CountActiveTranscodes returns active+queued transcode sessions globally.
func (r *PlaybackRepo) CountActiveTranscodes(ctx context.Context) (int, error) {
	var n int
	err := r.q.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM playback_session
		WHERE mode = 'transcode' AND status IN ('pending', 'queued', 'active')
	`).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("db playback count transcodes: %w", err)
	}
	return n, nil
}

// UpdateStatus sets status and optional error message.
func (r *PlaybackRepo) UpdateStatus(ctx context.Context, id string, status PlaybackStatus, errMsg string) error {
	res, err := r.q.ExecContext(ctx, `
		UPDATE playback_session SET status = ?, error_message = ?, last_access_at = datetime('now')
		WHERE id = ?
	`, status, nullStr(errMsg), id)
	if err != nil {
		return fmt.Errorf("db playback update status %q: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("db playback update status rows: %w", err)
	}
	if n == 0 {
		return ErrPlaybackSessionNotFound
	}
	return nil
}

// TouchLastAccess updates last_access_at for idle reaping.
func (r *PlaybackRepo) TouchLastAccess(ctx context.Context, id string) error {
	_, err := r.q.ExecContext(ctx, `
		UPDATE playback_session SET last_access_at = datetime('now') WHERE id = ?
	`, id)
	if err != nil {
		return fmt.Errorf("db playback touch %q: %w", id, err)
	}
	return nil
}

// Stop marks a session stopped.
func (r *PlaybackRepo) Stop(ctx context.Context, id string) error {
	return r.UpdateStatus(ctx, id, PlaybackStopped, "")
}

// ListIdleSessions returns sessions idle longer than cutoff.
func (r *PlaybackRepo) ListIdleSessions(ctx context.Context, cutoff time.Time) ([]PlaybackSession, error) {
	rows, err := r.q.QueryContext(ctx, `
		SELECT id, user_id, media_item_id, media_file_id, mode, status, cache_path,
		       start_position_ms, audio_stream_index, subtitle_stream_index,
		       expires_at, last_access_at, created_at, error_message
		FROM playback_session
		WHERE status IN ('pending', 'queued', 'active')
		  AND last_access_at < ?
	`, cutoff.UTC().Format("2006-01-02 15:04:05"))
	if err != nil {
		return nil, fmt.Errorf("db playback list idle: %w", err)
	}
	defer func() { _ = rows.Close() }()
	return r.scanRows(rows)
}

// ListExpired returns sessions past expires_at still marked active.
func (r *PlaybackRepo) ListExpired(ctx context.Context, now time.Time) ([]PlaybackSession, error) {
	rows, err := r.q.QueryContext(ctx, `
		SELECT id, user_id, media_item_id, media_file_id, mode, status, cache_path,
		       start_position_ms, audio_stream_index, subtitle_stream_index,
		       expires_at, last_access_at, created_at, error_message
		FROM playback_session
		WHERE status IN ('pending', 'queued', 'active')
		  AND expires_at < ?
	`, now.UTC().Format("2006-01-02 15:04:05"))
	if err != nil {
		return nil, fmt.Errorf("db playback list expired: %w", err)
	}
	defer func() { _ = rows.Close() }()
	return r.scanRows(rows)
}

// MarkExpired bulk-updates expired sessions.
func (r *PlaybackRepo) MarkExpired(ctx context.Context, ids []string) error {
	for _, id := range ids {
		if err := r.UpdateStatus(ctx, id, PlaybackExpired, ""); err != nil && !errors.Is(err, ErrPlaybackSessionNotFound) {
			return err
		}
	}
	return nil
}

func (r *PlaybackRepo) scanOne(row *sql.Row) (PlaybackSession, error) {
	var s PlaybackSession
	var cachePath, errMsg sql.NullString
	var audioIdx, subIdx sql.NullInt64
	var expires, lastAccess, created string
	err := row.Scan(
		&s.ID, &s.UserID, &s.MediaItemID, &s.MediaFileID,
		&s.Mode, &s.Status, &cachePath,
		&s.StartPositionMs, &audioIdx, &subIdx,
		&expires, &lastAccess, &created, &errMsg,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return PlaybackSession{}, ErrPlaybackSessionNotFound
	}
	if err != nil {
		return PlaybackSession{}, fmt.Errorf("db playback scan: %w", err)
	}
	s.CachePath = cachePath.String
	s.AudioStreamIndex = nullIntPtr(audioIdx)
	s.SubtitleStreamIndex = nullIntPtr(subIdx)
	s.ExpiresAt = parseSQLiteTime(expires)
	s.LastAccessAt = parseSQLiteTime(lastAccess)
	s.CreatedAt = parseSQLiteTime(created)
	s.ErrorMessage = errMsg.String
	return s, nil
}

func (r *PlaybackRepo) scanRows(rows *sql.Rows) ([]PlaybackSession, error) {
	var out []PlaybackSession
	for rows.Next() {
		var s PlaybackSession
		var cachePath, errMsg sql.NullString
		var audioIdx, subIdx sql.NullInt64
		var expires, lastAccess, created string
		if err := rows.Scan(
			&s.ID, &s.UserID, &s.MediaItemID, &s.MediaFileID,
			&s.Mode, &s.Status, &cachePath,
			&s.StartPositionMs, &audioIdx, &subIdx,
			&expires, &lastAccess, &created, &errMsg,
		); err != nil {
			return nil, fmt.Errorf("db playback scan row: %w", err)
		}
		s.CachePath = cachePath.String
		s.AudioStreamIndex = nullIntPtr(audioIdx)
		s.SubtitleStreamIndex = nullIntPtr(subIdx)
		s.ExpiresAt = parseSQLiteTime(expires)
		s.LastAccessAt = parseSQLiteTime(lastAccess)
		s.CreatedAt = parseSQLiteTime(created)
		s.ErrorMessage = errMsg.String
		out = append(out, s)
	}
	return out, rows.Err()
}
