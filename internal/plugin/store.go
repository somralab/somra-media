package plugin

import (
	"context"
	"encoding/json"
)

// Store persists plugin instance configuration.
type Store interface {
	Create(ctx context.Context, rec InstanceRecord) (int64, error)
	UpdateConfig(ctx context.Context, id int64, config json.RawMessage, secretsEnc string) error
	UpdateName(ctx context.Context, id int64, name string) error
	SetEnabled(ctx context.Context, id int64, enabled bool) error
	GetByID(ctx context.Context, id int64) (InstanceRecord, error)
	List(ctx context.Context) ([]InstanceRecord, error)
	Delete(ctx context.Context, id int64) error
}
