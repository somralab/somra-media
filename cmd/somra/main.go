// Command somra is the single Go binary that hosts every Somra module. In
// Sprint 01 (M1) it only serves the API gateway skeleton; later sprints
// register their own services against the same bootstrap.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/somralab/somra-media/internal/api"
	"github.com/somralab/somra-media/internal/platform/bootstrap"
	"github.com/somralab/somra-media/internal/platform/config"
	"github.com/somralab/somra-media/internal/platform/db"
	i18npkg "github.com/somralab/somra-media/internal/platform/i18n"
	platformlog "github.com/somralab/somra-media/internal/platform/log"
)

// Build identifiers populated at link time via -ldflags. Defaults keep local
// development builds informative.
var (
	version = "0.1.0-dev"
	commit  = ""
	builtAt = ""
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "somra: fatal: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger, err := platformlog.New(platformlog.Options{
		Level:  cfg.Log.Level,
		Format: cfg.Log.Format,
	})
	if err != nil {
		return fmt.Errorf("build logger: %w", err)
	}
	slog.SetDefault(logger)

	logger.Info("somra starting",
		slog.String("version", version),
		slog.String("commit", commit),
		slog.String("addr", cfg.HTTP.Addr),
	)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	dbCfg := db.Default()
	if cfg.Data.Dir != "" {
		dbCfg.DataDir = cfg.Data.Dir
	}
	components, err := bootstrap.NewWithStorage(ctx, logger, dbCfg)
	if err != nil {
		return fmt.Errorf("bootstrap platform: %w", err)
	}
	defer func() {
		if cerr := components.Close(); cerr != nil {
			logger.Error("close components", slog.Any("error", cerr))
		}
	}()

	components.Scheduler.Start(ctx)
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), cfg.Shutdown.Timeout)
		defer cancel()
		if serr := components.Scheduler.Stop(stopCtx); serr != nil {
			logger.Error("stop scheduler", slog.Any("error", serr))
		}
	}()

	libBundle := bootstrap.WireLibrary(components)
	authBundle, err := bootstrap.WireAuth(components, cfg.Auth, logger)
	if err != nil {
		return fmt.Errorf("bootstrap auth: %w", err)
	}
	settingsBundle := bootstrap.WireSettings(components, authBundle.Service)
	streamBundle := bootstrap.WireStreaming(components, cfg, logger)
	subtitlesBundle := bootstrap.WireSubtitles(components, cfg, settingsBundle.Settings, logger)

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
		Logger: logger,
		Build: api.BuildInfo{
			Version: version,
			Commit:  commit,
			BuiltAt: builtAt,
		},
		CORS:                cfg.CORS,
		WebDir:              cfg.Web.Dir,
		LocalizerMiddleware: components.I18n.Middleware(),
		HealthAggregator:    api.NewDiagnosticsAggregator(components.Diagnostics),
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
			Profiles: db.NewProfileRepo(components.DB.Querier()),
		},
		WatchHandlers: &api.WatchHandlers{
			Watch: db.NewWatchRepo(components.DB.Querier()),
		},
		LibraryHandlers: &api.LibraryHandlers{
			Service:    libBundle.Library,
			Locale:     localeFn,
			Onboarding: settingsBundle.Onboarding,
		},
		MediaHandlers: &api.MediaHandlers{
			DB:       components.DB,
			Metadata: libBundle.Metadata,
			Locale:   localeFn,
		},
		BrowseHandlers: &api.BrowseHandlers{
			Browse: db.NewBrowseRepo(components.DB.Querier()),
			Locale: localeFn,
		},
	}
	apiOpts.SystemHandlers = &api.SystemHandlers{
		DataDir:  cfg.Data.Dir,
		CacheDir: cfg.Data.CacheDir,
	}
	apiOpts.SettingsHandlers = &api.SettingsHandlers{Service: settingsBundle.Settings}
	apiOpts.OnboardingHandlers = &api.OnboardingHandlers{Onboarding: settingsBundle.Onboarding}
	if subtitlesBundle != nil && subtitlesBundle.Service != nil {
		apiOpts.SubtitleHandlers = &api.SubtitleHandlers{Service: subtitlesBundle.Service}
	}
	if streamBundle != nil && streamBundle.Service != nil {
		apiOpts.StreamingHandlers = &api.StreamingHandlers{
			Streaming: streamBundle.Service,
			Media:     db.NewMediaRepo(components.DB.Querier()),
			Library:   db.NewLibraryRepo(components.DB.Querier()),
			Playback:  db.NewPlaybackRepo(components.DB.Querier()),
			CacheRoot: cfg.Data.CacheDir,
		}
	}
	handler := api.New(apiOpts)

	srv := &http.Server{
		Addr:              cfg.HTTP.Addr,
		Handler:           handler,
		ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout,
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		IdleTimeout:       cfg.HTTP.IdleTimeout,
		BaseContext: func(_ net.Listener) context.Context {
			return platformlog.WithLogger(context.Background(), logger)
		},
	}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("http server listening", slog.String("addr", cfg.HTTP.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received", slog.Duration("grace", cfg.Shutdown.Timeout))
	case err := <-serverErr:
		if err != nil {
			return fmt.Errorf("http listen: %w", err)
		}
		return nil
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Shutdown.Timeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", slog.Any("error", err))
		return fmt.Errorf("shutdown: %w", err)
	}

	if err := <-serverErr; err != nil {
		return fmt.Errorf("server exit: %w", err)
	}

	logger.Info("somra stopped cleanly")
	return nil
}
