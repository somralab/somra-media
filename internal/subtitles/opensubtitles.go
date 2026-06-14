package subtitles

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/somralab/somra-media/internal/metadata"
)

const openSubtitlesBase = "https://api.opensubtitles.com/api/v1"

// OpenSubtitles implements the Provider interface for OpenSubtitles REST API v1.
type OpenSubtitles struct {
	APIKey string
	Client *http.Client
}

// NewOpenSubtitles returns a provider with SSRF-safe HTTP client.
func NewOpenSubtitles(apiKey string) *OpenSubtitles {
	return &OpenSubtitles{
		APIKey: apiKey,
		Client: metadata.SafeHTTPClient(15 * time.Second),
	}
}

func (p *OpenSubtitles) Name() string { return "opensubtitles" }

func (p *OpenSubtitles) Search(ctx context.Context, q SearchQuery) ([]SearchResult, error) {
	if p.APIKey == "" {
		return nil, fmt.Errorf("subtitles: opensubtitles api key not configured")
	}
	params := []string{"query=" + urlQueryEscape(q.Title)}
	if q.Year != nil {
		params = append(params, fmt.Sprintf("year=%d", *q.Year))
	}
	if q.Language != "" {
		params = append(params, "languages="+urlQueryEscape(q.Language))
	}
	rawURL := openSubtitlesBase + "/subtitles?" + strings.Join(params, "&")
	resp, err := p.doAPI(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("subtitles opensubtitles search: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("subtitles opensubtitles search: status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, err
	}
	var parsed osSearchResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("subtitles opensubtitles decode: %w", err)
	}
	out := make([]SearchResult, 0, len(parsed.Data))
	for _, item := range parsed.Data {
		out = append(out, SearchResult{
			Provider:      p.Name(),
			ExternalID:    item.ID,
			Language:      item.Attributes.Language,
			ReleaseName:   item.Attributes.Release,
			Score:         scoreResult(q, item),
			DownloadCount: item.Attributes.DownloadCount,
		})
	}
	return out, nil
}

func (p *OpenSubtitles) Download(ctx context.Context, externalID, language string) ([]byte, error) {
	if p.APIKey == "" {
		return nil, fmt.Errorf("subtitles: opensubtitles api key not configured")
	}
	rawURL := openSubtitlesBase + "/download"
	reqBody := fmt.Sprintf(`{"file_id":%s}`, jsonQuote(externalID))
	resp, err := p.doAPI(ctx, http.MethodPost, rawURL, strings.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("subtitles opensubtitles download: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("subtitles opensubtitles download: status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, err
	}
	var parsed struct {
		Link string `json:"link"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	if parsed.Link == "" {
		return nil, fmt.Errorf("subtitles opensubtitles: empty download link")
	}
	dlResp, err := metadata.DoRequest(ctx, p.Client, http.MethodGet, parsed.Link)
	if err != nil {
		return nil, err
	}
	defer func() { _ = dlResp.Body.Close() }()
	return io.ReadAll(io.LimitReader(dlResp.Body, 4<<20))
}

func (p *OpenSubtitles) doAPI(ctx context.Context, method, rawURL string, body io.Reader) (*http.Response, error) {
	if err := metadata.ValidateOutboundURLContext(ctx, rawURL); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, method, rawURL, body)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Api-Key", p.APIKey)
	return p.Client.Do(req)
}

type osSearchItem struct {
	ID         string `json:"id"`
	Attributes struct {
		Language       string `json:"language"`
		Release        string `json:"release"`
		DownloadCount  int    `json:"download_count"`
		FeatureDetails struct {
			Title string `json:"title"`
			Year  int    `json:"year"`
		} `json:"feature_details"`
	} `json:"attributes"`
}

type osSearchResponse struct {
	Data []osSearchItem `json:"data"`
}

func scoreResult(q SearchQuery, item osSearchItem) float64 {
	score := 50.0
	title := strings.ToLower(item.Attributes.FeatureDetails.Title)
	if strings.Contains(title, strings.ToLower(q.Title)) {
		score += 25
	}
	if q.Year != nil && item.Attributes.FeatureDetails.Year == *q.Year {
		score += 20
	}
	if q.Language != "" && strings.EqualFold(item.Attributes.Language, q.Language) {
		score += 5
	}
	if score > 100 {
		return 100
	}
	return score
}

func urlQueryEscape(s string) string {
	return strings.ReplaceAll(s, " ", "%20")
}

func jsonQuote(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
