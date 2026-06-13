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

func TestVersion_DefaultsWhenLdflagsEmpty(t *testing.T) {
	fixed := time.Date(2026, time.June, 13, 21, 0, 0, 0, time.UTC)
	h := newTestHandler(t, Options{
		Now: func() time.Time { return fixed },
	})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/version", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var body VersionResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "0.1.0-dev", body.Version)
	assert.Equal(t, "", body.Commit)
	assert.Equal(t, fixed.Format(time.RFC3339), body.BuiltAt)
}

func TestVersion_HonoursBuildInfo(t *testing.T) {
	build := BuildInfo{
		Version: "1.2.3",
		Commit:  "abc1234",
		BuiltAt: "2026-06-13T12:00:00Z",
	}
	h := newTestHandler(t, Options{Build: build})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/version", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var body VersionResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, build, BuildInfo(body))
}
