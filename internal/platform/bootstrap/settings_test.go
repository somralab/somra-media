package bootstrap_test

import (
	"context"
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/auth"
	"github.com/somralab/somra-media/internal/platform/bootstrap"
	"github.com/somralab/somra-media/internal/platform/config"
	"github.com/somralab/somra-media/internal/platform/db"
)

func TestWireSettingsAndSubtitles(t *testing.T) {
	ctx := context.Background()
	dbCfg := db.Default()
	dbCfg.DataDir = filepath.Join(t.TempDir(), "data")

	c, err := bootstrap.NewWithStorage(ctx, nil, dbCfg)
	require.NoError(t, err)
	t.Cleanup(func() { _ = c.Close() })

	secret := []byte("test-secret-key-at-least-32-bytes!!")
	jwtCfg := auth.DefaultJWTConfig(secret)
	q := c.DB.Querier()
	authSvc := auth.NewService(auth.ServiceConfig{
		Users:    db.NewUserRepo(q),
		Sessions: db.NewSessionRepo(q),
		Profiles: db.NewProfileRepo(q),
		Tokens:   auth.NewJWTService(jwtCfg),
		Refresh:  auth.NewSQLiteRefreshStore(db.NewSessionRepo(q), secret, jwtCfg.RefreshTTL),
		Hasher:   auth.NewPasswordHasher(auth.DefaultPasswordPolicy()),
		Lockout:  auth.NewLoginLockout(db.NewLoginAttemptRepo(q), auth.DefaultLockoutConfig()),
		JWT:      jwtCfg,
	})

	settingsBundle := bootstrap.WireSettings(c, authSvc)
	require.NotNil(t, settingsBundle)
	require.NotNil(t, settingsBundle.Settings)
	require.NotNil(t, settingsBundle.Onboarding)

	state, err := settingsBundle.Onboarding.Status(ctx)
	require.NoError(t, err)
	assert.Equal(t, "language", state.Phase)

	cfg := config.Default()
	cfg.Data.Dir = dbCfg.DataDir
	cfg.Data.CacheDir = filepath.Join(dbCfg.DataDir, "cache")

	subBundle := bootstrap.WireSubtitles(c, cfg, settingsBundle.Settings, slog.Default())
	require.NotNil(t, subBundle)
	require.NotNil(t, subBundle.Service)

	assert.Nil(t, bootstrap.WireSubtitles(nil, cfg, settingsBundle.Settings, slog.Default()))
}
