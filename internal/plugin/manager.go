package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	platformsecrets "github.com/somralab/somra-media/internal/platform/secrets"
)

type runtimeEntry struct {
	record InstanceRecord
	plugin Plugin
}

// ManagerOptions configures a Manager.
type ManagerOptions struct {
	Logger        *slog.Logger
	EncryptionKey string
}

// Manager orchestrates plugin factory registration and instance lifecycle.
type Manager struct {
	mu        sync.RWMutex
	store     Store
	logger    *slog.Logger
	secretKey []byte
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
		secretKey: platformsecrets.DeriveKey(opts.EncryptionKey),
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
	extra := m.secretFields(rec.PluginType, rec.Implementation)
	public, secrets, err := SplitConfig(rec.Config, extra)
	if err != nil {
		return 0, fmt.Errorf("create plugin instance split config: %w", err)
	}
	enc, err := platformsecrets.EncryptMap(m.secretKey, secrets)
	if err != nil {
		return 0, fmt.Errorf("create plugin instance encrypt secrets: %w", err)
	}
	rec.Config = public
	rec.SecretsEnc = enc

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

// Configure replaces instance config and rebuilds the runtime adapter when enabled.
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
	extra := m.secretFields(rec.PluginType, rec.Implementation)
	public, secrets, err := SplitConfig(config, extra)
	if err != nil {
		return fmt.Errorf("configure plugin instance %d split config: %w", id, err)
	}
	enc, err := platformsecrets.EncryptMap(m.secretKey, secrets)
	if err != nil {
		return fmt.Errorf("configure plugin instance %d encrypt secrets: %w", id, err)
	}
	if err := m.store.UpdateConfig(ctx, id, public, enc); err != nil {
		return fmt.Errorf("configure plugin instance %d: %w", id, err)
	}
	if !rec.Enabled {
		return nil
	}
	m.deactivate(id)
	rec.Config = public
	rec.SecretsEnc = enc
	if err := m.activate(ctx, rec); err != nil {
		return fmt.Errorf("configure plugin instance %d: %w", id, err)
	}
	return nil
}

// PatchConfig merges a config patch and rebuilds the runtime adapter when enabled.
func (m *Manager) PatchConfig(ctx context.Context, id int64, patch json.RawMessage) error {
	if m == nil || m.store == nil {
		return fmt.Errorf("patch plugin instance: manager store is nil")
	}
	rec, err := m.store.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("patch plugin instance %d: %w", id, mapStoreNotFound(err))
	}
	existingSecrets, err := m.decryptSecrets(rec.SecretsEnc)
	if err != nil {
		return fmt.Errorf("patch plugin instance %d decrypt secrets: %w", id, err)
	}
	extra := m.secretFields(rec.PluginType, rec.Implementation)
	public, secrets, err := MergeConfigPatch(rec.Config, existingSecrets, patch, extra)
	if err != nil {
		return fmt.Errorf("patch plugin instance %d merge config: %w", id, err)
	}
	enc, err := platformsecrets.EncryptMap(m.secretKey, secrets)
	if err != nil {
		return fmt.Errorf("patch plugin instance %d encrypt secrets: %w", id, err)
	}
	if err := m.store.UpdateConfig(ctx, id, public, enc); err != nil {
		return fmt.Errorf("patch plugin instance %d: %w", id, err)
	}
	if !rec.Enabled {
		return nil
	}
	m.deactivate(id)
	rec.Config = public
	rec.SecretsEnc = enc
	if err := m.activate(ctx, rec); err != nil {
		return fmt.Errorf("patch plugin instance %d: %w", id, err)
	}
	return nil
}

// UpdateName changes the display name of a persisted instance.
func (m *Manager) UpdateName(ctx context.Context, id int64, name string) error {
	if m == nil || m.store == nil {
		return fmt.Errorf("update plugin instance name: manager store is nil")
	}
	if _, err := m.store.GetByID(ctx, id); err != nil {
		return fmt.Errorf("update plugin instance name %d: %w", id, mapStoreNotFound(err))
	}
	if err := m.store.UpdateName(ctx, id, name); err != nil {
		return fmt.Errorf("update plugin instance name %d: %w", id, err)
	}
	m.mu.Lock()
	if entry, ok := m.runtime[id]; ok {
		entry.record.Name = name
		m.runtime[id] = entry
	}
	m.mu.Unlock()
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

// Delete removes a persisted instance and deactivates it at runtime.
func (m *Manager) Delete(ctx context.Context, id int64) error {
	if m == nil || m.store == nil {
		return fmt.Errorf("delete plugin instance: manager store is nil")
	}
	if _, err := m.store.GetByID(ctx, id); err != nil {
		return fmt.Errorf("delete plugin instance %d: %w", id, mapStoreNotFound(err))
	}
	m.deactivate(id)
	if err := m.store.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete plugin instance %d: %w", id, err)
	}
	return nil
}

// Get returns the persisted instance record (public config only; secrets remain encrypted).
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

// PublicConfig returns an API-safe config map with secret *Set flags for an instance.
func (m *Manager) PublicConfig(ctx context.Context, id int64) (map[string]any, error) {
	rec, err := m.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	secrets, err := m.decryptSecrets(rec.SecretsEnc)
	if err != nil {
		return nil, fmt.Errorf("public config instance %d: %w", id, err)
	}
	return RedactConfig(rec.Config, secrets)
}

// List returns all persisted instances.
func (m *Manager) List(ctx context.Context) ([]InstanceRecord, error) {
	if m == nil || m.store == nil {
		return nil, fmt.Errorf("list plugin instances: manager store is nil")
	}
	return m.store.List(ctx)
}

// CatalogEntry describes a registered plugin factory.
type CatalogEntry struct {
	PluginType      PluginType `json:"pluginType"`
	Implementation  string     `json:"implementation"`
	ContractVersion string     `json:"contractVersion"`
}

// Catalog returns registered plugin factories.
func (m *Manager) Catalog() []CatalogEntry {
	if m == nil {
		return nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []CatalogEntry
	for typ, byImpl := range m.factories {
		for impl := range byImpl {
			out = append(out, CatalogEntry{
				PluginType:      typ,
				Implementation:  impl,
				ContractVersion: ContractVersion,
			})
		}
	}
	return out
}

// TestResult is the outcome of a plugin connection test.
type TestResult struct {
	Success    bool           `json:"success"`
	MessageKey string         `json:"messageKey"`
	Details    map[string]any `json:"details,omitempty"`
}

// Test verifies connectivity for a plugin instance.
func (m *Manager) Test(ctx context.Context, id int64) (TestResult, error) {
	if m == nil || m.store == nil {
		return TestResult{}, fmt.Errorf("test plugin instance: manager store is nil")
	}
	rec, err := m.store.GetByID(ctx, id)
	if err != nil {
		return TestResult{}, mapStoreNotFound(err)
	}
	f, err := m.factory(rec.PluginType, rec.Implementation)
	if err != nil {
		return TestResult{}, err
	}
	merged, err := m.activationConfig(rec)
	if err != nil {
		return TestResult{}, fmt.Errorf("test plugin instance %d: %w", id, err)
	}
	p, err := f.New(ctx, fmt.Sprintf("%d", rec.ID), merged)
	if err != nil {
		return TestResult{
			Success:    false,
			MessageKey: "plugins.instances.test.failed",
			Details:    map[string]any{"error": err.Error()},
		}, nil
	}
	if err := ValidateContract(p); err != nil {
		return TestResult{
			Success:    false,
			MessageKey: "plugins.instances.test.failed",
			Details:    map[string]any{"error": err.Error()},
		}, nil
	}
	switch rec.PluginType {
	case PluginTypeIndexer:
		idx, ok := p.(Indexer)
		if !ok {
			return TestResult{Success: false, MessageKey: "plugins.instances.test.failed"}, nil
		}
		caps, err := idx.Capabilities(ctx)
		if err != nil {
			return TestResult{
				Success:    false,
				MessageKey: "plugins.instances.test.failed",
				Details:    map[string]any{"error": err.Error()},
			}, nil
		}
		return TestResult{
			Success:    true,
			MessageKey: "plugins.instances.test.success",
			Details:    map[string]any{"supportsSearch": caps.SupportsSearch, "protocols": caps.Protocols},
		}, nil
	case PluginTypeDownloadClient:
		client, ok := p.(DownloadClient)
		if !ok {
			return TestResult{Success: false, MessageKey: "plugins.instances.test.failed"}, nil
		}
		item, err := client.Add(ctx, AddRequest{ReleaseID: "__somra_test__"})
		if err != nil {
			return TestResult{
				Success:    false,
				MessageKey: "plugins.instances.test.failed",
				Details:    map[string]any{"error": err.Error()},
			}, nil
		}
		if _, err := client.Status(ctx, item.DownloadID); err != nil {
			return TestResult{
				Success:    false,
				MessageKey: "plugins.instances.test.failed",
				Details:    map[string]any{"error": err.Error()},
			}, nil
		}
		return TestResult{Success: true, MessageKey: "plugins.instances.test.success"}, nil
	default:
		return TestResult{Success: false, MessageKey: "plugins.instances.test.failed"}, nil
	}
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

func (m *Manager) secretFields(typ PluginType, implementation string) []string {
	f, err := m.factory(typ, implementation)
	if err != nil {
		return nil
	}
	if p, ok := f.(SecretFieldProvider); ok {
		return p.SecretFields()
	}
	return nil
}

func (m *Manager) decryptSecrets(enc string) (map[string]string, error) {
	return platformsecrets.DecryptMap(m.secretKey, enc)
}

func (m *Manager) activationConfig(rec InstanceRecord) ([]byte, error) {
	secrets, err := m.decryptSecrets(rec.SecretsEnc)
	if err != nil {
		return nil, err
	}
	return MergeConfig(rec.Config, secrets)
}

func (m *Manager) activate(ctx context.Context, rec InstanceRecord) error {
	f, err := m.factory(rec.PluginType, rec.Implementation)
	if err != nil {
		return err
	}
	config, err := m.activationConfig(rec)
	if err != nil {
		return fmt.Errorf("build plugin instance %d config: %w", rec.ID, err)
	}
	if len(config) == 0 {
		config = []byte("{}")
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
