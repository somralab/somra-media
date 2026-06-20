package plugin

import "context"

// Factory builds a plugin adapter from persisted instance configuration.
type Factory interface {
	Implementation() string
	Type() PluginType
	New(ctx context.Context, instanceID string, config []byte) (Plugin, error)
}
