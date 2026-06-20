package outbound

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPinnedClient_RejectsPrivateHost(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	c, err := NewPinnedClient(srv.URL, 0)
	require.NoError(t, err)
	_, err = c.Get(context.Background(), "http://127.0.0.1:9999/api", nil)
	require.Error(t, err)
}

func TestPinnedClient_GetSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api", r.URL.Path)
		require.Equal(t, "1", r.URL.Query().Get("t"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	t.Cleanup(srv.Close)

	c, err := NewPinnedClient(srv.URL, 0, AllowPrivateHosts())
	require.NoError(t, err)
	resp, err := c.Get(context.Background(), "/api", map[string][]string{"t": {"1"}})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()
}

func TestPinnedClient_RejectsHostMismatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	c, err := NewPinnedClient(srv.URL, 0, AllowPrivateHosts())
	require.NoError(t, err)
	_, err = c.Get(context.Background(), "http://evil.example/api", nil)
	require.Error(t, err)
}

func TestNewPinnedClient_InvalidURL(t *testing.T) {
	_, err := NewPinnedClient("://bad", 0)
	require.Error(t, err)
	_, err = NewPinnedClient("example.com", 0)
	require.Error(t, err)
}

func TestPinnedClient_PostForm(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
		require.NoError(t, r.ParseForm())
		require.Equal(t, "v", r.Form.Get("k"))
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	c, err := NewPinnedClient(srv.URL, 0, AllowPrivateHosts())
	require.NoError(t, err)
	resp, err := c.PostForm(context.Background(), "/form", map[string][]string{"k": {"v"}})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

func TestPinnedClient_NotInitialized(t *testing.T) {
	var c *PinnedClient
	_, err := c.Get(context.Background(), "/api", nil)
	require.Error(t, err)
}

func TestPinnedClient_BlocksLoopbackIP(t *testing.T) {
	c, err := NewPinnedClient("http://example.com", 0)
	require.NoError(t, err)
	_, err = c.Get(context.Background(), "http://127.0.0.1/", nil)
	require.Error(t, err)
}

func TestPinnedClient_BlocksLocalhostHostname(t *testing.T) {
	c, err := NewPinnedClient("http://example.com", 0)
	require.NoError(t, err)
	_, err = c.Get(context.Background(), "http://localhost/", nil)
	require.Error(t, err)
}

func TestPinnedClient_RedirectSchemeMismatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://"+r.Host+r.URL.Path, http.StatusFound)
	}))
	t.Cleanup(srv.Close)

	c, err := NewPinnedClient(srv.URL, 0, AllowPrivateHosts())
	require.NoError(t, err)
	_, err = c.Get(context.Background(), "/redirect", nil)
	require.Error(t, err)
}
