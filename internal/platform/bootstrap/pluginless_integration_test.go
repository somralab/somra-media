//go:build integration

package bootstrap_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/platform/bootstrap"
)

func TestIntegration_PluginlessCore(t *testing.T) {
	ts := bootstrap.NewTestServer(t)

	assert.Empty(t, ts.Plugins.Manager.EnabledIndexers())
	assert.Empty(t, ts.Plugins.Manager.EnabledDownloadClients())

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, ts.Server.URL+"/api/v1/health", nil)
	require.NoError(t, err)
	resp, err := ts.Server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var health map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&health))
	assert.Equal(t, "ok", health["status"])

	req, err = http.NewRequestWithContext(context.Background(), http.MethodGet, ts.Server.URL+"/api/v1/settings", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+ts.AdminToken)
	resp, err = ts.Server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	req, err = http.NewRequestWithContext(context.Background(), http.MethodGet, ts.Server.URL+"/api/v1/libraries", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+ts.AdminToken)
	resp, err = ts.Server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	createBody, _ := json.Marshal(map[string]any{
		"mediaKind":  "movie",
		"provider":   "tmdb",
		"externalId": "pluginless-1",
		"title":      "Pluginless Movie",
	})
	req, err = http.NewRequestWithContext(context.Background(), http.MethodPost, ts.Server.URL+"/api/v1/requests", bytes.NewReader(createBody))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+ts.AdminToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err = ts.Server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	assert.Equal(t, "approved", created["status"])

	req, err = http.NewRequestWithContext(context.Background(), http.MethodGet, ts.Server.URL+"/api/v1/plugins/instances", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+ts.AdminToken)
	resp, err = ts.Server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var plugins map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&plugins))
	instances, ok := plugins["instances"].([]any)
	require.True(t, ok, "instances key must be an array, got %T", plugins["instances"])
	assert.Empty(t, instances, fmt.Sprintf("expected no plugin instances, got %d", len(instances)))
}
