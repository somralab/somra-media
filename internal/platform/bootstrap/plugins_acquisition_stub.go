//go:build !acquisition

package bootstrap

import "github.com/somralab/somra-media/internal/plugin"

func registerAcquisitionFactories(_ *plugin.Manager) error {
	return nil
}
