package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPluginHandlers_DisableViaPatch(t *testing.T) {
	h, _, token := newPluginTestRouter(t)

	body := []byte(`{"pluginType":"indexer","implementation":"stub","name":"disable-me","enabled":true}`)
	req := authRequest(http.MethodPost, "/api/v1/plugins/instances", token, body)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var created map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))
	id := int64(created["id"].(float64))

	patchBody := []byte(`{"enabled":false}`)
	req = authRequest(http.MethodPatch, fmt.Sprintf("/api/v1/plugins/instances/%d", id), token, patchBody)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var patched map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &patched))
	assert.Equal(t, false, patched["enabled"])
}

func TestPluginHandlers_PatchConfigOnly(t *testing.T) {
	h, _, token := newPluginTestRouter(t)

	body := []byte(`{"pluginType":"indexer","implementation":"stub","name":"cfg-only","config":{"prefix":"a"}}`)
	req := authRequest(http.MethodPost, "/api/v1/plugins/instances", token, body)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var created map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))
	id := int64(created["id"].(float64))

	patchBody := []byte(`{"config":{"prefix":"b","apiKey":"new-secret"}}`)
	req = authRequest(http.MethodPatch, fmt.Sprintf("/api/v1/plugins/instances/%d", id), token, patchBody)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var patched map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &patched))
	cfg := patched["config"].(map[string]any)
	assert.Equal(t, "b", cfg["prefix"])
	assert.Equal(t, true, cfg["apiKeySet"])
}

func TestPluginHandlers_EnableViaPatch(t *testing.T) {
	h, _, token := newPluginTestRouter(t)

	body := []byte(`{"pluginType":"indexer","implementation":"stub","name":"enable-me","enabled":false}`)
	req := authRequest(http.MethodPost, "/api/v1/plugins/instances", token, body)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var created map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))
	id := int64(created["id"].(float64))
	assert.Equal(t, false, created["enabled"])

	patchBody := []byte(`{"enabled":true}`)
	req = authRequest(http.MethodPatch, fmt.Sprintf("/api/v1/plugins/instances/%d", id), token, patchBody)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var patched map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &patched))
	assert.Equal(t, true, patched["enabled"])
}

func TestPluginHandlers_ListMultipleInstances(t *testing.T) {
	h, _, token := newPluginTestRouter(t)

	for i := 0; i < 2; i++ {
		body := []byte(fmt.Sprintf(`{"pluginType":"indexer","implementation":"stub","name":"list-%d"}`, i))
		req := authRequest(http.MethodPost, "/api/v1/plugins/instances", token, body)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code)
	}

	req := authRequest(http.MethodGet, "/api/v1/plugins/instances", token, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var out map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	instances := out["instances"].([]any)
	require.Len(t, instances, 2)
}

func TestPluginHandlers_PatchDuplicateName(t *testing.T) {
	h, _, token := newPluginTestRouter(t)

	var id2 int64
	for i, name := range []string{"indexer-a", "indexer-b"} {
		createBody := []byte(`{"pluginType":"indexer","implementation":"stub","name":"` + name + `"}`)
		req := authRequest(http.MethodPost, "/api/v1/plugins/instances", token, createBody)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code)
		if i == 1 {
			var created map[string]any
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))
			id2 = int64(created["id"].(float64))
		}
	}

	patchBody := []byte(`{"name":"indexer-a"}`)
	req := authRequest(http.MethodPatch, fmt.Sprintf("/api/v1/plugins/instances/%d", id2), token, patchBody)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestPluginHandlers_PatchInvalidJSON(t *testing.T) {
	h, _, token := newPluginTestRouter(t)

	createBody := []byte(`{"pluginType":"indexer","implementation":"stub","name":"patch-target"}`)
	req := authRequest(http.MethodPost, "/api/v1/plugins/instances", token, createBody)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var created map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))
	id := int64(created["id"].(float64))

	req = authRequest(http.MethodPatch, fmt.Sprintf("/api/v1/plugins/instances/%d", id), token, []byte(`not-json`))
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestPluginHandlers_DeleteNotFound(t *testing.T) {
	h, _, token := newPluginTestRouter(t)
	req := authRequest(http.MethodDelete, "/api/v1/plugins/instances/99999", token, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestPluginHandlers_TestNotFound(t *testing.T) {
	h, _, token := newPluginTestRouter(t)
	req := authRequest(http.MethodPost, "/api/v1/plugins/instances/99999/test", token, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestPluginHandlers_CreateWithDefaultConfig(t *testing.T) {
	h, _, token := newPluginTestRouter(t)
	body := []byte(`{"pluginType":"indexer","implementation":"stub","name":"no-config"}`)
	req := authRequest(http.MethodPost, "/api/v1/plugins/instances", token, body)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)
}

func TestPluginHandlers_CreateUnknownImplementation(t *testing.T) {
	h, _, token := newPluginTestRouter(t)
	body := []byte(`{"pluginType":"indexer","implementation":"missing","name":"bad","config":{}}`)
	req := authRequest(http.MethodPost, "/api/v1/plugins/instances", token, body)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestPluginHandlers_TestInvalidInstanceID(t *testing.T) {
	h, _, token := newPluginTestRouter(t)
	req := authRequest(http.MethodPost, "/api/v1/plugins/instances/abc/test", token, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}
