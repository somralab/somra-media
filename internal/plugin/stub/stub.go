package stub

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/somralab/somra-media/internal/plugin"
)

const Implementation = "stub"

type stubConfig struct {
	Prefix string `json:"prefix"`
}

// IndexerFactory builds stub indexers for lifecycle tests.
type IndexerFactory struct{}

// NewIndexerFactory returns a stub indexer factory.
func NewIndexerFactory() IndexerFactory { return IndexerFactory{} }

func (IndexerFactory) Implementation() string { return Implementation }
func (IndexerFactory) Type() plugin.PluginType {
	return plugin.PluginTypeIndexer
}

func (IndexerFactory) New(_ context.Context, instanceID string, config []byte) (plugin.Plugin, error) {
	cfg := stubConfig{Prefix: "stub-indexer"}
	if len(config) > 0 {
		if err := json.Unmarshal(config, &cfg); err != nil {
			return nil, fmt.Errorf("stub indexer config: %w", err)
		}
	}
	return &stubIndexer{id: instanceID, prefix: cfg.Prefix}, nil
}

type stubIndexer struct {
	id     string
	prefix string
}

func (s *stubIndexer) ID() string              { return s.id }
func (s *stubIndexer) Type() plugin.PluginType { return plugin.PluginTypeIndexer }
func (s *stubIndexer) ContractVersion() string { return plugin.ContractVersion }

func (s *stubIndexer) Capabilities(_ context.Context) (plugin.Capabilities, error) {
	return plugin.Capabilities{
		Protocols:      []plugin.Protocol{plugin.ProtocolTorrent},
		SupportsSearch: true,
	}, nil
}

func (s *stubIndexer) Search(_ context.Context, q plugin.SearchQuery) ([]plugin.SearchResult, error) {
	title := q.Title
	if s.prefix != "" {
		title = s.prefix + ":" + title
	}
	return []plugin.SearchResult{{
		ReleaseID: s.id + "-release",
		IndexerID: s.id,
		Title:     title,
		Protocol:  plugin.ProtocolTorrent,
		SizeBytes: 1,
	}}, nil
}

var _ plugin.Indexer = (*stubIndexer)(nil)

// DownloadClientFactory builds stub download clients for lifecycle tests.
type DownloadClientFactory struct{}

// NewDownloadClientFactory returns a stub download client factory.
func NewDownloadClientFactory() DownloadClientFactory { return DownloadClientFactory{} }

func (DownloadClientFactory) Implementation() string { return Implementation }
func (DownloadClientFactory) Type() plugin.PluginType {
	return plugin.PluginTypeDownloadClient
}

func (DownloadClientFactory) New(_ context.Context, instanceID string, config []byte) (plugin.Plugin, error) {
	cfg := stubConfig{Prefix: "stub-client"}
	if len(config) > 0 {
		if err := json.Unmarshal(config, &cfg); err != nil {
			return nil, fmt.Errorf("stub download client config: %w", err)
		}
	}
	return &stubDownloadClient{id: instanceID, prefix: cfg.Prefix}, nil
}

type stubDownloadClient struct {
	id     string
	prefix string
	items  map[string]plugin.DownloadItem
}

func (s *stubDownloadClient) ID() string              { return s.id }
func (s *stubDownloadClient) Type() plugin.PluginType { return plugin.PluginTypeDownloadClient }
func (s *stubDownloadClient) ContractVersion() string {
	return plugin.ContractVersion
}

func (s *stubDownloadClient) Add(_ context.Context, req plugin.AddRequest) (plugin.DownloadItem, error) {
	if s.items == nil {
		s.items = make(map[string]plugin.DownloadItem)
	}
	downloadID := s.id + "-dl"
	item := plugin.DownloadItem{
		DownloadID: downloadID,
		ClientID:   s.id,
		ReleaseID:  req.ReleaseID,
		Status:     plugin.DownloadStatusQueued,
	}
	if s.prefix != "" {
		item.SavePath = s.prefix + "/" + req.ReleaseID
	}
	s.items[downloadID] = item
	return item, nil
}

func (s *stubDownloadClient) Status(_ context.Context, downloadID string) (plugin.DownloadItem, error) {
	item, ok := s.items[downloadID]
	if !ok {
		return plugin.DownloadItem{}, plugin.ErrUnsupportedCapability
	}
	return item, nil
}

var _ plugin.DownloadClient = (*stubDownloadClient)(nil)
