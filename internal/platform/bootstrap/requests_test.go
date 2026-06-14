package bootstrap_test

import (
	"context"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/platform/bootstrap"
	"github.com/somralab/somra-media/internal/platform/db"
)

func TestWireRequests_WiresDependencies(t *testing.T) {
	dbCfg := db.Default()
	dbCfg.DataDir = filepath.Join(t.TempDir(), "data")

	c, err := bootstrap.NewWithStorage(context.Background(), nil, dbCfg)
	require.NoError(t, err)
	t.Cleanup(func() { _ = c.Close() })

	chRepo := db.NewNotificationChannelRepo(c.DB.Querier())
	_, err = chRepo.Create(context.Background(), db.NotificationChannel{
		ChannelType: db.NotificationChannelWebhook,
		Name:        "Ops",
		Config:      `{"url":"https://example.com/hook"}`,
		Enabled:     true,
	})
	require.NoError(t, err)

	t.Setenv("SOMRA_TMDB_API_KEY", "test-key")
	lib := bootstrap.WireLibrary(c)
	bundle, err := bootstrap.WireRequests(c, lib, func(*http.Request) string { return "en-US" })
	require.NoError(t, err)
	require.NotNil(t, bundle.Requests)
	require.NotNil(t, bundle.Notifications)
	require.NotNil(t, bundle.Dispatcher)
}

func TestWireRequests_RequiresDB(t *testing.T) {
	t.Parallel()
	_, err := bootstrap.WireRequests(&bootstrap.Components{}, nil, nil)
	require.Error(t, err)
}
