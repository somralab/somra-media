package requests

import (
	"context"
	"fmt"

	"github.com/somralab/somra-media/internal/metadata"
)

// DiscoverSearchParams are inputs for GET /requests/discover.
type DiscoverSearchParams struct {
	Query  string
	Kind   MediaKind
	Year   *int
	Locale string
	Limit  int
}

// DiscoverHit is a provider result eligible for requesting (not in library).
type DiscoverHit struct {
	Provider   string  `json:"provider"`
	ExternalID string  `json:"externalId"`
	Title      string  `json:"title"`
	Year       *int    `json:"year,omitempty"`
	Overview   string  `json:"overview,omitempty"`
	PosterURL  string  `json:"posterUrl,omitempty"`
	Score      float64 `json:"score"`
}

// Discoverer searches TMDB and excludes items already linked in the library.
type Discoverer struct {
	Registry *metadata.Registry
	Library  LibraryLookup
	Limiter  *metadata.RateLimiter
	Provider string
}

// Search runs TMDB discover search and filters out in-library matches.
func (d *Discoverer) Search(ctx context.Context, params DiscoverSearchParams) ([]DiscoverHit, error) {
	if d == nil || d.Registry == nil {
		return nil, fmt.Errorf("requests discover: registry is nil")
	}
	providerName := d.Provider
	if providerName == "" {
		providerName = "tmdb"
	}
	provider, ok := d.Registry.Get(providerName)
	if !ok {
		return nil, fmt.Errorf("requests discover: provider %q not registered", providerName)
	}
	if d.Limiter != nil {
		if err := d.Limiter.Wait(ctx, providerName); err != nil {
			return nil, fmt.Errorf("requests discover rate limit: %w", err)
		}
	}
	locale := params.Locale
	if locale == "" {
		locale = "en-US"
	}
	results, err := provider.Search(ctx, metadata.SearchQuery{
		Title:  params.Query,
		Year:   params.Year,
		Kind:   string(params.Kind),
		Locale: locale,
	})
	if err != nil {
		return nil, fmt.Errorf("requests discover search: %w", err)
	}
	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}
	hits := make([]DiscoverHit, 0, len(results))
	for _, r := range results {
		if len(hits) >= limit {
			break
		}
		if d.Library != nil {
			found, _, err := d.Library.ExistsByProviderID(ctx, r.Provider, r.ExternalID)
			if err != nil {
				return nil, fmt.Errorf("requests discover library filter: %w", err)
			}
			if found {
				continue
			}
		}
		hits = append(hits, DiscoverHit{
			Provider:   r.Provider,
			ExternalID: r.ExternalID,
			Title:      r.Title,
			Year:       r.Year,
			Overview:   r.Overview,
			PosterURL:  r.PosterURL,
			Score:      r.Score,
		})
	}
	return hits, nil
}
