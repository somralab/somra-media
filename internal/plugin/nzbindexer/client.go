//go:build acquisition

package nzbindexer

import (
	"context"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/somralab/somra-media/internal/platform/outbound"
	"github.com/somralab/somra-media/internal/plugin"
)

// Protocol identifies torrent vs usenet for normalized results.
type Protocol string

const (
	ProtocolTorrent Protocol = "torrent"
	ProtocolUsenet  Protocol = "usenet"
)

// Config holds public indexer settings.
type Config struct {
	BaseURL    string `json:"baseUrl"`
	Categories []int  `json:"categories,omitempty"`
}

// Client talks to Torznab/Newznab-compatible APIs.
type Client struct {
	client   *outbound.PinnedClient
	apiKey   string
	protocol Protocol
}

// NewClient builds a client for the given protocol.
func NewClient(cfg Config, apiKey string, protocol Protocol) (*Client, error) {
	c, err := outbound.NewPinnedClient(cfg.BaseURL, 20*time.Second)
	if err != nil {
		return nil, err
	}
	return &Client{client: c, apiKey: apiKey, protocol: protocol}, nil
}

type capsResponse struct {
	XMLName xml.Name `xml:"caps"`
	Categories struct {
		Category []struct {
			ID   string `xml:"id,attr"`
			Name string `xml:"name,attr"`
		} `xml:"category"`
	} `xml:"categories"`
	Limits struct {
		Max string `xml:"max,attr"`
	} `xml:"limits"`
	Searching struct {
		Search struct {
			Available string `xml:"available,attr"`
		} `xml:"search"`
	} `xml:"searching"`
}

// Capabilities queries t=caps.
func (c *Client) Capabilities(ctx context.Context) (plugin.Capabilities, error) {
	q := url.Values{"t": {"caps"}}
	if c.apiKey != "" {
		q.Set("apikey", c.apiKey)
	}
	resp, err := c.client.Get(ctx, "/api", q)
	if err != nil {
		return defaultCaps(c.protocol), nil
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil || resp.StatusCode >= 400 {
		return defaultCaps(c.protocol), nil
	}
	var caps capsResponse
	if err := xml.Unmarshal(body, &caps); err != nil {
		return defaultCaps(c.protocol), nil
	}
	out := defaultCaps(c.protocol)
	for _, cat := range caps.Categories.Category {
		id, _ := strconv.Atoi(cat.ID)
		if id == 0 {
			continue
		}
		out.Categories = append(out.Categories, plugin.Category{ID: id, Name: cat.Name})
	}
	if max, err := strconv.Atoi(caps.Limits.Max); err == nil && max > 0 {
		out.Limits.MaxResults = max
	}
	if strings.EqualFold(caps.Searching.Search.Available, "yes") {
		out.SupportsSearch = true
	}
	return out, nil
}

type rssFeed struct {
	Channel struct {
		Items []rssItem `xml:"item"`
	} `xml:"channel"`
}

type rssItem struct {
	Title   string `xml:"title"`
	Link    string `xml:"link"`
	GUID    string `xml:"guid"`
	PubDate string `xml:"pubDate"`
	Attrs   []struct {
		Name  string `xml:"name,attr"`
		Value string `xml:"value,attr"`
	} `xml:"http://torznab.com/schemas/2015/feed attr"`
	Enclosure struct {
		Length string `xml:"length,attr"`
		URL    string `xml:"url,attr"`
	} `xml:"enclosure"`
}

// Search runs t=search|movie|tvsearch against the indexer.
func (c *Client) Search(ctx context.Context, q plugin.SearchQuery) ([]plugin.SearchResult, error) {
	params := url.Values{"t": {searchType(q)}}
	if c.apiKey != "" {
		params.Set("apikey", c.apiKey)
	}
	params.Set("q", q.Title)
	if q.Year != nil {
		params.Set("year", strconv.Itoa(*q.Year))
	}
	if q.Season != nil {
		params.Set("season", strconv.Itoa(*q.Season))
	}
	if q.Episode != nil {
		params.Set("ep", strconv.Itoa(*q.Episode))
	}
	resp, err := c.client.Get(ctx, "/api", params)
	if err != nil {
		return nil, fmt.Errorf("nzbindexer search: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 5<<20))
	if err != nil {
		return nil, fmt.Errorf("nzbindexer search read: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("nzbindexer search: http %d", resp.StatusCode)
	}
	var feed rssFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, fmt.Errorf("nzbindexer search parse: %w", err)
	}
	proto := plugin.ProtocolTorrent
	if c.protocol == ProtocolUsenet {
		proto = plugin.ProtocolUsenet
	}
	out := make([]plugin.SearchResult, 0, len(feed.Channel.Items))
	for _, item := range feed.Channel.Items {
		link := item.Link
		if link == "" {
			link = item.Enclosure.URL
		}
		if link == "" {
			link = item.GUID
		}
		size := int64(0)
		if item.Enclosure.Length != "" {
			size, _ = strconv.ParseInt(item.Enclosure.Length, 10, 64)
		}
		var seeders, peers *int
		for _, attr := range item.Attrs {
			switch strings.ToLower(attr.Name) {
			case "seeders":
				if n, err := strconv.Atoi(attr.Value); err == nil {
					seeders = &n
				}
			case "peers":
				if n, err := strconv.Atoi(attr.Value); err == nil {
					peers = &n
				}
			case "size":
				if size == 0 {
					if n, err := strconv.ParseInt(attr.Value, 10, 64); err == nil {
						size = n
					}
				}
			}
		}
		var pub *time.Time
		if item.PubDate != "" {
			if t, err := time.Parse(time.RFC1123Z, item.PubDate); err == nil {
				pub = &t
			}
		}
		out = append(out, plugin.SearchResult{
			ReleaseID: encodeReleaseID(link),
			Title:     item.Title,
			Protocol:  proto,
			SizeBytes: size,
			Seeders:   seeders,
			Peers:     peers,
			PublishAt: pub,
		})
	}
	return out, nil
}

func searchType(q plugin.SearchQuery) string {
	switch q.MediaKind {
	case plugin.MediaKindMovie:
		return "movie"
	case plugin.MediaKindTV:
		if q.Season != nil || q.Episode != nil {
			return "tvsearch"
		}
	}
	return "search"
}

func defaultCaps(p Protocol) plugin.Capabilities {
	proto := plugin.ProtocolTorrent
	if p == ProtocolUsenet {
		proto = plugin.ProtocolUsenet
	}
	return plugin.Capabilities{
		Protocols:      []plugin.Protocol{proto},
		SupportsSearch: true,
	}
}

// EncodeReleaseID exposes release id encoding for download clients.
func EncodeReleaseID(link string) string {
	return encodeReleaseID(link)
}

// DecodeReleaseID reverses EncodeReleaseID.
func DecodeReleaseID(id string) (string, error) {
	b, err := base64.RawURLEncoding.DecodeString(id)
	if err != nil {
		return "", fmt.Errorf("decode release id: %w", err)
	}
	return string(b), nil
}

func encodeReleaseID(link string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(link))
}
