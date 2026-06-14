//go:build integration

// Package api integration tests exercise the production bootstrap path:
// config → bootstrap (i18n + scheduler + DB diagnostics) → chi mux. They
// run under the `integration` build tag so the dedicated CI gate is
// meaningful instead of a trivial smoke. Plain unit tests stay in the
// default suite.
package api_test

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/api"
	"github.com/somralab/somra-media/internal/platform/bootstrap"
	"github.com/somralab/somra-media/internal/platform/config"
	"github.com/somralab/somra-media/internal/platform/db"
	platformerrors "github.com/somralab/somra-media/internal/platform/errors"
)

// integrationFixture mirrors the production wiring inside cmd/somra so
// the assertions in this suite hit the same code path as a real binary.
// The DB lives under t.TempDir() to keep tests hermetic.
type integrationFixture struct {
	srv        *httptest.Server
	components *bootstrap.Components
}

func newIntegrationFixture(t *testing.T) *integrationFixture {
	t.Helper()
	t.Setenv("SOMRA_DATA_DIR", filepath.Join(t.TempDir(), "data"))

	cfg, err := config.Load()
	require.NoError(t, err)

	bootCtx, bootCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer bootCancel()

	dbCfg := db.Default()
	dbCfg.DataDir = cfg.Data.Dir

	c, err := bootstrap.NewWithStorage(bootCtx, nil, dbCfg)
	require.NoError(t, err)

	// Scheduler owns its own context lifetime: pass context.Background()
	// here so the registry-driven SchedulerProvider reports "running"
	// throughout the test instead of degrading the moment bootCtx is
	// cancelled.
	c.Scheduler.Start(context.Background())

	handler := api.New(api.Options{
		Logger: c.Logger,
		Build: api.BuildInfo{
			Version: "0.0.0-integration",
			Commit:  "test",
			BuiltAt: "1970-01-01T00:00:00Z",
		},
		CORS:                cfg.CORS,
		LocalizerMiddleware: c.I18n.Middleware(),
		HealthAggregator:    api.NewDiagnosticsAggregator(c.Diagnostics),
		SSEHeartbeat:        50 * time.Millisecond,
	})

	srv := httptest.NewServer(handler)

	t.Cleanup(func() {
		srv.Close()
		stopCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = c.Scheduler.Stop(stopCtx)
		_ = c.Close()
	})

	return &integrationFixture{srv: srv, components: c}
}

func TestIntegration_HealthIncludesDBCheck(t *testing.T) {
	f := newIntegrationFixture(t)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, f.srv.URL+"/api/v1/health", nil)
	require.NoError(t, err)
	resp, err := f.srv.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "application/json")

	var body api.HealthResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Equal(t, "ok", body.Status)
	require.NotEmpty(t, body.Checks, "checks payload must be populated by the diagnostics registry")

	dbCheck, ok := body.Checks["database"]
	require.True(t, ok, "diagnostics registry must register the database provider")
	assert.Equal(t, "ok", dbCheck.Status)
}

func TestIntegration_VersionReturnsNonEmptyVersion(t *testing.T) {
	f := newIntegrationFixture(t)

	resp, err := f.srv.Client().Get(f.srv.URL + "/api/v1/version") //nolint:noctx
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	var body api.VersionResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.NotEmpty(t, body.Version)
}

func TestIntegration_EventsStreamEmitsHello(t *testing.T) {
	f := newIntegrationFixture(t)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.srv.URL+"/api/v1/events/stream", nil)
	require.NoError(t, err)
	req.Header.Set("Accept", "text/event-stream")

	resp, err := f.srv.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	reader := bufio.NewReader(resp.Body)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("read SSE: %v", err)
		}
		if strings.HasPrefix(strings.TrimSpace(line), "event: hello") {
			return
		}
	}
	t.Fatal("did not receive hello event within 2s")
}

func TestIntegration_NotFoundLocalizedTurkish(t *testing.T) {
	f := newIntegrationFixture(t)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, f.srv.URL+"/api/v1/does-not-exist", nil)
	require.NoError(t, err)
	req.Header.Set("Accept-Language", "tr-TR")

	resp, err := f.srv.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusNotFound, resp.StatusCode)

	var env platformerrors.Envelope
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&env))
	assert.Equal(t, platformerrors.CodeNotFound, env.Code)
	assert.Equal(t, "errors.not_found", env.MessageKey)
	assert.Equal(t, "İstenen kaynak bulunamadı.", env.Message,
		"404 envelope's Message field must be localized to Turkish when Accept-Language=tr-TR")
}
