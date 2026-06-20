package bootstrap

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/api"
	"github.com/somralab/somra-media/internal/platform/config"
	"github.com/somralab/somra-media/internal/platform/db"
	i18npkg "github.com/somralab/somra-media/internal/platform/i18n"
	"github.com/somralab/somra-media/internal/settings"
)

// TestServer mirrors cmd/somra wiring for integration tests.
type TestServer struct {
	Server     *httptest.Server
	Handler    http.Handler
	Components *Components
	Plugins    *PluginsBundle
	Auth       *AuthBundle
	Settings   *SettingsBundle
	Library    *LibraryBundle
	AdminToken string
	LocaleFn   func(*http.Request) string
}

// NewTestServer boots a fully wired API stack in a hermetic temp data directory.
func NewTestServer(t *testing.T) *TestServer {
	t.Helper()
	t.Setenv("SOMRA_DATA_DIR", filepath.Join(t.TempDir(), "data"))

	cfg, err := config.Load()
	require.NoError(t, err)

	bootCtx, bootCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer bootCancel()

	dbCfg := db.Default()
	dbCfg.DataDir = cfg.Data.Dir

	c, err := NewWithStorage(bootCtx, nil, dbCfg)
	require.NoError(t, err)
	c.Scheduler.Start(context.Background())

	libBundle := WireLibrary(c)
	authBundle, err := WireAuth(c, cfg.Auth, c.Logger)
	require.NoError(t, err)
	settingsBundle := WireSettings(c, authBundle.Service)
	streamBundle := WireStreaming(c, cfg, c.Logger)
	subtitlesBundle := WireSubtitles(c, cfg, settingsBundle.Settings, c.Logger)
	pluginsBundle, err := WirePlugins(c, cfg.Auth.JWTSecret)
	require.NoError(t, err)

	localeFn := func(r *http.Request) string {
		if loc, ok := api.AcceptLanguageFromContext(r.Context()); ok && loc != "" {
			return loc
		}
		if loc := i18npkg.FromContext(r.Context()); loc != nil {
			return loc.Tag().String()
		}
		return "en-US"
	}

	authMW := &api.AuthMiddleware{Service: authBundle.Service}
	apiOpts := api.Options{
		Logger:              c.Logger,
		Build:               api.BuildInfo{Version: "0.0.0-test", Commit: "test", BuiltAt: "1970-01-01T00:00:00Z"},
		CORS:                cfg.CORS,
		LocalizerMiddleware: c.I18n.Middleware(),
		HealthAggregator:    api.NewDiagnosticsAggregator(c.Diagnostics),
		EventBus:            libBundle.EventBus,
		AuthHandlers: &api.AuthHandlers{
			Service:      authBundle.Service,
			SecureCookie: cfg.Auth.SecureCookie,
			Onboarding:   settingsBundle.Onboarding,
		},
		AuthMiddleware: authMW,
		UserHandlers: &api.UserHandlers{
			Service: authBundle.Service,
			Users:   authBundle.Users,
		},
		ProfileHandlers: &api.ProfileHandlers{
			Profiles: db.NewProfileRepo(c.DB.Querier()),
		},
		WatchHandlers: &api.WatchHandlers{
			Watch: db.NewWatchRepo(c.DB.Querier()),
		},
		LibraryHandlers: &api.LibraryHandlers{
			Service:    libBundle.Library,
			Locale:     localeFn,
			Onboarding: settingsBundle.Onboarding,
		},
		MediaHandlers: &api.MediaHandlers{
			DB:       c.DB,
			Metadata: libBundle.Metadata,
			Locale:   localeFn,
		},
		BrowseHandlers: &api.BrowseHandlers{
			Browse: db.NewBrowseRepo(c.DB.Querier()),
			Locale: localeFn,
		},
		SystemHandlers: &api.SystemHandlers{
			DataDir:   cfg.Data.Dir,
			CacheDir:  cfg.Data.CacheDir,
			FFmpegBin: cfg.Streaming.FFmpegBin,
		},
		SettingsHandlers: &api.SettingsHandlers{
			Service: settingsBundle.Settings,
			OnPatched: func(ctx context.Context, category string) error {
				if streamBundle == nil || streamBundle.Service == nil {
					return nil
				}
				if category == settings.CategoryPlayback || category == settings.CategoryStreaming {
					return SyncStreamingSettings(ctx, settingsBundle.Settings, streamBundle.Service, cfg.Streaming.FFmpegBin)
				}
				return nil
			},
		},
		OnboardingHandlers: &api.OnboardingHandlers{Onboarding: settingsBundle.Onboarding},
	}
	if subtitlesBundle != nil && subtitlesBundle.Service != nil {
		apiOpts.SubtitleHandlers = &api.SubtitleHandlers{Service: subtitlesBundle.Service}
	}
	if streamBundle != nil && streamBundle.Service != nil {
		apiOpts.StreamingHandlers = &api.StreamingHandlers{
			Streaming: streamBundle.Service,
			Media:     db.NewMediaRepo(c.DB.Querier()),
			Library:   db.NewLibraryRepo(c.DB.Querier()),
			Playback:  db.NewPlaybackRepo(c.DB.Querier()),
			CacheRoot: cfg.Data.CacheDir,
		}
	}
	requestsBundle, err := WireRequests(c, libBundle, localeFn)
	require.NoError(t, err)
	automationBundle, err := WireAutomation(c, libBundle, pluginsBundle, requestsBundle)
	require.NoError(t, err)
	apiOpts.RequestHandlers = requestsBundle.Requests
	apiOpts.NotificationHandlers = requestsBundle.Notifications
	apiOpts.PluginHandlers = &api.PluginHandlers{Manager: pluginsBundle.Manager}
	apiOpts.AutomationHandlers = automationBundle.Handlers

	handler := api.New(apiOpts)
	srv := httptest.NewServer(handler)

	setupBody, _ := json.Marshal(map[string]string{"username": "admin", "password": "AdminPass1"})
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/setup/admin", bytes.NewReader(setupBody))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var tok map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &tok))
	adminToken := tok["accessToken"].(string)

	ts := &TestServer{
		Server:     srv,
		Handler:    handler,
		Components: c,
		Plugins:    pluginsBundle,
		Auth:       authBundle,
		Settings:   settingsBundle,
		Library:    libBundle,
		AdminToken: adminToken,
		LocaleFn:   localeFn,
	}

	t.Cleanup(func() {
		srv.Close()
		stopCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = c.Scheduler.Stop(stopCtx)
		_ = c.Close()
	})

	return ts
}
