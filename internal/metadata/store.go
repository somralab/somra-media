package metadata

import (
	"context"
	"fmt"

	"github.com/somralab/somra-media/internal/platform/db"
)

// DBStore adapts platform/db repositories to MediaStore.
type DBStore struct {
	DB *db.DB
}

// GetItem loads a media item for matching.
func (s *DBStore) GetItem(ctx context.Context, id int64, locale string) (MediaItemView, error) {
	repo := db.NewMediaRepo(s.DB.Querier())
	item, err := repo.GetItemByID(ctx, id, locale)
	if err != nil {
		return MediaItemView{}, err
	}
	return MediaItemView{
		ID:          item.ID,
		LibraryID:   item.LibraryID,
		Kind:        string(item.Kind),
		ParsedTitle: item.SortTitle,
		ParsedYear:  item.Year,
		Locale:      locale,
	}, nil
}

// ListUnmatched returns unmatched items with parsed titles from files.
func (s *DBStore) ListUnmatched(ctx context.Context, libraryID int64, limit int) ([]MediaItemView, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.DB.SQL().QueryContext(ctx, `
		SELECT DISTINCT mi.id, mi.library_id, mi.kind, COALESCE(mf.parsed_title, mi.sort_title, ''), mf.parsed_year
		FROM media_item mi
		LEFT JOIN media_file mf ON mf.media_item_id = mi.id
		WHERE mi.library_id = ? AND mi.match_status = 'unmatched'
		LIMIT ?
	`, libraryID, limit)
	if err != nil {
		return nil, fmt.Errorf("list unmatched: %w", err)
	}
	defer rows.Close()
	var out []MediaItemView
	for rows.Next() {
		var v MediaItemView
		var year *int
		var y int
		var yNull interface{}
		if err := rows.Scan(&v.ID, &v.LibraryID, &v.Kind, &v.ParsedTitle, &yNull); err != nil {
			return nil, err
		}
		if yNull != nil {
			switch t := yNull.(type) {
			case int64:
				y = int(t)
				year = &y
			case int:
				y = t
				year = &y
			}
		}
		v.ParsedYear = year
		out = append(out, v)
	}
	return out, rows.Err()
}

// ApplyMatch persists provider match results.
func (s *DBStore) ApplyMatch(ctx context.Context, itemID int64, provider, externalID, locale string, detail Detail) error {
	repo := db.NewMediaRepo(s.DB.Querier())
	if err := repo.SetProviderID(ctx, itemID, provider, externalID); err != nil {
		return err
	}
	score := 1.0
	if err := repo.SetMatch(ctx, itemID, db.MatchManual, &score); err != nil {
		return err
	}
	if detail.Title != "" {
		if err := repo.SetLocalizedText(ctx, itemID, locale, "title", detail.Title); err != nil {
			return err
		}
		if err := repo.SetLocalizedText(ctx, itemID, "en-US", "title", detail.Title); err != nil {
			return err
		}
	}
	if detail.Overview != "" {
		if err := repo.SetLocalizedText(ctx, itemID, locale, "overview", detail.Overview); err != nil {
			return err
		}
	}
	if detail.PosterURL != "" {
		if err := repo.UpsertArtwork(ctx, itemID, "poster", detail.PosterURL, ""); err != nil {
			return err
		}
	}
	if detail.BackdropURL != "" {
		if err := repo.UpsertArtwork(ctx, itemID, "backdrop", detail.BackdropURL, ""); err != nil {
			return err
		}
	}
	if detail.Title != "" {
		return repo.IndexFTS(ctx, itemID, detail.Title)
	}
	return nil
}
