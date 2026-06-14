package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// MatchStatus describes metadata matching state.
type MatchStatus string

const (
	MatchUnmatched MatchStatus = "unmatched"
	MatchMatched   MatchStatus = "matched"
	MatchManual    MatchStatus = "manual"
)

// MediaItem is a logical work (movie, series, album) within a library.
type MediaItem struct {
	ID          int64       `json:"id"`
	LibraryID   int64       `json:"libraryId"`
	Kind        LibraryKind `json:"kind"`
	SortTitle   string      `json:"sortTitle,omitempty"`
	Year        *int        `json:"year,omitempty"`
	MatchStatus MatchStatus `json:"matchStatus"`
	MatchScore  *float64    `json:"matchScore,omitempty"`
	Title       string      `json:"title"`
	Overview    string      `json:"overview,omitempty"`
	PosterURL   string      `json:"posterUrl,omitempty"`
	CreatedAt   time.Time   `json:"createdAt,omitempty"`
	UpdatedAt   time.Time   `json:"updatedAt,omitempty"`
}

// MediaFile is an on-disk file tracked by a scan.
type MediaFile struct {
	ID            int64
	LibraryID     int64
	MediaItemID   *int64
	EpisodeID     *int64
	Path          string
	FileName      string
	SizeBytes     int64
	MtimeNs       int64
	ContentHash   string
	ParsedTitle   string
	ParsedYear    *int
	ParsedSeason  *int
	ParsedEpisode *int
}

// MediaRepo handles media items, files, and localized text.
type MediaRepo struct {
	q Querier
}

// NewMediaRepo returns a repository bound to q.
func NewMediaRepo(q Querier) *MediaRepo {
	return &MediaRepo{q: q}
}

var ErrMediaItemNotFound = errors.New("db media item: not found")

// CreateItem inserts a media item row.
func (r *MediaRepo) CreateItem(ctx context.Context, libraryID int64, kind LibraryKind, sortTitle string, year *int) (int64, error) {
	res, err := r.q.ExecContext(ctx, `
		INSERT INTO media_item (library_id, kind, sort_title, year, created_at, updated_at)
		VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))
	`, libraryID, kind, nullStr(sortTitle), nullInt(year))
	if err != nil {
		return 0, fmt.Errorf("db media create item: %w", err)
	}
	return res.LastInsertId()
}

// UpsertFile inserts or updates a media file by path.
func (r *MediaRepo) UpsertFile(ctx context.Context, f MediaFile) (int64, error) {
	const q = `
		INSERT INTO media_file (
			library_id, media_item_id, episode_id, path, file_name,
			size_bytes, mtime_ns, content_hash,
			parsed_title, parsed_year, parsed_season, parsed_episode,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))
		ON CONFLICT(path) DO UPDATE SET
			library_id = excluded.library_id,
			media_item_id = excluded.media_item_id,
			episode_id = excluded.episode_id,
			file_name = excluded.file_name,
			size_bytes = excluded.size_bytes,
			mtime_ns = excluded.mtime_ns,
			content_hash = excluded.content_hash,
			parsed_title = excluded.parsed_title,
			parsed_year = excluded.parsed_year,
			parsed_season = excluded.parsed_season,
			parsed_episode = excluded.parsed_episode,
			updated_at = datetime('now')
	`
	res, err := r.q.ExecContext(ctx, q,
		f.LibraryID, nullInt64(f.MediaItemID), nullInt64(f.EpisodeID),
		f.Path, f.FileName, f.SizeBytes, f.MtimeNs, nullStr(f.ContentHash),
		nullStr(f.ParsedTitle), nullInt(f.ParsedYear), nullInt(f.ParsedSeason), nullInt(f.ParsedEpisode),
	)
	if err != nil {
		return 0, fmt.Errorf("db media upsert file %q: %w", f.Path, err)
	}
	id, err := res.LastInsertId()
	if err != nil || id == 0 {
		err = r.q.QueryRowContext(ctx, `SELECT id FROM media_file WHERE path = ?`, f.Path).Scan(&id)
		if err != nil {
			return 0, fmt.Errorf("db media upsert file id %q: %w", f.Path, err)
		}
	}
	return id, nil
}

// GetFileByPath returns a file row or sql.ErrNoRows.
func (r *MediaRepo) GetFileByPath(ctx context.Context, path string) (MediaFile, error) {
	var f MediaFile
	var itemID, epID sql.NullInt64
	var hash, title sql.NullString
	var year, season, episode sql.NullInt64
	err := r.q.QueryRowContext(ctx, `
		SELECT id, library_id, media_item_id, episode_id, path, file_name,
		       size_bytes, mtime_ns, content_hash,
		       parsed_title, parsed_year, parsed_season, parsed_episode
		FROM media_file WHERE path = ?
	`, path).Scan(
		&f.ID, &f.LibraryID, &itemID, &epID, &f.Path, &f.FileName,
		&f.SizeBytes, &f.MtimeNs, &hash,
		&title, &year, &season, &episode,
	)
	if err != nil {
		return MediaFile{}, fmt.Errorf("db media get file %q: %w", path, err)
	}
	f.MediaItemID = nullInt64Ptr(itemID)
	f.EpisodeID = nullInt64Ptr(epID)
	f.ContentHash = hash.String
	f.ParsedTitle = title.String
	f.ParsedYear = nullIntPtr(year)
	f.ParsedSeason = nullIntPtr(season)
	f.ParsedEpisode = nullIntPtr(episode)
	return f, nil
}

// ListItemsByLibrary returns media items with localized title for locale.
func (r *MediaRepo) ListItemsByLibrary(ctx context.Context, libraryID int64, locale string, limit, offset int) ([]MediaItem, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.q.QueryContext(ctx, `
		SELECT mi.id, mi.library_id, mi.kind, mi.sort_title, mi.year,
		       mi.match_status, mi.match_score, mi.created_at, mi.updated_at,
		       COALESCE(lt.value, mi.sort_title, '') AS title,
		       COALESCE(lo.value, '') AS overview,
		       COALESCE(a.source_url, a.local_path, '') AS poster
		FROM media_item mi
		LEFT JOIN localized_text lt ON lt.media_item_id = mi.id AND lt.locale = ? AND lt.field = 'title'
		LEFT JOIN localized_text lo ON lo.media_item_id = mi.id AND lo.locale = ? AND lo.field = 'overview'
		LEFT JOIN artwork a ON a.media_item_id = mi.id AND a.kind = 'poster'
		WHERE mi.library_id = ?
		ORDER BY mi.sort_title COLLATE NOCASE
		LIMIT ? OFFSET ?
	`, locale, locale, libraryID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("db media list items %d: %w", libraryID, err)
	}
	defer func() { _ = rows.Close() }()
	return scanMediaItems(rows)
}

// GetItemByID returns one media item.
func (r *MediaRepo) GetItemByID(ctx context.Context, id int64, locale string) (MediaItem, error) {
	rows, err := r.q.QueryContext(ctx, `
		SELECT mi.id, mi.library_id, mi.kind, mi.sort_title, mi.year,
		       mi.match_status, mi.match_score, mi.created_at, mi.updated_at,
		       COALESCE(lt.value, mi.sort_title, '') AS title,
		       COALESCE(lo.value, '') AS overview,
		       COALESCE(a.source_url, a.local_path, '') AS poster
		FROM media_item mi
		LEFT JOIN localized_text lt ON lt.media_item_id = mi.id AND lt.locale = ? AND lt.field = 'title'
		LEFT JOIN localized_text lo ON lo.media_item_id = mi.id AND lo.locale = ? AND lo.field = 'overview'
		LEFT JOIN artwork a ON a.media_item_id = mi.id AND a.kind = 'poster'
		WHERE mi.id = ?
	`, locale, locale, id)
	if err != nil {
		return MediaItem{}, fmt.Errorf("db media get item %d: %w", id, err)
	}
	defer func() { _ = rows.Close() }()
	items, err := scanMediaItems(rows)
	if err != nil {
		return MediaItem{}, err
	}
	if len(items) == 0 {
		return MediaItem{}, fmt.Errorf("db media get item %d: %w", id, ErrMediaItemNotFound)
	}
	return items[0], nil
}

// SetLocalizedText upserts a localized field.
func (r *MediaRepo) SetLocalizedText(ctx context.Context, itemID int64, locale, field, value string) error {
	const q = `
		INSERT INTO localized_text (media_item_id, locale, field, value)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(media_item_id, locale, field) DO UPDATE SET value = excluded.value
	`
	if _, err := r.q.ExecContext(ctx, q, itemID, locale, field, value); err != nil {
		return fmt.Errorf("db media set localized %d/%s/%s: %w", itemID, locale, field, err)
	}
	return nil
}

// SetMatch updates match status and score.
func (r *MediaRepo) SetMatch(ctx context.Context, itemID int64, status MatchStatus, score *float64) error {
	_, err := r.q.ExecContext(ctx, `
		UPDATE media_item SET match_status = ?, match_score = ?, updated_at = datetime('now')
		WHERE id = ?
	`, status, nullFloat(score), itemID)
	if err != nil {
		return fmt.Errorf("db media set match %d: %w", itemID, err)
	}
	return nil
}

// SetProviderID upserts an external provider identifier.
func (r *MediaRepo) SetProviderID(ctx context.Context, itemID int64, provider, externalID string) error {
	const q = `
		INSERT INTO provider_id (media_item_id, provider, external_id)
		VALUES (?, ?, ?)
		ON CONFLICT(media_item_id, provider) DO UPDATE SET external_id = excluded.external_id
	`
	if _, err := r.q.ExecContext(ctx, q, itemID, provider, externalID); err != nil {
		return fmt.Errorf("db media set provider %d/%s: %w", itemID, provider, err)
	}
	return nil
}

// UpsertArtwork stores artwork reference.
func (r *MediaRepo) UpsertArtwork(ctx context.Context, itemID int64, kind, sourceURL, localPath string) error {
	_, err := r.q.ExecContext(ctx, `
		DELETE FROM artwork WHERE media_item_id = ? AND kind = ?
	`, itemID, kind)
	if err != nil {
		return fmt.Errorf("db media delete artwork: %w", err)
	}
	_, err = r.q.ExecContext(ctx, `
		INSERT INTO artwork (media_item_id, kind, source_url, local_path)
		VALUES (?, ?, ?, ?)
	`, itemID, kind, nullStr(sourceURL), nullStr(localPath))
	if err != nil {
		return fmt.Errorf("db media upsert artwork: %w", err)
	}
	return nil
}

// UpsertTechnical stores ffprobe output for a file.
func (r *MediaRepo) UpsertTechnical(ctx context.Context, fileID int64, durationMs int64, container, videoCodec string, width, height int, audioCodec string, channels, subtitleCount int, rawJSON string) error {
	_, err := r.q.ExecContext(ctx, `
		INSERT INTO media_technical (
			media_file_id, duration_ms, container, video_codec, video_width, video_height,
			audio_codec, audio_channels, subtitle_count, raw_json, probed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'))
		ON CONFLICT(media_file_id) DO UPDATE SET
			duration_ms = excluded.duration_ms,
			container = excluded.container,
			video_codec = excluded.video_codec,
			video_width = excluded.video_width,
			video_height = excluded.video_height,
			audio_codec = excluded.audio_codec,
			audio_channels = excluded.audio_channels,
			subtitle_count = excluded.subtitle_count,
			raw_json = excluded.raw_json,
			probed_at = datetime('now')
	`, fileID, durationMs, nullStr(container), nullStr(videoCodec), width, height,
		nullStr(audioCodec), channels, subtitleCount, nullStr(rawJSON))
	if err != nil {
		return fmt.Errorf("db media upsert technical %d: %w", fileID, err)
	}
	return nil
}

// SearchTitleFTS performs a basic FTS5 title search.
func (r *MediaRepo) SearchTitleFTS(ctx context.Context, query string, limit int) ([]int64, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = 20
	}
	rows, err := r.q.QueryContext(ctx, `
		SELECT rowid FROM media_item_fts WHERE media_item_fts MATCH ? LIMIT ?
	`, query, limit)
	if err != nil {
		return nil, fmt.Errorf("db media fts search: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("db media fts scan: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// IndexFTS upserts a title into the FTS index.
func (r *MediaRepo) IndexFTS(ctx context.Context, itemID int64, title string) error {
	_, _ = r.q.ExecContext(ctx, `DELETE FROM media_item_fts WHERE rowid = ?`, itemID)
	_, err := r.q.ExecContext(ctx, `
		INSERT INTO media_item_fts(rowid, title) VALUES (?, ?)
	`, itemID, title)
	if err != nil {
		return fmt.Errorf("db media fts index %d: %w", itemID, err)
	}
	return nil
}

func scanMediaItems(rows *sql.Rows) ([]MediaItem, error) {
	var out []MediaItem
	for rows.Next() {
		var mi MediaItem
		var sortTitle, matchStatus sql.NullString
		var year sql.NullInt64
		var score sql.NullFloat64
		var created, updated string
		if err := rows.Scan(
			&mi.ID, &mi.LibraryID, &mi.Kind, &sortTitle, &year,
			&matchStatus, &score, &created, &updated,
			&mi.Title, &mi.Overview, &mi.PosterURL,
		); err != nil {
			return nil, fmt.Errorf("db media scan item: %w", err)
		}
		mi.SortTitle = sortTitle.String
		mi.Year = nullIntPtr(year)
		mi.MatchStatus = MatchStatus(matchStatus.String)
		mi.MatchScore = nullFloatPtr(score)
		mi.CreatedAt, _ = time.Parse(time.RFC3339, created)
		mi.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
		out = append(out, mi)
	}
	return out, rows.Err()
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullInt(v *int) any {
	if v == nil {
		return nil
	}
	return *v
}

func nullInt64(v *int64) any {
	if v == nil {
		return nil
	}
	return *v
}

func nullFloat(v *float64) any {
	if v == nil {
		return nil
	}
	return *v
}

func nullInt64Ptr(v sql.NullInt64) *int64 {
	if !v.Valid {
		return nil
	}
	n := v.Int64
	return &n
}

func nullIntPtr(v sql.NullInt64) *int {
	if !v.Valid {
		return nil
	}
	n := int(v.Int64)
	return &n
}

func nullFloatPtr(v sql.NullFloat64) *float64 {
	if !v.Valid {
		return nil
	}
	f := v.Float64
	return &f
}
