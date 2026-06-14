package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// BrowseFilters controls paginated library listing.
type BrowseFilters struct {
	Offset      int
	Limit       int
	Sort        string // title, year, created_at
	Genre       string
	Year        *int
	WatchStatus string // unwatched, in_progress, completed, all
	UserID      string
}

// PaginatedMediaItems is a page of media items with total count.
type PaginatedMediaItems struct {
	Items  []MediaItem `json:"items"`
	Total  int         `json:"total"`
	Offset int         `json:"offset"`
	Limit  int         `json:"limit"`
}

// CastMember is a person credited on a media item.
type CastMember struct {
	Name     string `json:"name"`
	Role     string `json:"role"`
	Order    int    `json:"order"`
	ImageURL string `json:"imageUrl,omitempty"`
}

// ArtworkImage is a poster/backdrop reference.
type ArtworkImage struct {
	Kind      string `json:"kind"`
	SourceURL string `json:"sourceUrl,omitempty"`
	LocalPath string `json:"localPath,omitempty"`
	Width     *int   `json:"width,omitempty"`
	Height    *int   `json:"height,omitempty"`
}

// EpisodeSummary is a lightweight episode row for detail views.
type EpisodeSummary struct {
	ID            int64  `json:"id"`
	SeasonNumber  int    `json:"seasonNumber"`
	EpisodeNumber int    `json:"episodeNumber"`
	Title         string `json:"title,omitempty"`
}

// SeasonDetail groups episodes under a season number.
type SeasonDetail struct {
	SeasonNumber int              `json:"seasonNumber"`
	Episodes     []EpisodeSummary `json:"episodes"`
}

// MediaDetail is the full detail payload for a media item.
type MediaDetail struct {
	MediaItem
	Genres      []string         `json:"genres"`
	BackdropURL string           `json:"backdropUrl,omitempty"`
	Cast        []CastMember     `json:"cast"`
	Images      []ArtworkImage   `json:"images"`
	Seasons     []SeasonDetail   `json:"seasons,omitempty"`
	IsFavorite  bool             `json:"isFavorite"`
	InWatchlist bool             `json:"inWatchlist"`
	WatchState  *WatchStateBrief `json:"watchState,omitempty"`
}

// WatchStateBrief is playback progress surfaced on cards and detail.
type WatchStateBrief struct {
	PositionMs int64 `json:"positionMs"`
	Completed  bool  `json:"completed"`
}

// MediaItemSummary adds optional watch progress for shelf cards.
type MediaItemSummary struct {
	MediaItem
	WatchState *WatchStateBrief `json:"watchState,omitempty"`
}

// DiscoverShelf is a rule-based home row.
type DiscoverShelf struct {
	ID       string             `json:"id"`
	TitleKey string             `json:"titleKey"`
	Items    []MediaItemSummary `json:"items"`
}

// DiscoverHome aggregates discover shelves for the authenticated user.
type DiscoverHome struct {
	Shelves []DiscoverShelf `json:"shelves"`
}

// SearchResult is a single FTS hit.
type SearchResult struct {
	MediaItem
	Score float64 `json:"score,omitempty"`
}

// BrowseRepo provides paginated browse, discover, search, and detail queries.
type BrowseRepo struct {
	q Querier
}

// NewBrowseRepo returns a repository bound to q.
func NewBrowseRepo(q Querier) *BrowseRepo {
	return &BrowseRepo{q: q}
}

// ListPaginated returns a filtered, sorted page of media items in a library.
func (r *BrowseRepo) ListPaginated(ctx context.Context, libraryID int64, locale string, f BrowseFilters) (PaginatedMediaItems, error) {
	if f.Limit <= 0 {
		f.Limit = 50
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	if f.WatchStatus == "" {
		f.WatchStatus = "all"
	}

	orderBy := "mi.sort_title COLLATE NOCASE"
	switch f.Sort {
	case "year":
		orderBy = "mi.year DESC, mi.sort_title COLLATE NOCASE"
	case "created_at":
		orderBy = "mi.created_at DESC"
	case "title":
		orderBy = "mi.sort_title COLLATE NOCASE"
	}

	var args []any
	where := strings.Builder{}
	where.WriteString("mi.library_id = ?")
	args = append(args, libraryID)

	if f.Genre != "" {
		where.WriteString(" AND EXISTS (SELECT 1 FROM media_genre mg JOIN genre g ON g.id = mg.genre_id WHERE mg.media_item_id = mi.id AND g.name = ?)")
		args = append(args, f.Genre)
	}
	if f.Year != nil {
		where.WriteString(" AND mi.year = ?")
		args = append(args, *f.Year)
	}
	if f.UserID != "" {
		switch f.WatchStatus {
		case "unwatched":
			where.WriteString(" AND NOT EXISTS (SELECT 1 FROM watch_state ws WHERE ws.media_item_id = mi.id AND ws.user_id = ?)")
			args = append(args, f.UserID)
		case "in_progress":
			where.WriteString(" AND EXISTS (SELECT 1 FROM watch_state ws WHERE ws.media_item_id = mi.id AND ws.user_id = ? AND ws.position_ms > 0 AND ws.completed = 0)")
			args = append(args, f.UserID)
		case "completed":
			where.WriteString(" AND EXISTS (SELECT 1 FROM watch_state ws WHERE ws.media_item_id = mi.id AND ws.user_id = ? AND ws.completed = 1)")
			args = append(args, f.UserID)
		}
	}

	countQ := fmt.Sprintf(`SELECT COUNT(*) FROM media_item mi WHERE %s`, where.String())
	var total int
	if err := r.q.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return PaginatedMediaItems{}, fmt.Errorf("db browse count: %w", err)
	}

	listQ := fmt.Sprintf(`
		SELECT mi.id, mi.library_id, mi.kind, mi.sort_title, mi.year,
		       mi.match_status, mi.match_score, mi.created_at, mi.updated_at,
		       COALESCE(lt.value, mi.sort_title, '') AS title,
		       COALESCE(lo.value, '') AS overview,
		       COALESCE(a.source_url, a.local_path, '') AS poster,
		       mi.content_rating
		FROM media_item mi
		LEFT JOIN localized_text lt ON lt.media_item_id = mi.id AND lt.locale = ? AND lt.field = 'title'
		LEFT JOIN localized_text lo ON lo.media_item_id = mi.id AND lo.locale = ? AND lo.field = 'overview'
		LEFT JOIN artwork a ON a.media_item_id = mi.id AND a.kind = 'poster'
		WHERE %s
		ORDER BY %s
		LIMIT ? OFFSET ?
	`, where.String(), orderBy)

	listArgs := []any{locale, locale}
	listArgs = append(listArgs, args...)
	listArgs = append(listArgs, f.Limit, f.Offset)

	rows, err := r.q.QueryContext(ctx, listQ, listArgs...)
	if err != nil {
		return PaginatedMediaItems{}, fmt.Errorf("db browse list: %w", err)
	}
	defer func() { _ = rows.Close() }()

	items, err := scanMediaItems(rows)
	if err != nil {
		return PaginatedMediaItems{}, err
	}
	return PaginatedMediaItems{Items: items, Total: total, Offset: f.Offset, Limit: f.Limit}, nil
}

// GetDetail loads full metadata for one media item.
func (r *BrowseRepo) GetDetail(ctx context.Context, itemID int64, locale, userID string) (MediaDetail, error) {
	item, err := NewMediaRepo(r.q).GetItemByID(ctx, itemID, locale)
	if err != nil {
		return MediaDetail{}, err
	}

	detail := MediaDetail{MediaItem: item}

	genres, err := r.listGenres(ctx, itemID)
	if err != nil {
		return MediaDetail{}, err
	}
	detail.Genres = genres

	cast, err := r.listCast(ctx, itemID)
	if err != nil {
		return MediaDetail{}, err
	}
	detail.Cast = cast

	images, err := r.listArtwork(ctx, itemID)
	if err != nil {
		return MediaDetail{}, err
	}
	detail.Images = images
	for _, img := range images {
		if img.Kind == "backdrop" && detail.BackdropURL == "" {
			if img.SourceURL != "" {
				detail.BackdropURL = img.SourceURL
			} else {
				detail.BackdropURL = img.LocalPath
			}
		}
	}

	if item.Kind == LibraryKindTV {
		seasons, err := r.listSeasons(ctx, itemID, locale)
		if err != nil {
			return MediaDetail{}, err
		}
		detail.Seasons = seasons
	}

	if userID != "" {
		watchRepo := NewWatchRepo(r.q)
		if ws, err := watchRepo.GetWatchState(ctx, userID, itemID); err == nil {
			detail.WatchState = &WatchStateBrief{PositionMs: ws.PositionMs, Completed: ws.Completed}
		}
		favs, _ := watchRepo.ListFavorites(ctx, userID)
		for _, id := range favs {
			if id == itemID {
				detail.IsFavorite = true
				break
			}
		}
		wl, _ := watchRepo.ListWatchlist(ctx, userID)
		for _, id := range wl {
			if id == itemID {
				detail.InWatchlist = true
				break
			}
		}
	}

	if detail.Genres == nil {
		detail.Genres = []string{}
	}
	if detail.Cast == nil {
		detail.Cast = []CastMember{}
	}
	if detail.Images == nil {
		detail.Images = []ArtworkImage{}
	}

	return detail, nil
}

// DiscoverHome builds rule-based shelves for a user.
func (r *BrowseRepo) DiscoverHome(ctx context.Context, userID, locale string) (DiscoverHome, error) {
	home := DiscoverHome{Shelves: []DiscoverShelf{}}

	continueItems, err := r.continueWatching(ctx, userID, locale, 20)
	if err != nil {
		return DiscoverHome{}, err
	}
	if len(continueItems) > 0 {
		home.Shelves = append(home.Shelves, DiscoverShelf{
			ID: "continueWatching", TitleKey: "shelves.continueWatching", Items: continueItems,
		})
	}

	recent, err := r.recentlyAdded(ctx, locale, 20)
	if err != nil {
		return DiscoverHome{}, err
	}
	if len(recent) > 0 {
		home.Shelves = append(home.Shelves, DiscoverShelf{
			ID: "recentlyAdded", TitleKey: "shelves.recentlyAdded", Items: recent,
		})
	}

	recommended, err := r.recommended(ctx, userID, locale, 20)
	if err != nil {
		return DiscoverHome{}, err
	}
	if len(recommended) > 0 {
		home.Shelves = append(home.Shelves, DiscoverShelf{
			ID: "recommended", TitleKey: "shelves.recommended", Items: recommended,
		})
	}

	return home, nil
}

// SearchFTS runs FTS and hydrates matching items.
func (r *BrowseRepo) SearchFTS(ctx context.Context, query, locale string, limit int) ([]SearchResult, error) {
	ids, err := NewMediaRepo(r.q).SearchTitleFTS(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	out := make([]SearchResult, 0, len(ids))
	for _, id := range ids {
		item, err := NewMediaRepo(r.q).GetItemByID(ctx, id, locale)
		if err != nil {
			continue
		}
		out = append(out, SearchResult{MediaItem: item})
	}
	return out, nil
}

func (r *BrowseRepo) continueWatching(ctx context.Context, userID, locale string, limit int) ([]MediaItemSummary, error) {
	rows, err := r.q.QueryContext(ctx, `
		SELECT mi.id, mi.library_id, mi.kind, mi.sort_title, mi.year,
		       mi.match_status, mi.match_score, mi.created_at, mi.updated_at,
		       COALESCE(lt.value, mi.sort_title, '') AS title,
		       COALESCE(lo.value, '') AS overview,
		       COALESCE(a.source_url, a.local_path, '') AS poster,
		       mi.content_rating,
		       ws.position_ms, ws.completed
		FROM watch_state ws
		JOIN media_item mi ON mi.id = ws.media_item_id
		LEFT JOIN localized_text lt ON lt.media_item_id = mi.id AND lt.locale = ? AND lt.field = 'title'
		LEFT JOIN localized_text lo ON lo.media_item_id = mi.id AND lo.locale = ? AND lo.field = 'overview'
		LEFT JOIN artwork a ON a.media_item_id = mi.id AND a.kind = 'poster'
		WHERE ws.user_id = ? AND ws.position_ms > 0 AND ws.completed = 0
		ORDER BY ws.updated_at DESC
		LIMIT ?
	`, locale, locale, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("db discover continue: %w", err)
	}
	defer func() { _ = rows.Close() }()
	return scanMediaSummaries(rows)
}

func (r *BrowseRepo) recentlyAdded(ctx context.Context, locale string, limit int) ([]MediaItemSummary, error) {
	rows, err := r.q.QueryContext(ctx, `
		SELECT mi.id, mi.library_id, mi.kind, mi.sort_title, mi.year,
		       mi.match_status, mi.match_score, mi.created_at, mi.updated_at,
		       COALESCE(lt.value, mi.sort_title, '') AS title,
		       COALESCE(lo.value, '') AS overview,
		       COALESCE(a.source_url, a.local_path, '') AS poster,
		       mi.content_rating,
		       NULL, NULL
		FROM media_item mi
		LEFT JOIN localized_text lt ON lt.media_item_id = mi.id AND lt.locale = ? AND lt.field = 'title'
		LEFT JOIN localized_text lo ON lo.media_item_id = mi.id AND lo.locale = ? AND lo.field = 'overview'
		LEFT JOIN artwork a ON a.media_item_id = mi.id AND a.kind = 'poster'
		ORDER BY mi.created_at DESC
		LIMIT ?
	`, locale, locale, limit)
	if err != nil {
		return nil, fmt.Errorf("db discover recent: %w", err)
	}
	defer func() { _ = rows.Close() }()
	return scanMediaSummaries(rows)
}

func (r *BrowseRepo) recommended(ctx context.Context, userID, locale string, limit int) ([]MediaItemSummary, error) {
	var lastGenre sql.NullString
	_ = r.q.QueryRowContext(ctx, `
		SELECT g.name FROM watch_state ws
		JOIN media_genre mg ON mg.media_item_id = ws.media_item_id
		JOIN genre g ON g.id = mg.genre_id
		WHERE ws.user_id = ?
		ORDER BY ws.updated_at DESC
		LIMIT 1
	`, userID).Scan(&lastGenre)

	var rows *sql.Rows
	var err error
	if lastGenre.Valid && lastGenre.String != "" {
		rows, err = r.q.QueryContext(ctx, `
			SELECT mi.id, mi.library_id, mi.kind, mi.sort_title, mi.year,
			       mi.match_status, mi.match_score, mi.created_at, mi.updated_at,
			       COALESCE(lt.value, mi.sort_title, '') AS title,
			       COALESCE(lo.value, '') AS overview,
			       COALESCE(a.source_url, a.local_path, '') AS poster,
			       mi.content_rating,
			       NULL, NULL
			FROM media_item mi
			JOIN media_genre mg ON mg.media_item_id = mi.id
			JOIN genre g ON g.id = mg.genre_id AND g.name = ?
			LEFT JOIN localized_text lt ON lt.media_item_id = mi.id AND lt.locale = ? AND lt.field = 'title'
			LEFT JOIN localized_text lo ON lo.media_item_id = mi.id AND lo.locale = ? AND lo.field = 'overview'
			LEFT JOIN artwork a ON a.media_item_id = mi.id AND a.kind = 'poster'
			ORDER BY RANDOM()
			LIMIT ?
		`, lastGenre.String, locale, locale, limit)
	} else {
		rows, err = r.q.QueryContext(ctx, `
			SELECT mi.id, mi.library_id, mi.kind, mi.sort_title, mi.year,
			       mi.match_status, mi.match_score, mi.created_at, mi.updated_at,
			       COALESCE(lt.value, mi.sort_title, '') AS title,
			       COALESCE(lo.value, '') AS overview,
			       COALESCE(a.source_url, a.local_path, '') AS poster,
			       mi.content_rating,
			       NULL, NULL
			FROM media_item mi
			LEFT JOIN localized_text lt ON lt.media_item_id = mi.id AND lt.locale = ? AND lt.field = 'title'
			LEFT JOIN localized_text lo ON lo.media_item_id = mi.id AND lo.locale = ? AND lo.field = 'overview'
			LEFT JOIN artwork a ON a.media_item_id = mi.id AND a.kind = 'poster'
			ORDER BY RANDOM()
			LIMIT ?
		`, locale, locale, limit)
	}
	if err != nil {
		return nil, fmt.Errorf("db discover recommended: %w", err)
	}
	defer func() { _ = rows.Close() }()
	return scanMediaSummaries(rows)
}

func (r *BrowseRepo) listGenres(ctx context.Context, itemID int64) ([]string, error) {
	rows, err := r.q.QueryContext(ctx, `
		SELECT g.name FROM genre g
		JOIN media_genre mg ON mg.genre_id = g.id
		WHERE mg.media_item_id = ?
		ORDER BY g.name
	`, itemID)
	if err != nil {
		return nil, fmt.Errorf("db detail genres: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var out []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		out = append(out, name)
	}
	return out, rows.Err()
}

func (r *BrowseRepo) listCast(ctx context.Context, itemID int64) ([]CastMember, error) {
	rows, err := r.q.QueryContext(ctx, `
		SELECT p.name, mp.role, mp.sort_order
		FROM media_person mp
		JOIN person p ON p.id = mp.person_id
		WHERE mp.media_item_id = ?
		ORDER BY mp.sort_order, p.name
		LIMIT 24
	`, itemID)
	if err != nil {
		return nil, fmt.Errorf("db detail cast: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var out []CastMember
	for rows.Next() {
		var c CastMember
		if err := rows.Scan(&c.Name, &c.Role, &c.Order); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *BrowseRepo) listArtwork(ctx context.Context, itemID int64) ([]ArtworkImage, error) {
	rows, err := r.q.QueryContext(ctx, `
		SELECT kind, source_url, local_path, width, height
		FROM artwork WHERE media_item_id = ?
	`, itemID)
	if err != nil {
		return nil, fmt.Errorf("db detail artwork: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var out []ArtworkImage
	for rows.Next() {
		var img ArtworkImage
		var src, local sql.NullString
		var w, h sql.NullInt64
		if err := rows.Scan(&img.Kind, &src, &local, &w, &h); err != nil {
			return nil, err
		}
		img.SourceURL = src.String
		img.LocalPath = local.String
		if w.Valid {
			n := int(w.Int64)
			img.Width = &n
		}
		if h.Valid {
			n := int(h.Int64)
			img.Height = &n
		}
		out = append(out, img)
	}
	return out, rows.Err()
}

func (r *BrowseRepo) listSeasons(ctx context.Context, itemID int64, _ string) ([]SeasonDetail, error) {
	rows, err := r.q.QueryContext(ctx, `
		SELECT s.season_number, e.id, e.episode_number,
		       COALESCE(e.sort_title, '') AS title
		FROM season s
		JOIN episode e ON e.season_id = s.id
		WHERE s.media_item_id = ?
		ORDER BY s.season_number, e.episode_number
	`, itemID)
	if err != nil {
		return nil, fmt.Errorf("db detail seasons: %w", err)
	}
	defer func() { _ = rows.Close() }()

	bySeason := map[int]*SeasonDetail{}
	var order []int
	for rows.Next() {
		var seasonNum, epNum int
		var epID int64
		var title string
		if err := rows.Scan(&seasonNum, &epID, &epNum, &title); err != nil {
			return nil, err
		}
		if _, ok := bySeason[seasonNum]; !ok {
			bySeason[seasonNum] = &SeasonDetail{SeasonNumber: seasonNum}
			order = append(order, seasonNum)
		}
		bySeason[seasonNum].Episodes = append(bySeason[seasonNum].Episodes, EpisodeSummary{
			ID: epID, SeasonNumber: seasonNum, EpisodeNumber: epNum, Title: title,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	out := make([]SeasonDetail, 0, len(order))
	for _, n := range order {
		out = append(out, *bySeason[n])
	}
	return out, nil
}

func scanMediaSummaries(rows *sql.Rows) ([]MediaItemSummary, error) {
	var out []MediaItemSummary
	for rows.Next() {
		var mi MediaItem
		var sortTitle, matchStatus sql.NullString
		var year sql.NullInt64
		var score sql.NullFloat64
		var contentRating sql.NullString
		var created, updated string
		var pos sql.NullInt64
		var completed sql.NullInt64
		if err := rows.Scan(
			&mi.ID, &mi.LibraryID, &mi.Kind, &sortTitle, &year,
			&matchStatus, &score, &created, &updated,
			&mi.Title, &mi.Overview, &mi.PosterURL, &contentRating,
			&pos, &completed,
		); err != nil {
			return nil, fmt.Errorf("db browse scan summary: %w", err)
		}
		mi.SortTitle = sortTitle.String
		mi.Year = nullIntPtr(year)
		mi.MatchStatus = MatchStatus(matchStatus.String)
		mi.MatchScore = nullFloatPtr(score)
		if contentRating.Valid {
			s := contentRating.String
			mi.ContentRating = &s
		}
		summary := MediaItemSummary{MediaItem: mi}
		if pos.Valid {
			summary.WatchState = &WatchStateBrief{
				PositionMs: pos.Int64,
				Completed:  completed.Valid && completed.Int64 != 0,
			}
		}
		out = append(out, summary)
	}
	return out, rows.Err()
}
