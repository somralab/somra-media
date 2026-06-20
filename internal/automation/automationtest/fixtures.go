package automationtest

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/platform/db"
	"github.com/somralab/somra-media/internal/plugin"
	"github.com/somralab/somra-media/internal/plugin/stub"
)

// OpenDB returns an initialized test database.
func OpenDB(t *testing.T) *db.DB {
	t.Helper()
	cfg := db.Default()
	cfg.DataDir = t.TempDir()
	d, err := db.Initialize(context.Background(), cfg, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = d.Close() })
	return d
}

// NewManager registers stub factories and returns a plugin manager.
func NewManager(t *testing.T, d *db.DB) *plugin.Manager {
	t.Helper()
	mgr := plugin.NewManager(newPluginStore(d), plugin.ManagerOptions{EncryptionKey: "test-jwt-secret"})
	require.NoError(t, mgr.RegisterFactory(stub.NewIndexerFactory()))
	require.NoError(t, mgr.RegisterFactory(stub.NewDownloadClientFactory()))
	return mgr
}

// CreateStubIndexer enables a stub indexer instance.
func CreateStubIndexer(t *testing.T, mgr *plugin.Manager, name string) int64 {
	t.Helper()
	id, err := mgr.Create(context.Background(), plugin.InstanceRecord{
		PluginType:     plugin.PluginTypeIndexer,
		Implementation: stub.Implementation,
		Name:           name,
		Config:         []byte("{}"),
		Enabled:        true,
	})
	require.NoError(t, err)
	return id
}

// CreateStubDownloadClient enables a stub download client instance.
func CreateStubDownloadClient(t *testing.T, mgr *plugin.Manager, name string) int64 {
	t.Helper()
	id, err := mgr.Create(context.Background(), plugin.InstanceRecord{
		PluginType:     plugin.PluginTypeDownloadClient,
		Implementation: stub.Implementation,
		Name:           name,
		Config:         []byte("{}"),
		Enabled:        true,
	})
	require.NoError(t, err)
	return id
}

// CreateApprovedRequest inserts an approved media request for automation tests.
func CreateApprovedRequest(t *testing.T, d *db.DB, title string) int64 {
	t.Helper()
	ctx := context.Background()
	users := db.NewUserRepo(d.Querier())
	userID := uuid.NewString()
	_, err := users.Create(ctx, userID, "auto-"+uuid.NewString()[:8], "hash", []string{"user"})
	require.NoError(t, err)

	id, err := db.NewRequestRepo(d.Querier()).Create(ctx, db.Request{
		UserID:            userID,
		MediaKind:         db.RequestMediaKindMovie,
		Provider:          "tmdb",
		ExternalID:        uuid.NewString(),
		Title:             title,
		QualityResolution: db.RequestQualityAny,
		Status:            db.RequestStatusApproved,
	})
	require.NoError(t, err)
	return id
}

type pluginStore struct {
	repo *db.PluginInstanceRepo
}

func newPluginStore(d *db.DB) plugin.Store {
	return &pluginStore{repo: db.NewPluginInstanceRepo(d.Querier())}
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
		SecretsEnc:     rec.SecretsEnc,
		Enabled:        rec.Enabled,
	})
	if err != nil {
		if errors.Is(err, db.ErrPluginInstanceDuplicate) {
			return 0, plugin.ErrDuplicateInstance
		}
		return 0, err
	}
	return id, nil
}

func (s *pluginStore) UpdateConfig(ctx context.Context, id int64, config json.RawMessage, secretsEnc string) error {
	raw := string(config)
	if raw == "" {
		raw = "{}"
	}
	return s.repo.UpdateConfig(ctx, id, raw, secretsEnc)
}

func (s *pluginStore) UpdateName(ctx context.Context, id int64, name string) error {
	if err := s.repo.UpdateName(ctx, id, name); err != nil {
		if errors.Is(err, db.ErrPluginInstanceDuplicate) {
			return plugin.ErrDuplicateInstance
		}
		if errors.Is(err, db.ErrPluginInstanceNotFound) {
			return plugin.ErrPluginNotFound
		}
		return err
	}
	return nil
}

func (s *pluginStore) SetEnabled(ctx context.Context, id int64, enabled bool) error {
	if err := s.repo.SetEnabled(ctx, id, enabled); err != nil {
		if errors.Is(err, db.ErrPluginInstanceNotFound) {
			return plugin.ErrPluginNotFound
		}
		return err
	}
	return nil
}

func (s *pluginStore) GetByID(ctx context.Context, id int64) (plugin.InstanceRecord, error) {
	inst, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, db.ErrPluginInstanceNotFound) {
			return plugin.InstanceRecord{}, plugin.ErrPluginNotFound
		}
		return plugin.InstanceRecord{}, err
	}
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
		SecretsEnc:     inst.SecretsEnc,
		Enabled:        inst.Enabled,
		CreatedAt:      inst.CreatedAt,
		UpdatedAt:      inst.UpdatedAt,
	}, nil
}

func (s *pluginStore) List(ctx context.Context) ([]plugin.InstanceRecord, error) {
	rows, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]plugin.InstanceRecord, 0, len(rows))
	for _, row := range rows {
		config := json.RawMessage(row.Config)
		if len(config) == 0 {
			config = json.RawMessage("{}")
		}
		out = append(out, plugin.InstanceRecord{
			ID:             row.ID,
			PluginType:     plugin.PluginType(row.PluginType),
			Implementation: row.Implementation,
			Name:           row.Name,
			Config:         config,
			SecretsEnc:     row.SecretsEnc,
			Enabled:        row.Enabled,
			CreatedAt:      row.CreatedAt,
			UpdatedAt:      row.UpdatedAt,
		})
	}
	return out, nil
}

func (s *pluginStore) Delete(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, db.ErrPluginInstanceNotFound) {
			return plugin.ErrPluginNotFound
		}
		return err
	}
	return nil
}
