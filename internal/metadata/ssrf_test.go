package metadata_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/metadata"
)

func TestValidateOutboundURL_BlocksPrivateHosts(t *testing.T) {
	t.Parallel()
	err := metadata.ValidateOutboundURL("https://127.0.0.1/movie")
	require.Error(t, err)

	err = metadata.ValidateOutboundURL("http://api.themoviedb.org/3/movie/1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "https")
}

func TestValidateOutboundURL_BlocksUnknownHost(t *testing.T) {
	t.Parallel()
	err := metadata.ValidateOutboundURL("https://evil.example.com/payload")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "allowlist")
}

func TestValidateOutboundURL_AllowsTMDB(t *testing.T) {
	t.Parallel()
	err := metadata.ValidateOutboundURL("https://api.themoviedb.org/3/movie/550")
	require.NoError(t, err)
}

func TestSafeHTTPClient_NotNil(t *testing.T) {
	t.Parallel()
	c := metadata.SafeHTTPClient(0)
	require.NotNil(t, c)
}
