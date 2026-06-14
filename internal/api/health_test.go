package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestHandler(t *testing.T, opts Options) http.Handler {
	t.Helper()
	if opts.Now == nil {
		fixed := time.Date(2026, time.June, 13, 21, 0, 0, 0, time.UTC)
		opts.Now = func() time.Time { return fixed }
	}
	return New(opts)
}

type stubHealthCheck struct {
	name   string
	status HealthStatus
}

func (s stubHealthCheck) Name() string        { return s.name }
func (s stubHealthCheck) Check() HealthStatus { return s.status }

func TestHealth_OK(t *testing.T) {
	h := newTestHandler(t, Options{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")

	var body HealthResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "ok", body.Status)
	_, err := time.Parse(time.RFC3339Nano, body.Time)
	require.NoError(t, err, "time field must be RFC3339")
	assert.Empty(t, body.Checks)
}

func TestHealth_DegradedWhenCheckFails(t *testing.T) {
	checks := []HealthCheck{
		stubHealthCheck{name: "db", status: HealthStatus{Status: "ok"}},
		stubHealthCheck{name: "scheduler", status: HealthStatus{Status: "down", Detail: "stopped"}},
	}
	h := newTestHandler(t, Options{HealthCheck: checks})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var body HealthResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "degraded", body.Status)
	assert.Equal(t, "ok", body.Checks["db"].Status)
	assert.Equal(t, "down", body.Checks["scheduler"].Status)
}
