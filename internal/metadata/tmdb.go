package metadata

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// TMDBProvider implements MetadataProvider for The Movie Database.
type TMDBProvider struct {
	APIKey string
	Client *http.Client
	Base   string
	fetch  func(ctx context.Context, client *http.Client, method, rawURL string) (*http.Response, error)
}

func (p *TMDBProvider) doHTTP(ctx context.Context, method, rawURL string) (*http.Response, error) {
	if p.fetch != nil {
		return p.fetch(ctx, p.Client, method, rawURL)
	}
	return DoRequest(ctx, p.Client, method, rawURL)
}

// NewTMDBProvider returns a TMDB provider.
func NewTMDBProvider(apiKey string, client *http.Client) *TMDBProvider {
	if client == nil {
		client = SafeHTTPClient(10 * time.Second)
	}
	return &TMDBProvider{
		APIKey: apiKey,
		Client: client,
		Base:   "https://api.themoviedb.org/3",
	}
}

func (p *TMDBProvider) Name() string { return "tmdb" }

func (p *TMDBProvider) Search(ctx context.Context, q SearchQuery) ([]SearchResult, error) {
	if p.APIKey == "" {
		return nil, fmt.Errorf("tmdb: api key missing")
	}
	endpoint := "/search/movie"
	if strings.EqualFold(q.Kind, "tv") {
		endpoint = "/search/tv"
	}
	params := url.Values{}
	params.Set("api_key", p.APIKey)
	params.Set("query", q.Title)
	params.Set("language", localeToTMDB(q.Locale))
	if q.Year != nil {
		if endpoint == "/search/movie" {
			params.Set("year", strconv.Itoa(*q.Year))
		} else {
			params.Set("first_air_date_year", strconv.Itoa(*q.Year))
		}
	}
	raw := p.Base + endpoint + "?" + params.Encode()
	resp, err := p.doHTTP(ctx, http.MethodGet, raw)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("tmdb search: http %d", resp.StatusCode)
	}

	var parsed struct {
		Results []struct {
			ID           int    `json:"id"`
			Title        string `json:"title"`
			Name         string `json:"name"`
			Overview     string `json:"overview"`
			PosterPath   string `json:"poster_path"`
			ReleaseDate  string `json:"release_date"`
			FirstAirDate string `json:"first_air_date"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}

	out := make([]SearchResult, 0, len(parsed.Results))
	for _, r := range parsed.Results {
		title := r.Title
		if title == "" {
			title = r.Name
		}
		year := parseYear(r.ReleaseDate, r.FirstAirDate)
		poster := ""
		if r.PosterPath != "" {
			poster = "https://image.tmdb.org/t/p/w500" + r.PosterPath
		}
		out = append(out, SearchResult{
			Provider:   p.Name(),
			ExternalID: strconv.Itoa(r.ID),
			Title:      title,
			Year:       year,
			Overview:   r.Overview,
			PosterURL:  poster,
			Score:      ScoreMatch(q, title, year),
		})
	}
	return out, nil
}

func (p *TMDBProvider) Detail(ctx context.Context, externalID, locale string) (Detail, error) {
	raw := fmt.Sprintf("%s/movie/%s?api_key=%s&language=%s", p.Base, externalID, url.QueryEscape(p.APIKey), localeToTMDB(locale))
	resp, err := p.doHTTP(ctx, http.MethodGet, raw)
	if err != nil {
		return Detail{}, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	var r struct {
		ID           int    `json:"id"`
		Title        string `json:"title"`
		Overview     string `json:"overview"`
		PosterPath   string `json:"poster_path"`
		BackdropPath string `json:"backdrop_path"`
		ReleaseDate  string `json:"release_date"`
		Genres       []struct {
			Name string `json:"name"`
		} `json:"genres"`
	}
	if err := json.Unmarshal(body, &r); err != nil {
		return Detail{}, err
	}
	genres := make([]string, 0, len(r.Genres))
	for _, g := range r.Genres {
		genres = append(genres, g.Name)
	}
	poster, backdrop := "", ""
	if r.PosterPath != "" {
		poster = "https://image.tmdb.org/t/p/w500" + r.PosterPath
	}
	if r.BackdropPath != "" {
		backdrop = "https://image.tmdb.org/t/p/original" + r.BackdropPath
	}
	return Detail{
		Provider:    p.Name(),
		ExternalID:  externalID,
		Title:       r.Title,
		Overview:    r.Overview,
		Year:        parseYear(r.ReleaseDate, ""),
		Genres:      genres,
		PosterURL:   poster,
		BackdropURL: backdrop,
	}, nil
}

func (p *TMDBProvider) Images(_ context.Context, externalID string) (string, string, string, error) {
	return "", "", "", fmt.Errorf("tmdb images: use detail")
}

func localeToTMDB(locale string) string {
	switch locale {
	case "tr-TR":
		return "tr-TR"
	default:
		return "en-US"
	}
}

func parseYear(dates ...string) *int {
	for _, d := range dates {
		if len(d) >= 4 {
			if y, err := strconv.Atoi(d[:4]); err == nil {
				return &y
			}
		}
	}
	return nil
}

// Registry holds registered metadata providers.
type Registry struct {
	mu        sync.RWMutex
	providers map[string]MetadataProvider
}

// NewRegistry returns an empty provider registry.
func NewRegistry() *Registry {
	return &Registry{providers: make(map[string]MetadataProvider)}
}

// Register adds a provider.
func (r *Registry) Register(p MetadataProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[p.Name()] = p
}

// Get returns a provider by name.
func (r *Registry) Get(name string) (MetadataProvider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.providers[name]
	return p, ok
}

// All returns all providers.
func (r *Registry) All() []MetadataProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]MetadataProvider, 0, len(r.providers))
	for _, p := range r.providers {
		out = append(out, p)
	}
	return out
}

// ResponseCache caches provider HTTP responses in memory.
type ResponseCache struct {
	mu    sync.RWMutex
	items map[string]cacheEntry
	ttl   time.Duration
}

type cacheEntry struct {
	value     []byte
	expiresAt time.Time
}

// NewResponseCache returns a cache with ttl.
func NewResponseCache(ttl time.Duration) *ResponseCache {
	if ttl <= 0 {
		ttl = time.Hour
	}
	return &ResponseCache{items: make(map[string]cacheEntry), ttl: ttl}
}

// Get returns cached bytes if present.
func (c *ResponseCache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.items[key]
	if !ok || time.Now().After(e.expiresAt) {
		return nil, false
	}
	return e.value, true
}

// Set stores bytes under key.
func (c *ResponseCache) Set(key string, value []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = cacheEntry{value: value, expiresAt: time.Now().Add(c.ttl)}
}

// RateLimiter provides a simple token bucket per provider.
type RateLimiter struct {
	mu       sync.Mutex
	interval time.Duration
	last     map[string]time.Time
}

// NewRateLimiter returns a limiter with min interval between calls per key.
func NewRateLimiter(interval time.Duration) *RateLimiter {
	if interval <= 0 {
		interval = 250 * time.Millisecond
	}
	return &RateLimiter{interval: interval, last: make(map[string]time.Time)}
}

// Wait blocks until the key may proceed.
func (l *RateLimiter) Wait(ctx context.Context, key string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if t, ok := l.last[key]; ok {
		wait := l.interval - time.Since(t)
		if wait > 0 {
			timer := time.NewTimer(wait)
			defer timer.Stop()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-timer.C:
			}
		}
	}
	l.last[key] = time.Now()
	return nil
}
