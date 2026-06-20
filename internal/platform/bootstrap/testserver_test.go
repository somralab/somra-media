package bootstrap

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTestServer_BootsFullStack(t *testing.T) {
	ts := NewTestServer(t)
	require.NotEmpty(t, ts.AdminToken)
	require.NotNil(t, ts.Plugins.Manager)
	assert.Empty(t, ts.Plugins.Manager.EnabledIndexers())

	resp, err := ts.Server.Client().Get(ts.Server.URL + "/api/v1/health") //nolint:noctx
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
