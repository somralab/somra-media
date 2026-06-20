//go:build acquisition

package nzbindexer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/platform/outbound"
	"github.com/somralab/somra-media/internal/plugin"
)

const capsXML = `<?xml version="1.0"?><caps><searching><search available="yes"/></searching><categories><category id="2000" name="Movies"/></categories><limits max="100"/></caps>`

const searchXML = `<?xml version="1.0"?><rss><channel><item>
<title>Test.Movie.2024.1080p.WEB-DL.x265</title>
<link>http://127.0.0.1:0/d/download</link>
<enclosure url="http://127.0.0.1:0/d/download" length="12345"/>
<torznab:attr xmlns:torznab="http://torznab.com/schemas/2015/feed" name="seeders" value="10"/>
</item></channel></rss>`

func TestClient_Search(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("t") {
		case "caps":
			_, _ = w.Write([]byte(capsXML))
		default:
			require.Equal(t, "movie", r.URL.Query().Get("t"))
			_, _ = w.Write([]byte(searchXML))
		}
	}))
	t.Cleanup(srv.Close)

	pinned, err := outbound.NewPinnedClient(srv.URL, 0, outbound.AllowPrivateHosts())
	require.NoError(t, err)
	c := &Client{
		client:   pinned,
		apiKey:   "key",
		protocol: ProtocolTorrent,
	}
	caps, err := c.Capabilities(context.Background())
	require.NoError(t, err)
	require.True(t, caps.SupportsSearch)

	results, err := c.Search(context.Background(), plugin.SearchQuery{Title: "Test Movie", MediaKind: plugin.MediaKindMovie})
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, "Test.Movie.2024.1080p.WEB-DL.x265", results[0].Title)
	require.NotEmpty(t, results[0].ReleaseID)
}

func TestDecodeReleaseID(t *testing.T) {
	id := EncodeReleaseID("magnet:?xt=urn:btih:abc")
	link, err := DecodeReleaseID(id)
	require.NoError(t, err)
	require.Equal(t, "magnet:?xt=urn:btih:abc", link)
}
