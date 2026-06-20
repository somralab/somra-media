package indexer

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/plugin"
)

func TestDedupeResults(t *testing.T) {
	in := []plugin.SearchResult{
		{Title: "A", SizeBytes: 1, Protocol: plugin.ProtocolTorrent},
		{Title: "a", SizeBytes: 1, Protocol: plugin.ProtocolTorrent},
		{Title: "B", SizeBytes: 2, Protocol: plugin.ProtocolTorrent},
	}
	out := dedupeResults(in)
	require.Len(t, out, 2)
}
