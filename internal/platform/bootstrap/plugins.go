package bootstrap

import (
	"context"
	"fmt"

	"github.com/somralab/somra-media/internal/platform/db"
	"github.com/somralab/somra-media/internal/plugin"
	"github.com/somralab/somra-media/internal/plugin/stub"
)

// PluginsBundle groups Sprint 09 plugin lifecycle dependencies.
type PluginsBundle struct {
	Manager *plugin.Manager
}

// WirePlugins constructs the plugin lifecycle manager and hydrates enabled instances.
func WirePlugins(c *Components, encryptionKey string) (*PluginsBundle, error) {
	if c == nil || c.DB == nil {
		return nil, fmt.Errorf("bootstrap plugins: db required")
	}
	repo := db.NewPluginInstanceRepo(c.DB.Querier())
	mgr := plugin.NewManager(newPluginStore(repo), plugin.ManagerOptions{
		Logger:        c.Logger,
		EncryptionKey: encryptionKey,
	})

	for _, f := range []plugin.Factory{
		stub.NewIndexerFactory(),
		stub.NewDownloadClientFactory(),
	} {
		if err := mgr.RegisterFactory(f); err != nil {
			return nil, fmt.Errorf("bootstrap plugins: register %s: %w", f.Implementation(), err)
		}
	}
	if err := registerAcquisitionFactories(mgr); err != nil {
		return nil, fmt.Errorf("bootstrap plugins: %w", err)
	}

	if err := mgr.LoadEnabled(context.Background()); err != nil {
		return nil, fmt.Errorf("bootstrap plugins: load enabled: %w", err)
	}

	return &PluginsBundle{Manager: mgr}, nil
}
