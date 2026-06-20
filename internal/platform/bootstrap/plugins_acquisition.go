//go:build acquisition

package bootstrap

import (
	"fmt"

	"github.com/somralab/somra-media/internal/plugin"
	"github.com/somralab/somra-media/internal/plugin/newznab"
	"github.com/somralab/somra-media/internal/plugin/qbittorrent"
	"github.com/somralab/somra-media/internal/plugin/sabnzbd"
	"github.com/somralab/somra-media/internal/plugin/torznab"
)

func registerAcquisitionFactories(mgr *plugin.Manager) error {
	for _, f := range []plugin.Factory{
		torznab.NewFactory(),
		newznab.NewFactory(),
		qbittorrent.NewFactory(),
		sabnzbd.NewFactory(),
	} {
		if err := mgr.RegisterFactory(f); err != nil {
			return fmt.Errorf("register acquisition factory %s: %w", f.Implementation(), err)
		}
	}
	return nil
}
