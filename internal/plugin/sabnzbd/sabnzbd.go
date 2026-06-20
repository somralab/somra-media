//go:build acquisition

package sabnzbd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/somralab/somra-media/internal/platform/outbound"
	"github.com/somralab/somra-media/internal/plugin"
	"github.com/somralab/somra-media/internal/plugin/nzbindexer"
)

const Implementation = "sabnzbd"

type clientConfig struct {
	BaseURL  string `json:"baseUrl"`
	Category string `json:"category,omitempty"`
	APIKey   string `json:"apiKey,omitempty"`
}

// Factory builds SABnzbd download client plugins.
type Factory struct{}

// NewFactory returns a SABnzbd factory.
func NewFactory() Factory { return Factory{} }

func (Factory) Implementation() string { return Implementation }
func (Factory) Type() plugin.PluginType {
	return plugin.PluginTypeDownloadClient
}

func (Factory) New(_ context.Context, instanceID string, config []byte) (plugin.Plugin, error) {
	var cfg clientConfig
	if len(config) > 0 {
		if err := json.Unmarshal(config, &cfg); err != nil {
			return nil, fmt.Errorf("sabnzbd config: %w", err)
		}
	}
	httpClient, err := outbound.NewPinnedClient(cfg.BaseURL, 20*time.Second)
	if err != nil {
		return nil, err
	}
	return &client{
		id:   instanceID,
		cfg:  cfg,
		http: httpClient,
		nzo:  make(map[string]string),
	}, nil
}

type client struct {
	id   string
	cfg  clientConfig
	http *outbound.PinnedClient
	nzo  map[string]string
}

func (c *client) ID() string              { return c.id }
func (c *client) Type() plugin.PluginType { return plugin.PluginTypeDownloadClient }
func (c *client) ContractVersion() string {
	return plugin.ContractVersion
}

func (c *client) Add(ctx context.Context, req plugin.AddRequest) (plugin.DownloadItem, error) {
	link, err := nzbindexer.DecodeReleaseID(req.ReleaseID)
	if err != nil {
		link = req.ReleaseID
	}
	q := url.Values{
		"mode":   {"addurl"},
		"name":   {link},
		"apikey": {c.cfg.APIKey},
	}
	if c.cfg.Category != "" {
		q.Set("cat", c.cfg.Category)
	}
	resp, err := c.http.Get(ctx, "/api", q)
	if err != nil {
		return plugin.DownloadItem{}, fmt.Errorf("sabnzbd add: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return plugin.DownloadItem{}, err
	}
	if resp.StatusCode >= 400 {
		return plugin.DownloadItem{}, fmt.Errorf("sabnzbd add: http %d", resp.StatusCode)
	}
	var out struct {
		Status bool   `json:"status"`
		NZOIDs []string `json:"nzo_ids"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return plugin.DownloadItem{}, err
	}
	if !out.Status || len(out.NZOIDs) == 0 {
		return plugin.DownloadItem{}, fmt.Errorf("sabnzbd add: rejected")
	}
	nzo := out.NZOIDs[0]
	downloadID := c.id + "-" + nzo
	c.nzo[downloadID] = nzo
	return plugin.DownloadItem{
		DownloadID: downloadID,
		ClientID:   c.id,
		ReleaseID:  req.ReleaseID,
		Status:     plugin.DownloadStatusQueued,
	}, nil
}

func (c *client) Status(ctx context.Context, downloadID string) (plugin.DownloadItem, error) {
	nzo, ok := c.nzo[downloadID]
	if !ok {
		return plugin.DownloadItem{}, plugin.ErrUnsupportedCapability
	}
	q := url.Values{
		"mode":   {"queue"},
		"apikey": {c.cfg.APIKey},
	}
	resp, err := c.http.Get(ctx, "/api", q)
	if err != nil {
		return plugin.DownloadItem{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return plugin.DownloadItem{}, err
	}
	var queue struct {
		Queue struct {
			Slots []struct {
				NZO_ID      string  `json:"nzo_id"`
				Filename    string  `json:"filename"`
				Percentage  string  `json:"percentage"`
				Status      string  `json:"status"`
				MB          string  `json:"mb"`
				MBLeft      string  `json:"mbleft"`
				Storage     string  `json:"storage"`
				Completed   bool    `json:"completed"`
				Progress    float64 `json:"-"`
			} `json:"slots"`
		} `json:"queue"`
	}
	if err := json.Unmarshal(body, &queue); err != nil {
		return plugin.DownloadItem{}, err
	}
	for _, slot := range queue.Queue.Slots {
		if slot.NZO_ID != nzo {
			continue
		}
		progress := parsePercent(slot.Percentage)
		status := plugin.DownloadStatusDownloading
		if slot.Completed || strings.EqualFold(slot.Status, "Completed") {
			status = plugin.DownloadStatusCompleted
		}
		item := plugin.DownloadItem{
			DownloadID: downloadID,
			ClientID:   c.id,
			Status:     status,
			Progress:   progress,
			SavePath:   slot.Storage,
		}
		if status == plugin.DownloadStatusCompleted {
			now := time.Now().UTC()
			item.CompletedAt = &now
			item.Progress = 1
		}
		return item, nil
	}
	return plugin.DownloadItem{
		DownloadID: downloadID,
		ClientID:   c.id,
		Status:     plugin.DownloadStatusCompleted,
		Progress:   1,
	}, nil
}

func parsePercent(raw string) float64 {
	raw = strings.TrimSpace(strings.TrimSuffix(raw, "%"))
	if raw == "" {
		return 0
	}
	var v float64
	fmt.Sscanf(raw, "%f", &v)
	return v / 100
}

var _ plugin.DownloadClient = (*client)(nil)
