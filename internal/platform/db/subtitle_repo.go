package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// SubtitleSource describes how a subtitle was obtained.
type SubtitleSource string

const (
	SubtitleEmbedded SubtitleSource = "embedded"
	SubtitleExternal SubtitleSource = "external"
	SubtitleUploaded SubtitleSource = "uploaded"
)

// SubtitleFile is a managed subtitle track for a media item.
type SubtitleFile struct {
	ID          int64          `json:"id"`
	MediaItemID int64          `json:"mediaItemId"`
	Language    string         `json:"language"`
	Source      SubtitleSource `json:"source"`
	Path        string         `json:"path,omitempty"`
	Provider    string         `json:"provider,omitempty"`
	ExternalID  string         `json:"-"`
	CreatedAt   time.Time      `json:"createdAt,omitempty"`
}

// SubtitleRepo persists external/uploaded subtitle files.
type SubtitleRepo struct {
	q Querier
}

// NewSubtitleRepo returns a repository bound to q.
func NewSubtitleRepo(q Querier) *SubtitleRepo {
	return &SubtitleRepo{q: q}
}

// ListByMediaItem returns subtitle files for a media item.
func (r *SubtitleRepo) ListByMediaItem(ctx context.Context, mediaItemID int64) ([]SubtitleFile, error) {
	rows, err := r.q.QueryContext(ctx, `
		SELECT id, media_item_id, language, source, path, provider, external_id, created_at
		FROM subtitle_file WHERE media_item_id = ? ORDER BY id
	`, mediaItemID)
	if err != nil {
		return nil, fmt.Errorf("db subtitle list: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var out []SubtitleFile
	for rows.Next() {
		var f SubtitleFile
		var provider, externalID sql.NullString
		var createdAt sql.NullString
		if err := rows.Scan(&f.ID, &f.MediaItemID, &f.Language, &f.Source, &f.Path, &provider, &externalID, &createdAt); err != nil {
			return nil, fmt.Errorf("db subtitle scan: %w", err)
		}
		if provider.Valid {
			f.Provider = provider.String
		}
		if externalID.Valid {
			f.ExternalID = externalID.String
		}
		if createdAt.Valid {
			if t, err := time.Parse("2006-01-02 15:04:05", createdAt.String); err == nil {
				f.CreatedAt = t
			}
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

// Create inserts a subtitle file row.
func (r *SubtitleRepo) Create(ctx context.Context, f SubtitleFile) (int64, error) {
	res, err := r.q.ExecContext(ctx, `
		INSERT INTO subtitle_file (media_item_id, language, source, path, provider, external_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, datetime('now'))
	`, f.MediaItemID, f.Language, f.Source, f.Path, nullStr(f.Provider), nullStr(f.ExternalID))
	if err != nil {
		return 0, fmt.Errorf("db subtitle create: %w", err)
	}
	return res.LastInsertId()
}

// HasLanguage reports whether a subtitle exists for the given language.
func (r *SubtitleRepo) HasLanguage(ctx context.Context, mediaItemID int64, language string) (bool, error) {
	var n int
	err := r.q.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM subtitle_file WHERE media_item_id = ? AND LOWER(language) = LOWER(?)
	`, mediaItemID, language).Scan(&n)
	if err != nil {
		return false, fmt.Errorf("db subtitle has language: %w", err)
	}
	return n > 0, nil
}

// ListMediaItemsMissingLanguages returns item IDs missing any preferred language.
func (r *SubtitleRepo) ListMediaItemsMissingLanguages(ctx context.Context, langs []string, limit int) ([]int64, error) {
	if len(langs) == 0 {
		return nil, nil
	}
	if limit <= 0 {
		limit = 100
	}
	// Items with at least one media file but missing subtitle for any preferred lang.
	query := `
		SELECT DISTINCT mi.id
		FROM media_item mi
		JOIN media_file mf ON mf.media_item_id = mi.id
		WHERE NOT EXISTS (
			SELECT 1 FROM subtitle_file sf
			WHERE sf.media_item_id = mi.id AND LOWER(sf.language) = LOWER(?)
		)
		LIMIT ?
	`
	var ids []int64
	seen := make(map[int64]struct{})
	for _, lang := range langs {
		rows, err := r.q.QueryContext(ctx, query, lang, limit)
		if err != nil {
			return nil, fmt.Errorf("db subtitle missing list: %w", err)
		}
		for rows.Next() {
			var id int64
			if err := rows.Scan(&id); err != nil {
				_ = rows.Close()
				return nil, err
			}
			if _, ok := seen[id]; !ok {
				seen[id] = struct{}{}
				ids = append(ids, id)
			}
		}
		_ = rows.Close()
		if err := rows.Err(); err != nil {
			return nil, err
		}
	}
	return ids, nil
}

var ErrSubtitleNotFound = errors.New("db subtitle: not found")
