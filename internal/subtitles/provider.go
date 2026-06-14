package subtitles

import "context"

// SearchQuery identifies a media work for subtitle lookup.
type SearchQuery struct {
	Title    string
	Year     *int
	Language string
	FileHash string
}

// SearchResult is a candidate subtitle from a provider.
type SearchResult struct {
	Provider      string  `json:"provider"`
	ExternalID    string  `json:"externalId"`
	Language      string  `json:"language"`
	ReleaseName   string  `json:"releaseName,omitempty"`
	Score         float64 `json:"score"`
	DownloadCount int     `json:"downloadCount,omitempty"`
}

// Provider searches and downloads subtitles from an external source.
type Provider interface {
	Name() string
	Search(ctx context.Context, q SearchQuery) ([]SearchResult, error)
	Download(ctx context.Context, externalID, language string) ([]byte, error)
}
