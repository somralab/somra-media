package metadata

import "context"

// TVDBProvider is a minimal TVDB stub for series metadata.
type TVDBProvider struct {
	APIKey string
}

func (p *TVDBProvider) Name() string { return "tvdb" }

func (p *TVDBProvider) Search(ctx context.Context, q SearchQuery) ([]SearchResult, error) {
	return nil, nil
}

func (p *TVDBProvider) Detail(ctx context.Context, externalID, locale string) (Detail, error) {
	return Detail{Provider: p.Name(), ExternalID: externalID}, nil
}

func (p *TVDBProvider) Images(ctx context.Context, externalID string) (string, string, string, error) {
	return "", "", "", nil
}

// MusicBrainzProvider is a minimal MusicBrainz stub.
type MusicBrainzProvider struct{}

func (p *MusicBrainzProvider) Name() string { return "musicbrainz" }

func (p *MusicBrainzProvider) Search(ctx context.Context, q SearchQuery) ([]SearchResult, error) {
	return nil, nil
}

func (p *MusicBrainzProvider) Detail(ctx context.Context, externalID, locale string) (Detail, error) {
	return Detail{Provider: p.Name(), ExternalID: externalID}, nil
}

func (p *MusicBrainzProvider) Images(ctx context.Context, externalID string) (string, string, string, error) {
	return "", "", "", nil
}

// FanartProvider is a minimal fanart.tv stub.
type FanartProvider struct {
	APIKey string
}

func (p *FanartProvider) Name() string { return "fanart" }

func (p *FanartProvider) Search(ctx context.Context, q SearchQuery) ([]SearchResult, error) {
	return nil, nil
}

func (p *FanartProvider) Detail(ctx context.Context, externalID, locale string) (Detail, error) {
	return Detail{Provider: p.Name(), ExternalID: externalID}, nil
}

func (p *FanartProvider) Images(ctx context.Context, externalID string) (string, string, string, error) {
	return "", "", "", nil
}

// TestProvider is an offline TMDB-shaped provider for unit tests.
type TestProvider struct{}

func (TestProvider) Name() string { return "tmdb" }

func (TestProvider) Search(_ context.Context, q SearchQuery) ([]SearchResult, error) {
	y := 2010
	if q.Year != nil {
		y = *q.Year
	}
	extID := q.Title
	if extID == "" {
		extID = "1"
	}
	return []SearchResult{{
		Provider: "tmdb", ExternalID: extID, Title: q.Title, Year: &y, Score: 0.95,
	}}, nil
}

func (TestProvider) Detail(_ context.Context, externalID, _ string) (Detail, error) {
	return Detail{
		Provider: "tmdb", ExternalID: externalID,
		Title: "Inception", Overview: "A mind-bending thriller",
		PosterURL: "https://image.tmdb.org/t/p/w500/test.jpg",
	}, nil
}

func (TestProvider) Images(context.Context, string) (string, string, string, error) {
	return "", "", "", nil
}

// MockProvider returns deterministic results for tests.
type MockProvider struct {
	Results      []SearchResult
	PresetDetail Detail
}

func (p *MockProvider) Name() string { return "mock" }

func (p *MockProvider) Search(_ context.Context, q SearchQuery) ([]SearchResult, error) {
	if len(p.Results) > 0 {
		return p.Results, nil
	}
	y := 2020
	if q.Year != nil {
		y = *q.Year
	}
	return []SearchResult{{
		Provider: "mock", ExternalID: "1", Title: q.Title, Year: &y, Score: 0.95,
		PosterURL: "https://image.tmdb.org/t/p/w500/mock.jpg",
	}}, nil
}

func (p *MockProvider) Detail(_ context.Context, externalID, _ string) (Detail, error) {
	if p.PresetDetail.ExternalID != "" {
		return p.PresetDetail, nil
	}
	return Detail{
		Provider: "mock", ExternalID: externalID,
		Title: "Mock Title", Overview: "Mock overview",
		PosterURL: "https://image.tmdb.org/t/p/w500/mock.jpg",
	}, nil
}

func (p *MockProvider) Images(_ context.Context, _ string) (string, string, string, error) {
	return "https://image.tmdb.org/t/p/w500/mock.jpg", "", "", nil
}
