package subtitles_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/subtitles"
)

type stubTransport func(*http.Request) (*http.Response, error)

func (f stubTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestOpenSubtitlesSearchWithStubTransport(t *testing.T) {
	p := subtitles.NewOpenSubtitles("test-api-key")
	p.Client = &http.Client{
		Transport: stubTransport(func(req *http.Request) (*http.Response, error) {
			body := `{"data":[{"id":"42","attributes":{"language":"en","release":"Release","download_count":10,"feature_details":{"title":"Matrix","year":1999}}}]}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	results, err := p.Search(context.Background(), subtitles.SearchQuery{Title: "Matrix", Language: "en"})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "42", results[0].ExternalID)
	assert.Greater(t, results[0].Score, 50.0)
}

func TestOpenSubtitlesDownloadWithStubTransport(t *testing.T) {
	p := subtitles.NewOpenSubtitles("test-api-key")
	p.Client = &http.Client{
		Transport: stubTransport(func(req *http.Request) (*http.Response, error) {
			if req.Method == http.MethodPost {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(`{"link":"https://dl.opensubtitles.org/download/sub.srt"}`)),
					Header:     make(http.Header),
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("1\n00:00:01,000 --> 00:00:02,000\nHi")),
				Header:     make(http.Header),
			}, nil
		}),
	}

	data, err := p.Download(context.Background(), "42", "en")
	require.NoError(t, err)
	assert.Contains(t, string(data), "Hi")
}

func TestOpenSubtitlesSearchBadStatus(t *testing.T) {
	p := subtitles.NewOpenSubtitles("test-api-key")
	p.Client = &http.Client{
		Transport: stubTransport(func(_ *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusBadGateway, Body: io.NopCloser(strings.NewReader("")), Header: make(http.Header)}, nil
		}),
	}
	_, err := p.Search(context.Background(), subtitles.SearchQuery{Title: "X"})
	require.Error(t, err)
}

func TestOpenSubtitlesDownloadEmptyLink(t *testing.T) {
	p := subtitles.NewOpenSubtitles("test-api-key")
	p.Client = &http.Client{
		Transport: stubTransport(func(req *http.Request) (*http.Response, error) {
			if req.Method == http.MethodPost {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(`{"link":""}`)),
					Header:     make(http.Header),
				}, nil
			}
			return nil, nil
		}),
	}
	_, err := p.Download(context.Background(), "1", "en")
	require.Error(t, err)
}

func TestOpenSubtitlesSearchInvalidJSON(t *testing.T) {
	p := subtitles.NewOpenSubtitles("test-api-key")
	p.Client = &http.Client{
		Transport: stubTransport(func(_ *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`not-json`)),
				Header:     make(http.Header),
			}, nil
		}),
	}
	_, err := p.Search(context.Background(), subtitles.SearchQuery{Title: "X"})
	require.Error(t, err)
}
