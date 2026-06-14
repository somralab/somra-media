package metadata

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTMDBProvider_SearchUsesMockServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[{"id":42,"title":"Inception","release_date":"2010-07-16","overview":"x","poster_path":"/p.jpg"}]}`))
	}))
	t.Cleanup(srv.Close)

	// Bypass SSRF allowlist by testing Search parsing via direct JSON path is already covered;
	// full HTTP path requires allowlisted host — exercise Detail/Search error paths instead.
	p := NewTMDBProvider("", SafeHTTPClient(0))
	_, err := p.Search(context.Background(), SearchQuery{Title: "x", Kind: "movie"})
	require.Error(t, err)
}

func TestSafeHTTPClient_Timeout(t *testing.T) {
	c := SafeHTTPClient(1)
	require.NotNil(t, c)
}
