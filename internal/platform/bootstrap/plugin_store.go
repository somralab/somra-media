package bootstrap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/somralab/somra-media/internal/platform/db"
	"github.com/somralab/somra-media/internal/plugin"
)

type pluginStore struct {
	repo *db.PluginInstanceRepo
}

func newPluginStore(repo *db.PluginInstanceRepo) plugin.Store {
	return &pluginStore{repo: repo}
}

func (s *pluginStore) Create(ctx context.Context, rec plugin.InstanceRecord) (int64, error) {
	config := string(rec.Config)
	if config == "" {
		config = "{}"
	}
	id, err := s.repo.Create(ctx, db.PluginInstance{
		PluginType:     db.PluginInstanceType(rec.PluginType),
		Implementation: rec.Implementation,
		Name:           rec.Name,
		Config:         config,
		Enabled:        rec.Enabled,
	})
	if err != nil {
		if errors.Is(err, db.ErrPluginInstanceDuplicate) {
			return 0, fmt.Errorf("create plugin instance: %w", plugin.ErrDuplicateInstance)
		}
		return 0, err
	}
	return id, nil
}

func (s *pluginStore) UpdateConfig(ctx context.Context, id int64, config json.RawMessage) error {
	raw := string(config)
	if raw == "" {
		raw = "{}"
	}
	return s.repo.UpdateConfig(ctx, id, raw)
}

func (s *pluginStore) SetEnabled(ctx context.Context, id int64, enabled bool) error {
	return s.repo.SetEnabled(ctx, id, enabled)
}

func (s *pluginStore) GetByID(ctx context.Context, id int64) (plugin.InstanceRecord, error) {
	inst, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, db.ErrPluginInstanceNotFound) {
			return plugin.InstanceRecord{}, plugin.ErrPluginNotFound
		}
		return plugin.InstanceRecord{}, err
	}
	return toInstanceRecord(inst), nil
}

func (s *pluginStore) List(ctx context.Context) ([]plugin.InstanceRecord, error) {
	rows, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]plugin.InstanceRecord, 0, len(rows))
	for _, row := range rows {
		out = append(out, toInstanceRecord(row))
	}
	return out, nil
}

func toInstanceRecord(inst db.PluginInstance) plugin.InstanceRecord {
	config := json.RawMessage(inst.Config)
	if len(config) == 0 {
		config = json.RawMessage("{}")
	}
	return plugin.InstanceRecord{
		ID:             inst.ID,
		PluginType:     plugin.PluginType(inst.PluginType),
		Implementation: inst.Implementation,
		Name:           inst.Name,
		Config:         config,
		Enabled:        inst.Enabled,
		CreatedAt:      inst.CreatedAt,
		UpdatedAt:      inst.UpdatedAt,
	}
}
