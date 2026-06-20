package indexer

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/somralab/somra-media/internal/automation/releaseparse"
	"github.com/somralab/somra-media/internal/plugin"
)

// SearchService fans out queries to enabled indexers.
type SearchService struct {
	Manager *plugin.Manager
	Logger  *slog.Logger
	Timeout time.Duration
}

// SearchRequest carries query parameters and optional indexer filter.
type SearchRequest struct {
	Query      plugin.SearchQuery
	IndexerIDs []int64
}

// SearchResponse merges indexer results.
type SearchResponse struct {
	Results []plugin.SearchResult `json:"results"`
	Errors  []IndexerError        `json:"errors,omitempty"`
}

// IndexerError records a per-indexer failure.
type IndexerError struct {
	IndexerID string `json:"indexerId"`
	Error     string `json:"error"`
}

// Search queries enabled indexers in parallel.
func (s *SearchService) Search(ctx context.Context, req SearchRequest) (SearchResponse, error) {
	if s == nil || s.Manager == nil {
		return SearchResponse{}, fmt.Errorf("automation search: manager required")
	}
	timeout := s.Timeout
	if timeout <= 0 {
		timeout = 20 * time.Second
	}
	filter := make(map[string]struct{}, len(req.IndexerIDs))
	for _, id := range req.IndexerIDs {
		filter[strconv.FormatInt(id, 10)] = struct{}{}
	}
	indexers := s.Manager.EnabledIndexers()
	if len(filter) > 0 {
		filtered := make([]plugin.Indexer, 0, len(indexers))
		for _, idx := range indexers {
			if _, ok := filter[idx.ID()]; ok {
				filtered = append(filtered, idx)
			}
		}
		indexers = filtered
	}
	var (
		mu      sync.Mutex
		results []plugin.SearchResult
		errs    []IndexerError
	)
	g, gctx := errgroup.WithContext(ctx)
	for _, idx := range indexers {
		idx := idx
		g.Go(func() error {
			cctx, cancel := context.WithTimeout(gctx, timeout)
			defer cancel()
			rows, err := idx.Search(cctx, req.Query)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = append(errs, IndexerError{IndexerID: idx.ID(), Error: err.Error()})
				return nil
			}
			for _, row := range rows {
				row.IndexerID = idx.ID()
				row = releaseparse.Enrich(row)
				results = append(results, row)
			}
			return nil
		})
	}
	_ = g.Wait()
	results = dedupeResults(results)
	return SearchResponse{Results: results, Errors: errs}, nil
}

func dedupeResults(in []plugin.SearchResult) []plugin.SearchResult {
	seen := make(map[string]struct{}, len(in))
	out := make([]plugin.SearchResult, 0, len(in))
	for _, r := range in {
		key := dedupeKey(r)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, r)
	}
	return out
}

func dedupeKey(r plugin.SearchResult) string {
	title := strings.ToLower(strings.TrimSpace(r.Title))
	return fmt.Sprintf("%s|%d|%s", title, r.SizeBytes, r.Protocol)
}
