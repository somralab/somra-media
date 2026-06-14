package bootstrap_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/platform/bootstrap"
	"github.com/somralab/somra-media/internal/platform/config"
	"github.com/somralab/somra-media/internal/platform/db"
)

func TestWireAuth(t *testing.T) {
	dbCfg := db.Default()
	dbCfg.DataDir = filepath.Join(t.TempDir(), "data")

	c, err := bootstrap.NewWithStorage(context.Background(), nil, dbCfg)
	require.NoError(t, err)
	t.Cleanup(func() { _ = c.Close() })

	bundle, err := bootstrap.WireAuth(c, config.AuthConfig{
		JWTSecret:     "test-secret-key-at-least-32-bytes!!",
		RefreshPepper: "pepper1234567890",
	}, nil)
	require.NoError(t, err)
	require.NotNil(t, bundle.Service)
	require.NotNil(t, bundle.Users)
}

func TestWireAuth_ShortSecretGeneratesEphemeral(t *testing.T) {
	dbCfg := db.Default()
	dbCfg.DataDir = filepath.Join(t.TempDir(), "data")

	c, err := bootstrap.NewWithStorage(context.Background(), nil, dbCfg)
	require.NoError(t, err)
	t.Cleanup(func() { _ = c.Close() })

	bundle, err := bootstrap.WireAuth(c, config.AuthConfig{JWTSecret: "short"}, nil)
	require.NoError(t, err)
	require.NotNil(t, bundle.Service)
}

func TestWireAuth_NilComponents(t *testing.T) {
	_, err := bootstrap.WireAuth(nil, config.AuthConfig{}, nil)
	require.Error(t, err)
}

func TestDevJWTSecret(t *testing.T) {
	t.Setenv("SOMRA_JWT_SECRET", "from-env-secret")
	assert.Equal(t, "from-env-secret", bootstrap.DevJWTSecret())
}
