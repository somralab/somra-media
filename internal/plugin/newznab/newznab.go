//go:build acquisition

package newznab

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/somralab/somra-media/internal/plugin"
	"github.com/somralab/somra-media/internal/plugin/nzbindexer"
)

const Implementation = "newznab"

// Factory builds Newznab indexer plugins.
type Factory struct{}

// NewFactory returns a Newznab factory.
func NewFactory() Factory { return Factory{} }

func (Factory) Implementation() string { return Implementation }
func (Factory) Type() plugin.PluginType {
	return plugin.PluginTypeIndexer
}

func (Factory) New(_ context.Context, instanceID string, config []byte) (plugin.Plugin, error) {
	var cfg nzbindexer.Config
	if len(config) > 0 {
		if err := json.Unmarshal(config, &cfg); err != nil {
			return nil, fmt.Errorf("newznab config: %w", err)
		}
	}
	apiKey, _ := configSecret(config, "apiKey")
	client, err := nzbindexer.NewClient(cfg, apiKey, nzbindexer.ProtocolUsenet)
	if err != nil {
		return nil, err
	}
	return &indexer{id: instanceID, client: client}, nil
}

type indexer struct {
	id     string
	client *nzbindexer.Client
}

func (i *indexer) ID() string              { return i.id }
func (i *indexer) Type() plugin.PluginType { return plugin.PluginTypeIndexer }
func (i *indexer) ContractVersion() string {
	return plugin.ContractVersion
}

func (i *indexer) Capabilities(ctx context.Context) (plugin.Capabilities, error) {
	return i.client.Capabilities(ctx)
}

func (i *indexer) Search(ctx context.Context, q plugin.SearchQuery) ([]plugin.SearchResult, error) {
	results, err := i.client.Search(ctx, q)
	if err != nil {
		return nil, err
	}
	for idx := range results {
		results[idx].IndexerID = i.id
	}
	return results, nil
}

var _ plugin.Indexer = (*indexer)(nil)

func configSecret(config []byte, key string) (string, bool) {
	var m map[string]any
	if err := json.Unmarshal(config, &m); err != nil {
		return "", false
	}
	v, ok := m[key].(string)
	return v, ok
}
