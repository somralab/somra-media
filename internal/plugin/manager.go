package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
)

type runtimeEntry struct {
	record InstanceRecord
	plugin Plugin
}

// ManagerOptions configures a Manager.
type ManagerOptions struct {
	Logger *slog.Logger
}

// Manager orchestrates plugin factory registration and instance lifecycle.
type Manager struct {
	mu        sync.RWMutex
	store     Store
	logger    *slog.Logger
	factories map[PluginType]map[string]Factory
	runtime   map[int64]runtimeEntry
}

// NewManager returns a lifecycle manager backed by store.
func NewManager(store Store, opts ManagerOptions) *Manager {
	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}
	return &Manager{
		store:     store,
		logger:    logger,
		factories: make(map[PluginType]map[string]Factory),
		runtime:   make(map[int64]runtimeEntry),
	}
}

// RegisterFactory adds an implementation factory to the catalog.
func (m *Manager) RegisterFactory(f Factory) error {
	if m == nil {
		return fmt.Errorf("register plugin factory: manager is nil")
	}
	if f == nil {
		return fmt.Errorf("register plugin factory: factory is nil")
	}
	impl := f.Implementation()
	if impl == "" {
		return fmt.Errorf("register plugin factory: implementation is required")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	byType, ok := m.factories[f.Type()]
	if !ok {
		byType = make(map[string]Factory)
		m.factories[f.Type()] = byType
	}
	if _, exists := byType[impl]; exists {
		return fmt.Errorf("register plugin factory %q: %w", impl, ErrDuplicateFactory)
	}
	byType[impl] = f
	return nil
}

// Create persists a new instance and activates it when enabled.
func (m *Manager) Create(ctx context.Context, rec InstanceRecord) (int64, error) {
	if m == nil || m.store == nil {
		return 0, fmt.Errorf("create plugin instance: manager store is nil")
	}
	if _, err := m.factory(rec.PluginType, rec.Implementation); err != nil {
		return 0, err
	}
	if len(rec.Config) == 0 {
		rec.Config = json.RawMessage("{}")
	}
	id, err := m.store.Create(ctx, rec)
	if err != nil {
		return 0, fmt.Errorf("create plugin instance: %w", err)
	}
	if rec.Enabled {
		stored, err := m.store.GetByID(ctx, id)
		if err != nil {
			return id, fmt.Errorf("create plugin instance load %d: %w", id, err)
		}
		if err := m.activate(ctx, stored); err != nil {
			return id, fmt.Errorf("create plugin instance activate %d: %w", id, err)
		}
	}
	return id, nil
}

// Configure updates instance config and rebuilds the runtime adapter when enabled.
func (m *Manager) Configure(ctx context.Context, id int64, config json.RawMessage) error {
	if m == nil || m.store == nil {
		return fmt.Errorf("configure plugin instance: manager store is nil")
	}
	if len(config) == 0 {
		config = json.RawMessage("{}")
	}
	rec, err := m.store.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("configure plugin instance %d: %w", id, mapStoreNotFound(err))
	}
	if err := m.store.UpdateConfig(ctx, id, config); err != nil {
		return fmt.Errorf("configure plugin instance %d: %w", id, err)
	}
	if !rec.Enabled {
		return nil
	}
	m.deactivate(id)
	rec.Config = config
	if err := m.activate(ctx, rec); err != nil {
		return fmt.Errorf("configure plugin instance %d: %w", id, err)
	}
	return nil
}

// Enable activates a persisted instance at runtime.
func (m *Manager) Enable(ctx context.Context, id int64) error {
	if m == nil || m.store == nil {
		return fmt.Errorf("enable plugin instance: manager store is nil")
	}
	rec, err := m.store.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("enable plugin instance %d: %w", id, mapStoreNotFound(err))
	}
	if err := m.store.SetEnabled(ctx, id, true); err != nil {
		return fmt.Errorf("enable plugin instance %d: %w", id, err)
	}
	rec.Enabled = true
	if err := m.activate(ctx, rec); err != nil {
		return fmt.Errorf("enable plugin instance %d: %w", id, err)
	}
	return nil
}

// Disable deactivates an instance and persists the disabled flag.
func (m *Manager) Disable(ctx context.Context, id int64) error {
	if m == nil || m.store == nil {
		return fmt.Errorf("disable plugin instance: manager store is nil")
	}
	if _, err := m.store.GetByID(ctx, id); err != nil {
		return fmt.Errorf("disable plugin instance %d: %w", id, mapStoreNotFound(err))
	}
	m.deactivate(id)
	if err := m.store.SetEnabled(ctx, id, false); err != nil {
		return fmt.Errorf("disable plugin instance %d: %w", id, err)
	}
	return nil
}

// Get returns the persisted instance record.
func (m *Manager) Get(ctx context.Context, id int64) (InstanceRecord, error) {
	if m == nil || m.store == nil {
		return InstanceRecord{}, fmt.Errorf("get plugin instance: manager store is nil")
	}
	rec, err := m.store.GetByID(ctx, id)
	if err != nil {
		return InstanceRecord{}, mapStoreNotFound(err)
	}
	return rec, nil
}

// List returns all persisted instances.
func (m *Manager) List(ctx context.Context) ([]InstanceRecord, error) {
	if m == nil || m.store == nil {
		return nil, fmt.Errorf("list plugin instances: manager store is nil")
	}
	return m.store.List(ctx)
}

// LoadEnabled hydrates all enabled instances into the runtime registry.
func (m *Manager) LoadEnabled(ctx context.Context) error {
	if m == nil || m.store == nil {
		return fmt.Errorf("load enabled plugins: manager store is nil")
	}
	rows, err := m.store.List(ctx)
	if err != nil {
		return fmt.Errorf("load enabled plugins: %w", err)
	}
	for _, rec := range rows {
		if !rec.Enabled {
			continue
		}
		if err := m.activate(ctx, rec); err != nil {
			m.logger.Warn("load enabled plugin skipped",
				slog.Int64("id", rec.ID),
				slog.String("implementation", rec.Implementation),
				slog.Any("error", err),
			)
		}
	}
	return nil
}

// Indexer returns an enabled indexer instance by id.
func (m *Manager) Indexer(ctx context.Context, id int64) (Indexer, error) {
	entry, err := m.runtimeEntry(ctx, id)
	if err != nil {
		return nil, err
	}
	if entry.record.PluginType != PluginTypeIndexer {
		return nil, fmt.Errorf("plugin instance %d: %w", id, ErrUnsupportedCapability)
	}
	idx, ok := entry.plugin.(Indexer)
	if !ok {
		return nil, fmt.Errorf("plugin instance %d: %w", id, ErrUnsupportedCapability)
	}
	return idx, nil
}

// DownloadClient returns an enabled download client instance by id.
func (m *Manager) DownloadClient(ctx context.Context, id int64) (DownloadClient, error) {
	entry, err := m.runtimeEntry(ctx, id)
	if err != nil {
		return nil, err
	}
	if entry.record.PluginType != PluginTypeDownloadClient {
		return nil, fmt.Errorf("plugin instance %d: %w", id, ErrUnsupportedCapability)
	}
	client, ok := entry.plugin.(DownloadClient)
	if !ok {
		return nil, fmt.Errorf("plugin instance %d: %w", id, ErrUnsupportedCapability)
	}
	return client, nil
}

// EnabledIndexers returns a snapshot of active indexer adapters.
func (m *Manager) EnabledIndexers() []Indexer {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Indexer, 0, len(m.runtime))
	for _, entry := range m.runtime {
		if entry.record.PluginType != PluginTypeIndexer {
			continue
		}
		if idx, ok := entry.plugin.(Indexer); ok {
			out = append(out, idx)
		}
	}
	return out
}

// EnabledDownloadClients returns a snapshot of active download client adapters.
func (m *Manager) EnabledDownloadClients() []DownloadClient {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]DownloadClient, 0, len(m.runtime))
	for _, entry := range m.runtime {
		if entry.record.PluginType != PluginTypeDownloadClient {
			continue
		}
		if client, ok := entry.plugin.(DownloadClient); ok {
			out = append(out, client)
		}
	}
	return out
}

func (m *Manager) factory(typ PluginType, implementation string) (Factory, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	byType, ok := m.factories[typ]
	if !ok {
		return nil, fmt.Errorf("plugin factory %q for %q: %w", implementation, typ, ErrFactoryNotFound)
	}
	f, ok := byType[implementation]
	if !ok {
		return nil, fmt.Errorf("plugin factory %q for %q: %w", implementation, typ, ErrFactoryNotFound)
	}
	return f, nil
}

func (m *Manager) activate(ctx context.Context, rec InstanceRecord) error {
	f, err := m.factory(rec.PluginType, rec.Implementation)
	if err != nil {
		return err
	}
	config := rec.Config
	if len(config) == 0 {
		config = json.RawMessage("{}")
	}
	p, err := f.New(ctx, fmt.Sprintf("%d", rec.ID), config)
	if err != nil {
		return fmt.Errorf("build plugin instance %d: %w", rec.ID, err)
	}
	if err := ValidateContract(p); err != nil {
		return fmt.Errorf("validate plugin instance %d: %w", rec.ID, err)
	}
	m.mu.Lock()
	m.runtime[rec.ID] = runtimeEntry{record: rec, plugin: p}
	m.mu.Unlock()
	return nil
}

func (m *Manager) deactivate(id int64) {
	m.mu.Lock()
	delete(m.runtime, id)
	m.mu.Unlock()
}

func (m *Manager) runtimeEntry(ctx context.Context, id int64) (runtimeEntry, error) {
	if m == nil {
		return runtimeEntry{}, fmt.Errorf("plugin instance %d: %w", id, ErrPluginNotFound)
	}
	m.mu.RLock()
	entry, ok := m.runtime[id]
	m.mu.RUnlock()
	if !ok {
		rec, err := m.store.GetByID(ctx, id)
		if err != nil {
			return runtimeEntry{}, mapStoreNotFound(err)
		}
		if !rec.Enabled {
			return runtimeEntry{}, fmt.Errorf("plugin instance %d: %w", id, ErrPluginDisabled)
		}
		return runtimeEntry{}, fmt.Errorf("plugin instance %d: %w", id, ErrPluginDisabled)
	}
	return entry, nil
}

func mapStoreNotFound(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrPluginNotFound) {
		return ErrPluginNotFound
	}
	return err
}
