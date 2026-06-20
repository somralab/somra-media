package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/auth"
	"github.com/somralab/somra-media/internal/platform/db"
	"github.com/somralab/somra-media/internal/plugin"
	"github.com/somralab/somra-media/internal/plugin/stub"
)

func newPluginTestRouter(t *testing.T) (http.Handler, *db.DB, string) {
	t.Helper()
	d := openTestDB(t)
	svc := newTestAuthService(t, d)
	mgr := plugin.NewManager(newPluginTestStore(d), plugin.ManagerOptions{EncryptionKey: "test-jwt-secret"})
	require.NoError(t, mgr.RegisterFactory(stub.NewIndexerFactory()))
	require.NoError(t, mgr.RegisterFactory(stub.NewDownloadClientFactory()))

	h := New(Options{
		AuthHandlers:   &AuthHandlers{Service: svc},
		AuthMiddleware: &AuthMiddleware{Service: svc},
		PluginHandlers: &PluginHandlers{Manager: mgr},
	})

	setupBody, _ := json.Marshal(map[string]string{"username": "admin", "password": "AdminPass1"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/setup/admin", bytes.NewReader(setupBody))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var tok map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &tok))
	return h, d, tok["accessToken"].(string)
}

type pluginTestStore struct {
	repo *db.PluginInstanceRepo
}

func newPluginTestStore(d *db.DB) plugin.Store {
	return &pluginTestStore{repo: db.NewPluginInstanceRepo(d.Querier())}
}

func (s *pluginTestStore) Create(ctx context.Context, rec plugin.InstanceRecord) (int64, error) {
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

func (s *pluginTestStore) UpdateConfig(ctx context.Context, id int64, config json.RawMessage, secretsEnc string) error {
	raw := string(config)
	if raw == "" {
		raw = "{}"
	}
	return s.repo.UpdateConfig(ctx, id, raw, secretsEnc)
}

func (s *pluginTestStore) UpdateName(ctx context.Context, id int64, name string) error {
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

func (s *pluginTestStore) SetEnabled(ctx context.Context, id int64, enabled bool) error {
	if err := s.repo.SetEnabled(ctx, id, enabled); err != nil {
		if errors.Is(err, db.ErrPluginInstanceNotFound) {
			return plugin.ErrPluginNotFound
		}
		return err
	}
	return nil
}

func (s *pluginTestStore) GetByID(ctx context.Context, id int64) (plugin.InstanceRecord, error) {
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

func (s *pluginTestStore) List(ctx context.Context) ([]plugin.InstanceRecord, error) {
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

func (s *pluginTestStore) Delete(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, db.ErrPluginInstanceNotFound) {
			return plugin.ErrPluginNotFound
		}
		return err
	}
	return nil
}

func TestPluginHandlers_CatalogAndCRUD(t *testing.T) {
	h, _, token := newPluginTestRouter(t)

	req := authRequest(http.MethodGet, "/api/v1/plugins/catalog", token, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var catalog map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &catalog))
	items := catalog["catalog"].([]any)
	require.GreaterOrEqual(t, len(items), 2)

	req = authRequest(http.MethodGet, "/api/v1/plugins/instances", token, nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	createBody, _ := json.Marshal(map[string]any{
		"pluginType":     "indexer",
		"implementation": stub.Implementation,
		"name":           "test-indexer",
		"config":         map[string]any{"prefix": "demo", "apiKey": "secret-key"},
		"enabled":        true,
	})
	req = authRequest(http.MethodPost, "/api/v1/plugins/instances", token, createBody)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var created map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))
	id := int64(created["id"].(float64))
	cfg := created["config"].(map[string]any)
	assert.Equal(t, true, cfg["apiKeySet"])
	assert.NotContains(t, cfg, "apiKey")
	assert.NotContains(t, fmt.Sprint(created), "secret-key")

	req = authRequest(http.MethodGet, fmt.Sprintf("/api/v1/plugins/instances/%d", id), token, nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	patchBody, _ := json.Marshal(map[string]any{"config": map[string]any{"prefix": "updated"}})
	req = authRequest(http.MethodPatch, fmt.Sprintf("/api/v1/plugins/instances/%d", id), token, patchBody)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	req = authRequest(http.MethodPost, fmt.Sprintf("/api/v1/plugins/instances/%d/test", id), token, nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var testResult map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &testResult))
	assert.Equal(t, true, testResult["success"])

	req = authRequest(http.MethodDelete, fmt.Sprintf("/api/v1/plugins/instances/%d", id), token, nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNoContent, rec.Code)
}

func TestPluginHandlers_CreateInvalid(t *testing.T) {
	h, _, token := newPluginTestRouter(t)
	req := authRequest(http.MethodPost, "/api/v1/plugins/instances", token, []byte(`{"pluginType":"indexer"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestPluginHandlers_NotFound(t *testing.T) {
	h, _, token := newPluginTestRouter(t)
	req := authRequest(http.MethodGet, "/api/v1/plugins/instances/99999", token, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestPluginHandlers_UserDenied(t *testing.T) {
	h, d, _ := newPluginTestRouter(t)
	ctx := context.Background()
	svc := newTestAuthService(t, d)
	_, err := svc.Register(ctx, "pluginuser", "UserPass1", []string{auth.RoleUser})
	require.NoError(t, err)
	_, pair, err := svc.Login(ctx, "pluginuser", "UserPass1", "web", "127.0.0.1")
	require.NoError(t, err)

	req := authRequest(http.MethodGet, "/api/v1/plugins/instances", pair.AccessToken, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)
}
