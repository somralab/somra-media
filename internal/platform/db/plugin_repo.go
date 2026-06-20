package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// PluginInstanceType identifies an acquisition plugin role.
type PluginInstanceType string

const (
	PluginInstanceTypeIndexer        PluginInstanceType = "indexer"
	PluginInstanceTypeDownloadClient PluginInstanceType = "download_client"
)

// PluginInstance is a configured acquisition plugin row.
type PluginInstance struct {
	ID             int64              `json:"id"`
	PluginType     PluginInstanceType `json:"pluginType"`
	Implementation string             `json:"implementation"`
	Name           string             `json:"name"`
	Config         string             `json:"config"`
	Enabled        bool               `json:"enabled"`
	CreatedAt      time.Time          `json:"createdAt"`
	UpdatedAt      time.Time          `json:"updatedAt"`
}

// PluginInstanceRepo persists plugin instance configuration.
type PluginInstanceRepo struct {
	q Querier
}

// NewPluginInstanceRepo returns a repository bound to q.
func NewPluginInstanceRepo(q Querier) *PluginInstanceRepo {
	return &PluginInstanceRepo{q: q}
}

var (
	ErrPluginInstanceNotFound  = errors.New("db plugin instance: not found")
	ErrPluginInstanceDuplicate = errors.New("db plugin instance: duplicate name")
)

// Create inserts a plugin instance.
func (r *PluginInstanceRepo) Create(ctx context.Context, inst PluginInstance) (int64, error) {
	if inst.PluginType == "" {
		return 0, fmt.Errorf("db plugin instance create: plugin type is required")
	}
	if strings.TrimSpace(inst.Implementation) == "" {
		return 0, fmt.Errorf("db plugin instance create: implementation is required")
	}
	config := inst.Config
	if config == "" {
		config = "{}"
	}
	res, err := r.q.ExecContext(ctx, `
		INSERT INTO plugin_instances (plugin_type, implementation, name, config, enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now'))
	`, inst.PluginType, inst.Implementation, nullStr(inst.Name), config, boolToInt(inst.Enabled))
	if err != nil {
		if isUniqueViolation(err) {
			return 0, fmt.Errorf("db plugin instance create: %w", ErrPluginInstanceDuplicate)
		}
		return 0, fmt.Errorf("db plugin instance create: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("db plugin instance create id: %w", err)
	}
	return id, nil
}

// GetByID returns an instance by primary key.
func (r *PluginInstanceRepo) GetByID(ctx context.Context, id int64) (PluginInstance, error) {
	var inst PluginInstance
	var enabled int
	var created, updated string
	err := r.q.QueryRowContext(ctx, `
		SELECT id, plugin_type, implementation, name, config, enabled, created_at, updated_at
		FROM plugin_instances WHERE id = ?
	`, id).Scan(&inst.ID, &inst.PluginType, &inst.Implementation, &inst.Name, &inst.Config, &enabled, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return PluginInstance{}, ErrPluginInstanceNotFound
	}
	if err != nil {
		return PluginInstance{}, fmt.Errorf("db plugin instance get: %w", err)
	}
	inst.Enabled = enabled != 0
	inst.CreatedAt = parseSQLiteTime(created)
	inst.UpdatedAt = parseSQLiteTime(updated)
	return inst, nil
}

// List returns all instances ordered by id.
func (r *PluginInstanceRepo) List(ctx context.Context) ([]PluginInstance, error) {
	rows, err := r.q.QueryContext(ctx, `
		SELECT id, plugin_type, implementation, name, config, enabled, created_at, updated_at
		FROM plugin_instances ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("db plugin instance list: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []PluginInstance
	for rows.Next() {
		var inst PluginInstance
		var enabled int
		var created, updated string
		if err := rows.Scan(&inst.ID, &inst.PluginType, &inst.Implementation, &inst.Name, &inst.Config, &enabled, &created, &updated); err != nil {
			return nil, fmt.Errorf("db plugin instance scan: %w", err)
		}
		inst.Enabled = enabled != 0
		inst.CreatedAt = parseSQLiteTime(created)
		inst.UpdatedAt = parseSQLiteTime(updated)
		out = append(out, inst)
	}
	return out, rows.Err()
}

// UpdateConfig replaces the JSON config blob.
func (r *PluginInstanceRepo) UpdateConfig(ctx context.Context, id int64, config string) error {
	if config == "" {
		config = "{}"
	}
	res, err := r.q.ExecContext(ctx, `
		UPDATE plugin_instances SET config = ?, updated_at = datetime('now') WHERE id = ?
	`, config, id)
	if err != nil {
		return fmt.Errorf("db plugin instance update config: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("db plugin instance update config rows: %w", err)
	}
	if n == 0 {
		return ErrPluginInstanceNotFound
	}
	return nil
}

// SetEnabled toggles an instance.
func (r *PluginInstanceRepo) SetEnabled(ctx context.Context, id int64, enabled bool) error {
	res, err := r.q.ExecContext(ctx, `
		UPDATE plugin_instances SET enabled = ?, updated_at = datetime('now') WHERE id = ?
	`, boolToInt(enabled), id)
	if err != nil {
		return fmt.Errorf("db plugin instance set enabled: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("db plugin instance set enabled rows: %w", err)
	}
	if n == 0 {
		return ErrPluginInstanceNotFound
	}
	return nil
}
