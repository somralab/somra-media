package plugin

import (
	"context"
	"time"
)

// MediaKind distinguishes movie vs TV show search targets.
type MediaKind string

const (
	MediaKindMovie MediaKind = "movie"
	MediaKindTV    MediaKind = "tv"
)

// Protocol identifies the acquisition transport for a release.
type Protocol string

const (
	ProtocolTorrent Protocol = "torrent"
	ProtocolUsenet  Protocol = "usenet"
)

// SearchQuery carries normalized search parameters for an indexer.
type SearchQuery struct {
	Title      string    `json:"title"`
	Year       *int      `json:"year,omitempty"`
	MediaKind  MediaKind `json:"mediaKind"`
	Season     *int      `json:"season,omitempty"`
	Episode    *int      `json:"episode,omitempty"`
	Categories []int     `json:"categories,omitempty"`
}

// Category describes an indexer content category.
type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Limits describes indexer-imposed constraints.
type Limits struct {
	MaxPageSize int `json:"maxPageSize,omitempty"`
	MaxResults  int `json:"maxResults,omitempty"`
}

// Capabilities describes what an indexer adapter supports.
type Capabilities struct {
	Protocols      []Protocol `json:"protocols"`
	Categories     []Category `json:"categories,omitempty"`
	Limits         Limits     `json:"limits,omitempty"`
	SupportsSearch bool       `json:"supportsSearch"`
	SupportsRSS    bool       `json:"supportsRSS"`
}

// SearchResult is a normalized release candidate from an indexer.
type SearchResult struct {
	ReleaseID  string     `json:"releaseId"`
	IndexerID  string     `json:"indexerId"`
	Title      string     `json:"title"`
	Protocol   Protocol   `json:"protocol"`
	SizeBytes  int64      `json:"sizeBytes"`
	Seeders    *int       `json:"seeders,omitempty"`
	Peers      *int       `json:"peers,omitempty"`
	PublishAt  *time.Time `json:"publishAt,omitempty"`
	Resolution string     `json:"resolution,omitempty"`
	Codec      string     `json:"codec,omitempty"`
	Source     string     `json:"source,omitempty"`
}

// Indexer searches external indexers and reports adapter capabilities.
type Indexer interface {
	Plugin
	Capabilities(ctx context.Context) (Capabilities, error)
	Search(ctx context.Context, q SearchQuery) ([]SearchResult, error)
}
