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
	"github.com/somralab/somra-media/internal/platform/config"
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

	handler := api.New(api.Options{
		Logger: logger,
		Build: api.BuildInfo{
			Version: version,
			Commit:  commit,
			BuiltAt: builtAt,
		},
		CORS: cfg.CORS,
	})

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

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

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
