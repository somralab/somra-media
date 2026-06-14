package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeSPAFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "index.html"),
		[]byte("<!doctype html><title>spa</title>"), 0o644))
	assetDir := filepath.Join(dir, "assets")
	require.NoError(t, os.MkdirAll(assetDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(assetDir, "app.js"),
		[]byte("console.log('app');"), 0o644))
	return dir
}

func TestRouter_SPADisabledByDefault(t *testing.T) {
	h := newTestHandler(t, Options{})
	req := httptest.NewRequest(http.MethodGet, "/about", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")
}

func TestRouter_SPAServesIndexOnUnknownRoute(t *testing.T) {
	dir := writeSPAFixture(t)
	h := newTestHandler(t, Options{WebDir: dir})

	req := httptest.NewRequest(http.MethodGet, "/about/somewhere", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "<title>spa</title>")
}

func TestRouter_SPAServesIndexAtRoot(t *testing.T) {
	dir := writeSPAFixture(t)
	h := newTestHandler(t, Options{WebDir: dir})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "spa")
}

func TestRouter_SPAServesStaticAsset(t *testing.T) {
	dir := writeSPAFixture(t)
	h := newTestHandler(t, Options{WebDir: dir})

	req := httptest.NewRequest(http.MethodGet, "/assets/app.js", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "console.log")
}

func TestRouter_SPADoesNotShadowAPI(t *testing.T) {
	dir := writeSPAFixture(t)
	h := newTestHandler(t, Options{WebDir: dir})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")
}

func TestRouter_SPAUnknownAPIPathStillJSON(t *testing.T) {
	dir := writeSPAFixture(t)
	h := newTestHandler(t, Options{WebDir: dir})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/nope", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")
}

func TestRouter_SPARejectsTraversal(t *testing.T) {
	dir := writeSPAFixture(t)
	h := newTestHandler(t, Options{WebDir: dir})

	req := httptest.NewRequest(http.MethodGet, "/../../etc/passwd", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.True(t, rec.Code == http.StatusOK || rec.Code == http.StatusNotFound,
		"traversal must not escape; got %d", rec.Code)
	if rec.Code == http.StatusOK {
		assert.False(t, strings.Contains(rec.Body.String(), "root:x:"), "must not leak /etc/passwd")
	}
}

func TestRouter_SPASkippedWhenDirMissing(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist")
	h := newTestHandler(t, Options{WebDir: missing})

	req := httptest.NewRequest(http.MethodGet, "/about", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")
}

func TestRouter_SPAServesHEAD(t *testing.T) {
	dir := writeSPAFixture(t)
	h := newTestHandler(t, Options{WebDir: dir})

	req := httptest.NewRequest(http.MethodHead, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "text/html")
	assert.Empty(t, rec.Body.String(), "HEAD must not write a response body")
}

func TestRouter_SPASkippedWhenIndexMissing(t *testing.T) {
	dir := t.TempDir()
	h := newTestHandler(t, Options{WebDir: dir})

	req := httptest.NewRequest(http.MethodGet, "/about", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}
