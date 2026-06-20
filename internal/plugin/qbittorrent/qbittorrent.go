//go:build acquisition

package qbittorrent

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
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

const Implementation = "qbittorrent"

type clientConfig struct {
	BaseURL  string `json:"baseUrl"`
	Category string `json:"category,omitempty"`
	SavePath string `json:"savePath,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// Factory builds qBittorrent download client plugins.
type Factory struct{}

// NewFactory returns a qBittorrent factory.
func NewFactory() Factory { return Factory{} }

func (Factory) Implementation() string { return Implementation }
func (Factory) Type() plugin.PluginType {
	return plugin.PluginTypeDownloadClient
}

func (Factory) New(_ context.Context, instanceID string, config []byte) (plugin.Plugin, error) {
	var cfg clientConfig
	if len(config) > 0 {
		if err := json.Unmarshal(config, &cfg); err != nil {
			return nil, fmt.Errorf("qbittorrent config: %w", err)
		}
	}
	httpClient, err := outbound.NewPinnedClient(cfg.BaseURL, 20*time.Second)
	if err != nil {
		return nil, err
	}
	return &client{
		id:     instanceID,
		cfg:    cfg,
		http:   httpClient,
		hashes: make(map[string]string),
	}, nil
}

type client struct {
	id     string
	cfg    clientConfig
	http   *outbound.PinnedClient
	hashes map[string]string // downloadID -> infohash
}

func (c *client) ID() string              { return c.id }
func (c *client) Type() plugin.PluginType { return plugin.PluginTypeDownloadClient }
func (c *client) ContractVersion() string {
	return plugin.ContractVersion
}

func (c *client) Add(ctx context.Context, req plugin.AddRequest) (plugin.DownloadItem, error) {
	if err := c.login(ctx); err != nil {
		return plugin.DownloadItem{}, err
	}
	link, err := nzbindexer.DecodeReleaseID(req.ReleaseID)
	if err != nil {
		link = req.ReleaseID
	}
	form := url.Values{"urls": {link}}
	if c.cfg.Category != "" {
		form.Set("category", c.cfg.Category)
	}
	if c.cfg.SavePath != "" {
		form.Set("savepath", c.cfg.SavePath)
	}
	resp, err := c.http.PostForm(ctx, "/api/v2/torrents/add", form)
	if err != nil {
		return plugin.DownloadItem{}, fmt.Errorf("qbittorrent add: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return plugin.DownloadItem{}, fmt.Errorf("qbittorrent add: http %d: %s", resp.StatusCode, body)
	}
	hash := infoHashFromLink(link)
	downloadID := c.id + "-" + hash[:8]
	c.hashes[downloadID] = hash
	return plugin.DownloadItem{
		DownloadID: downloadID,
		ClientID:   c.id,
		ReleaseID:  req.ReleaseID,
		Status:     plugin.DownloadStatusQueued,
		Progress:   0,
		SavePath:   c.cfg.SavePath,
	}, nil
}

func (c *client) Status(ctx context.Context, downloadID string) (plugin.DownloadItem, error) {
	hash, ok := c.hashes[downloadID]
	if !ok {
		return plugin.DownloadItem{}, plugin.ErrUnsupportedCapability
	}
	if err := c.login(ctx); err != nil {
		return plugin.DownloadItem{}, err
	}
	q := url.Values{"hashes": {hash}}
	resp, err := c.http.Get(ctx, "/api/v2/torrents/info", q)
	if err != nil {
		return plugin.DownloadItem{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return plugin.DownloadItem{}, err
	}
	var rows []struct {
		State        string  `json:"state"`
		Progress     float64 `json:"progress"`
		Size         int64   `json:"size"`
		Completed    int64   `json:"completed"`
		SavePath     string  `json:"save_path"`
		Category     string  `json:"category"`
		Name         string  `json:"name"`
		DownloadPath string  `json:"download_path"`
	}
	if err := json.Unmarshal(body, &rows); err != nil {
		return plugin.DownloadItem{}, err
	}
	if len(rows) == 0 {
		return plugin.DownloadItem{
			DownloadID: downloadID,
			ClientID:   c.id,
			Status:     plugin.DownloadStatusQueued,
		}, nil
	}
	row := rows[0]
	item := plugin.DownloadItem{
		DownloadID:      downloadID,
		ClientID:        c.id,
		Status:          mapState(row.State),
		Progress:        row.Progress,
		TotalBytes:      row.Size,
		DownloadedBytes: row.Completed,
		SavePath:        row.SavePath,
	}
	if item.Status == plugin.DownloadStatusCompleted {
		now := time.Now().UTC()
		item.CompletedAt = &now
	}
	return item, nil
}

func (c *client) login(ctx context.Context) error {
	if c.cfg.Username == "" {
		return nil
	}
	form := url.Values{
		"username": {c.cfg.Username},
		"password": {c.cfg.Password},
	}
	resp, err := c.http.PostForm(ctx, "/api/v2/auth/login", form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("qbittorrent login: http %d", resp.StatusCode)
	}
	return nil
}

func mapState(state string) plugin.DownloadStatus {
	switch strings.ToLower(state) {
	case "downloading", "stalleddl", "forceddl", "metadl", "allocating":
		return plugin.DownloadStatusDownloading
	case "pauseddl", "pausedup":
		return plugin.DownloadStatusPaused
	case "uploading", "stalledup", "forcedup", "queuedup":
		return plugin.DownloadStatusCompleted
	case "error", "missingfiles":
		return plugin.DownloadStatusFailed
	default:
		if strings.Contains(strings.ToLower(state), "up") {
			return plugin.DownloadStatusCompleted
		}
		return plugin.DownloadStatusQueued
	}
}

func infoHashFromLink(link string) string {
	if strings.HasPrefix(link, "magnet:") {
		if idx := strings.Index(link, "btih:"); idx >= 0 {
			raw := link[idx+5:]
			if end := strings.IndexAny(raw, "&"); end >= 0 {
				raw = raw[:end]
			}
			return strings.ToLower(raw)
		}
	}
	sum := sha1.Sum([]byte(link))
	return hex.EncodeToString(sum[:])
}

var _ plugin.DownloadClient = (*client)(nil)
