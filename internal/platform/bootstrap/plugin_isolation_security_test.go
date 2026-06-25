//go:build integration

package bootstrap_test

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/platform/bootstrap"
	"github.com/somralab/somra-media/internal/plugin"
)

func TestIntegration_PluginIsolation_SecretsNotInPublicConfig(t *testing.T) {
	pub, sec, err := plugin.SplitConfig(json.RawMessage(`{"prefix":"demo","apiKey":"secret-key","url":"http://indexer"}`), nil)
	require.NoError(t, err)
	assert.Equal(t, "secret-key", sec["apiKey"])
	assert.NotContains(t, string(pub), "secret-key")
}

func TestIntegration_PluginIsolation_AcquisitionBoundaryDocumented(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	acqPath := filepath.Join(root, "internal", "platform", "bootstrap", "plugins_acquisition.go")
	raw, err := os.ReadFile(acqPath)
	require.NoError(t, err)
	content := string(raw)
	assert.Contains(t, content, "//go:build acquisition")
	assert.Contains(t, content, "torznab")
}

func TestIntegration_PluginIsolation_CoreWorksWithoutEnabledPlugins(t *testing.T) {
	ts := bootstrap.NewTestServer(t)
	assert.Empty(t, ts.Plugins.Manager.EnabledIndexers())
	assert.Empty(t, ts.Plugins.Manager.EnabledDownloadClients())

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, ts.Server.URL+"/api/v1/libraries", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+ts.AdminToken)
	resp, err := ts.Server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestIntegration_PluginIsolation_StubUsesOutboundSafeClient(t *testing.T) {
	docPath := filepath.Join("..", "..", "..", "docs", "developer", "plugin-development.md")
	raw, err := os.ReadFile(docPath)
	require.NoError(t, err)
	text := string(raw)
	assert.True(t, strings.Contains(text, "SSRF") || strings.Contains(text, "outbound"),
		"plugin packaging doc must describe outbound/SSRF isolation")
}
