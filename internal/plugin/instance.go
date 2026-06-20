package plugin

import (
	"encoding/json"
	"time"
)

// InstanceRecord is the persisted configuration for one plugin instance.
type InstanceRecord struct {
	ID             int64           `json:"id"`
	PluginType     PluginType      `json:"pluginType"`
	Implementation string          `json:"implementation"`
	Name           string          `json:"name"`
	Config         json.RawMessage `json:"config"`
	Enabled        bool            `json:"enabled"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt"`
}
