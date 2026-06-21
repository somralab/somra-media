//go:build integration

package bootstrap_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/platform/bootstrap"
)

func TestIntegration_PluginLifecycle(t *testing.T) {
	ts := bootstrap.NewTestServer(t)
	ctx := context.Background()

	createBody := []byte(`{"pluginType":"indexer","implementation":"stub","name":"lifecycle-idx","enabled":true}`)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ts.Server.URL+"/api/v1/plugins/instances", bytes.NewReader(createBody))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+ts.AdminToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := ts.Server.Client().Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var created map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	resp.Body.Close()
	id := int64(created["id"].(float64))

	patchBody := []byte(`{"name":"lifecycle-renamed","config":{"prefix":"test-prefix"},"enabled":false}`)
	req, err = http.NewRequestWithContext(ctx, http.MethodPatch, ts.Server.URL+"/api/v1/plugins/instances/"+strconv.FormatInt(id, 10), bytes.NewReader(patchBody))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+ts.AdminToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err = ts.Server.Client().Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	enableBody := []byte(`{"enabled":true}`)
	req, err = http.NewRequestWithContext(ctx, http.MethodPatch, ts.Server.URL+"/api/v1/plugins/instances/"+strconv.FormatInt(id, 10), bytes.NewReader(enableBody))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+ts.AdminToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err = ts.Server.Client().Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	req, err = http.NewRequestWithContext(ctx, http.MethodDelete, ts.Server.URL+"/api/v1/plugins/instances/"+strconv.FormatInt(id, 10), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+ts.AdminToken)
	resp, err = ts.Server.Client().Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()
}
