package plugin

// ContractVersion is the plugin contract generation supported by the core.
const ContractVersion = "1"

// PluginType identifies the acquisition role of a plugin adapter.
type PluginType string

const (
	PluginTypeIndexer        PluginType = "indexer"
	PluginTypeDownloadClient PluginType = "download_client"
)

// Plugin is the common identity contract for every acquisition adapter.
type Plugin interface {
	ID() string
	Type() PluginType
	ContractVersion() string
}
