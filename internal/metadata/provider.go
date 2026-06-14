package metadata

import (
	"context"
)

// SearchQuery carries pre-parsed hints for provider search.
type SearchQuery struct {
	Title   string
	Year    *int
	Kind    string
	Locale  string
	Season  *int
	Episode *int
}

// SearchResult is a candidate match from a provider.
type SearchResult struct {
	Provider   string  `json:"provider"`
	ExternalID string  `json:"externalId"`
	Title      string  `json:"title"`
	Year       *int    `json:"year,omitempty"`
	Overview   string  `json:"overview,omitempty"`
	PosterURL  string  `json:"posterUrl,omitempty"`
	Score      float64 `json:"score"`
}

// Detail holds rich metadata for a matched item.
type Detail struct {
	Provider    string
	ExternalID  string
	Title       string
	Overview    string
	Year        *int
	Genres      []string
	PosterURL   string
	BackdropURL string
}

// MetadataProvider is the common contract for external metadata sources.
type MetadataProvider interface {
	Name() string
	Search(ctx context.Context, q SearchQuery) ([]SearchResult, error)
	Detail(ctx context.Context, externalID, locale string) (Detail, error)
	Images(ctx context.Context, externalID string) (poster, backdrop, logo string, err error)
}
