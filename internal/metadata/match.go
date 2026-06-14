package metadata

import (
	"context"
	"fmt"
	"math"
	"strings"
)

// ScoreMatch ranks a candidate against the search query.
func ScoreMatch(q SearchQuery, title string, year *int) float64 {
	queryTitle := normalizeTitle(q.Title)
	candidate := normalizeTitle(title)
	if queryTitle == "" || candidate == "" {
		return 0
	}
	if queryTitle == candidate {
		score := 1.0
		if q.Year != nil && year != nil && *q.Year == *year {
			score += 0.2
		}
		return math.Min(score, 1.0)
	}
	if strings.Contains(candidate, queryTitle) || strings.Contains(queryTitle, candidate) {
		score := 0.7
		if q.Year != nil && year != nil && *q.Year == *year {
			score += 0.15
		}
		return score
	}
	if q.Year != nil && year != nil && *q.Year == *year {
		return 0.4
	}
	return 0.1
}

func normalizeTitle(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	replacer := strings.NewReplacer(":", "", "-", " ", "_", " ", ".", " ")
	return strings.Join(strings.Fields(replacer.Replace(s)), " ")
}

// Matcher selects the best provider result for a query.
type Matcher struct {
	Registry *Registry
	Limiter  *RateLimiter
}

// Match searches all providers and returns ranked results.
func (m *Matcher) Match(ctx context.Context, q SearchQuery) ([]SearchResult, error) {
	var all []SearchResult
	for _, p := range m.Registry.All() {
		if m.Limiter != nil {
			if err := m.Limiter.Wait(ctx, p.Name()); err != nil {
				return nil, err
			}
		}
		results, err := p.Search(ctx, q)
		if err != nil {
			continue
		}
		all = append(all, results...)
	}
	sortResults(all)
	return all, nil
}

func sortResults(r []SearchResult) {
	for i := 0; i < len(r); i++ {
		for j := i + 1; j < len(r); j++ {
			if r[j].Score > r[i].Score {
				r[i], r[j] = r[j], r[i]
			}
		}
	}
}

// Service applies metadata matching to media items.
type Service struct {
	DB       MediaStore
	Registry *Registry
	Matcher  *Matcher
}

// MediaStore abstracts persistence for metadata operations.
type MediaStore interface {
	GetItem(ctx context.Context, id int64, locale string) (MediaItemView, error)
	ListUnmatched(ctx context.Context, libraryID int64, limit int) ([]MediaItemView, error)
	ApplyMatch(ctx context.Context, itemID int64, provider, externalID, locale string, detail Detail) error
}

// MediaItemView is a read model for matching.
type MediaItemView struct {
	ID          int64
	LibraryID   int64
	Kind        string
	ParsedTitle string
	ParsedYear  *int
	Locale      string
}

// AutoMatch attempts automatic matching for unmatched items in a library.
func (s *Service) AutoMatch(ctx context.Context, libraryID int64, locale string, limit int) (int, error) {
	items, err := s.DB.ListUnmatched(ctx, libraryID, limit)
	if err != nil {
		return 0, err
	}
	matched := 0
	for _, item := range items {
		q := SearchQuery{
			Title:  item.ParsedTitle,
			Year:   item.ParsedYear,
			Kind:   item.Kind,
			Locale: locale,
		}
		results, err := s.Matcher.Match(ctx, q)
		if err != nil || len(results) == 0 {
			continue
		}
		best := results[0]
		if best.Score < 0.6 {
			continue
		}
		provider, ok := s.Registry.Get(best.Provider)
		if !ok {
			continue
		}
		detail, err := provider.Detail(ctx, best.ExternalID, locale)
		if err != nil {
			continue
		}
		if err := s.DB.ApplyMatch(ctx, item.ID, best.Provider, best.ExternalID, locale, detail); err != nil {
			return matched, err
		}
		matched++
	}
	return matched, nil
}

// Rematch applies a manual provider selection.
func (s *Service) Rematch(ctx context.Context, itemID int64, providerName, externalID, locale string) error {
	provider, ok := s.Registry.Get(providerName)
	if !ok {
		return fmt.Errorf("unknown provider %q", providerName)
	}
	detail, err := provider.Detail(ctx, externalID, locale)
	if err != nil {
		return err
	}
	return s.DB.ApplyMatch(ctx, itemID, providerName, externalID, locale, detail)
}

// SearchCandidates returns provider search results for rematch UI.
func (s *Service) SearchCandidates(ctx context.Context, itemID int64, locale string) ([]SearchResult, error) {
	item, err := s.DB.GetItem(ctx, itemID, locale)
	if err != nil {
		return nil, err
	}
	return s.Matcher.Match(ctx, SearchQuery{
		Title:  item.ParsedTitle,
		Year:   item.ParsedYear,
		Kind:   item.Kind,
		Locale: locale,
	})
}
