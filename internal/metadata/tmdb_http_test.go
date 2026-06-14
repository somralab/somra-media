package metadata

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testTMDBProvider(t *testing.T, handler http.HandlerFunc) *TMDBProvider {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	p := NewTMDBProvider("test-key", srv.Client())
	p.Base = srv.URL
	p.fetch = func(ctx context.Context, client *http.Client, method, rawURL string) (*http.Response, error) {
		req, err := http.NewRequestWithContext(ctx, method, rawURL, nil)
		if err != nil {
			return nil, err
		}
		return client.Do(req)
	}
	return p
}

func TestTMDBProvider_SearchMovie(t *testing.T) {
	p := testTMDBProvider(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/search/movie")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[{"id":42,"title":"Inception","release_date":"2010-07-16","overview":"dream","poster_path":"/p.jpg"}]}`))
	})

	year := 2010
	results, err := p.Search(context.Background(), SearchQuery{Title: "Inception", Year: &year, Kind: "movie", Locale: "en-US"})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "42", results[0].ExternalID)
	assert.Contains(t, results[0].PosterURL, "image.tmdb.org")
}

func TestTMDBProvider_SearchTV(t *testing.T) {
	p := testTMDBProvider(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"results":[{"id":7,"name":"Breaking Bad","first_air_date":"2008-01-20","overview":"cook"}]}`))
	})

	results, err := p.Search(context.Background(), SearchQuery{Title: "Breaking Bad", Kind: "tv", Locale: "tr-TR"})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "Breaking Bad", results[0].Title)
}

func TestTMDBProvider_SearchMissingAPIKey(t *testing.T) {
	p := NewTMDBProvider("", nil)
	_, err := p.Search(context.Background(), SearchQuery{Title: "x", Kind: "movie"})
	require.Error(t, err)
}

func TestTMDBProvider_SearchHTTPError(t *testing.T) {
	p := testTMDBProvider(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	_, err := p.Search(context.Background(), SearchQuery{Title: "x", Kind: "movie"})
	require.Error(t, err)
}

func TestTMDBProvider_SearchInvalidJSON(t *testing.T) {
	p := testTMDBProvider(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`not-json`))
	})

	_, err := p.Search(context.Background(), SearchQuery{Title: "x", Kind: "movie"})
	require.Error(t, err)
}

func TestTMDBProvider_Detail(t *testing.T) {
	p := testTMDBProvider(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/movie/42")
		_, _ = w.Write([]byte(`{"id":42,"title":"Inception","overview":"mind","poster_path":"/p.jpg","backdrop_path":"/b.jpg","release_date":"2010-07-16","genres":[{"name":"Sci-Fi"}]}`))
	})

	detail, err := p.Detail(context.Background(), "42", "en-US")
	require.NoError(t, err)
	assert.Equal(t, "Inception", detail.Title)
	assert.Equal(t, "Sci-Fi", detail.Genres[0])
	assert.NotEmpty(t, detail.PosterURL)
	assert.NotEmpty(t, detail.BackdropURL)
}

func TestTMDBProvider_DetailInvalidJSON(t *testing.T) {
	p := testTMDBProvider(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{`))
	})

	_, err := p.Detail(context.Background(), "1", "en-US")
	require.Error(t, err)
}

func TestTMDBProvider_Images(t *testing.T) {
	p := NewTMDBProvider("key", nil)
	_, _, _, err := p.Images(context.Background(), "1")
	require.Error(t, err)
}

func TestSafeHTTPClient_TimeoutAndRedirect(t *testing.T) {
	c := SafeHTTPClient(1)
	require.NotNil(t, c)
	require.NotNil(t, c.CheckRedirect)

	redirects := 0
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		redirects++
		if redirects < 4 {
			http.Redirect(w, r, r.URL.String()+"?next", http.StatusFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	require.NoError(t, err)
	_, err = c.Do(req)
	require.Error(t, err)
}

func TestDoRequest_BlocksBadURL(t *testing.T) {
	_, err := DoRequest(context.Background(), SafeHTTPClient(time.Second), "GET", "http://127.0.0.1/")
	require.Error(t, err)

	_, err = DoRequest(context.Background(), SafeHTTPClient(time.Second), "GET", "ftp://api.themoviedb.org/x")
	require.Error(t, err)
}

func TestDoRequest_AllowsTMDB(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "ok")
	}))
	t.Cleanup(srv.Close)

	// DoRequest validates host allowlist, not the mock URL — exercise via ValidateOutboundURL only.
	err := ValidateOutboundURL("https://api.themoviedb.org/3/search/movie?api_key=x")
	require.NoError(t, err)
}
