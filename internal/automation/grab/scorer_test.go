package grab

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/plugin"
	"github.com/somralab/somra-media/internal/requests"
)

func TestPickBest(t *testing.T) {
	spec, err := ParseProfileSpec(`{"preferredResolutions":["1080p","720p"]}`)
	require.NoError(t, err)
	results := []plugin.SearchResult{
		{Title: "A 720p HDTV", Resolution: "720p", Codec: "h264", Protocol: plugin.ProtocolTorrent},
		{Title: "B 1080p WEB-DL", Resolution: "1080p", Codec: "hevc", Protocol: plugin.ProtocolTorrent, SizeBytes: 1000},
	}
	best := PickBest(results, spec, requests.Request{QualityResolution: requests.QualityHD1080})
	require.NotNil(t, best)
	require.Contains(t, best.Title, "1080p")
}

func TestPickBestRejectsIgnored(t *testing.T) {
	spec, err := ParseProfileSpec(`{"ignoredTerms":["cam"]}`)
	require.NoError(t, err)
	results := []plugin.SearchResult{{Title: "Movie CAM rip", Resolution: "1080p"}}
	require.Nil(t, PickBest(results, spec, requests.Request{}))
}
