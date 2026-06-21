//go:build soak

package bootstrap_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/api"
	"github.com/somralab/somra-media/internal/platform/bootstrap"
	"github.com/somralab/somra-media/internal/platform/config"
	"github.com/somralab/somra-media/internal/platform/db"
)

// TestSoakHealthProbes runs a shortened soak loop for manual/CI -tags=soak runs.
func TestSoakHealthProbes(t *testing.T) {
	t.Setenv("SOMRA_DATA_DIR", filepath.Join(t.TempDir(), "data"))

	cfg, err := config.Load()
	require.NoError(t, err)

	dbCfg := db.Default()
	dbCfg.DataDir = cfg.Data.Dir

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	c, err := bootstrap.NewWithStorage(ctx, nil, dbCfg)
	require.NoError(t, err)
	t.Cleanup(func() { _ = c.Close() })

	c.Scheduler.Start(context.Background())
	t.Cleanup(func() {
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer stopCancel()
		_ = c.Scheduler.Stop(stopCtx)
	})

	handler := api.New(api.Options{
		Logger:           c.Logger,
		HealthAggregator: api.NewDiagnosticsAggregator(c.Diagnostics),
	})
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	client := srv.Client()
	deadline := time.Now().Add(90 * time.Second)
	var lastRSS int
	for time.Now().Before(deadline) {
		resp, err := client.Get(srv.URL + "/api/v1/health")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		_ = resp.Body.Close()
		lastRSS++
		time.Sleep(5 * time.Second)
	}
	require.Greater(t, lastRSS, 10, "expected multiple health probes")
}
