// Package api owns the HTTP gateway: chi router, middleware chain, and the
// skeleton handlers (health, version, SSE) used by the M1 milestone. Real
// business endpoints land in later sprints behind the same router.
package api

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	"github.com/somralab/somra-media/internal/auth"
	"github.com/somralab/somra-media/internal/platform/config"
	"github.com/somralab/somra-media/internal/settings"
)

// Options configures the router. All fields are optional; sensible defaults
// keep the router usable from tests without wiring every dependency.
type Options struct {
	Logger      *slog.Logger
	Build       BuildInfo
	CORS        config.CORSConfig
	RateLimiter RateLimiter
	HealthCheck []HealthCheck
	// HealthAggregator, when non-nil, enriches /api/v1/health with the
	// platform diagnostics registry. Test handlers can leave it nil and
	// rely on HealthCheck alone.
	HealthAggregator HealthAggregator

	// Now is the clock used by handlers; nil falls back to time.Now. Tests
	// inject a fixed time to assert on payload shape.
	Now func() time.Time

	// SSEHeartbeat overrides the default heartbeat interval. Useful for
	// tests that want a faster keep-alive cadence.
	SSEHeartbeat time.Duration

	// WebDir, when non-empty, points at a directory of static assets to
	// serve under "/". Used by the Docker image to ship the React SPA out
	// of the same binary; left empty during local dev (Vite proxies API).
	WebDir string

	// LocalizerMiddleware, when non-nil, is mounted into the chain so the
	// downstream handlers can resolve user-facing strings via the
	// negotiated locale. cmd/somra wires this from the i18n bundle so
	// error envelopes pick up Accept-Language; tests that don't care
	// about localization can leave it nil and receive raw message keys.
	LocalizerMiddleware func(http.Handler) http.Handler

	// EventBus broadcasts SSE events (scan progress, etc.).
	EventBus *EventBus

	// LibraryHandlers, when non-nil, mounts /libraries routes.
	LibraryHandlers *LibraryHandlers

	// MediaHandlers, when non-nil, mounts media/metadata routes.
	MediaHandlers *MediaHandlers

	// BrowseHandlers mounts discover/search/paginated browse/detail routes.
	BrowseHandlers *BrowseHandlers

	// AuthHandlers, when non-nil, mounts auth/setup routes (public subset).
	AuthHandlers *AuthHandlers

	// AuthMiddleware validates bearer tokens on protected routes.
	AuthMiddleware *AuthMiddleware

	// UserHandlers mounts admin user CRUD (protected).
	UserHandlers *UserHandlers

	// ProfileHandlers mounts profile CRUD (protected).
	ProfileHandlers *ProfileHandlers

	// WatchHandlers mounts watch state / favorites / watchlist (protected).
	WatchHandlers *WatchHandlers

	// StreamingHandlers mounts playback/streaming routes (protected).
	StreamingHandlers *StreamingHandlers

	// SystemHandlers mounts system detection routes (public).
	SystemHandlers *SystemHandlers

	// SettingsHandlers mounts central settings API (admin).
	SettingsHandlers *SettingsHandlers

	// OnboardingHandlers mounts wizard state endpoints.
	OnboardingHandlers *OnboardingHandlers

	// SubtitleHandlers mounts subtitle management endpoints.
	SubtitleHandlers *SubtitleHandlers

	// RequestHandlers mounts content request workflow endpoints.
	RequestHandlers *RequestHandlers

	// NotificationHandlers mounts notification preference/channel endpoints.
	NotificationHandlers *NotificationHandlers

	// PluginHandlers mounts plugin catalog/instance management endpoints.
	PluginHandlers *PluginHandlers

	// AutomationHandlers mounts indexer search, downloads, and quality profiles.
	AutomationHandlers *AutomationHandlers

	// Onboarding, when set, is notified after first admin creation.
	Onboarding *settings.Onboarding
}

// New returns a chi router with the canonical middleware chain mounted at
// /api/v1.
func New(opts Options) http.Handler {
	r := chi.NewRouter()

	if opts.RateLimiter == nil {
		opts.RateLimiter = NoopRateLimiter{}
	}

	r.Use(RequestIDMiddleware)
	r.Use(RecoverMiddleware)
	r.Use(LoggerMiddleware(opts.Logger))
	r.Use(corsMiddleware(opts.CORS))
	r.Use(RateLimitMiddleware(opts.RateLimiter))
	r.Use(ContentTypeMiddleware)
	if opts.LocalizerMiddleware != nil {
		r.Use(opts.LocalizerMiddleware)
	}

	r.NotFound(func(w http.ResponseWriter, req *http.Request) {
		writeError(w, req, errNotFound)
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, req *http.Request) {
		writeError(w, req, errMethodNotAllow)
	})

	r.Route("/api/v1", func(api chi.Router) {
		api.Get("/health", healthHandler(opts.Now, opts.HealthCheck, opts.HealthAggregator))
		api.Get("/version", versionHandler(opts.Build, opts.Now))
		api.Get("/events/stream", sseEventsHandler(opts.SSEHeartbeat, opts.EventBus))

		if opts.AuthHandlers != nil {
			opts.AuthHandlers.Mount(api)
		}
		if opts.SystemHandlers != nil {
			opts.SystemHandlers.Mount(api)
		}
		if opts.OnboardingHandlers != nil {
			opts.OnboardingHandlers.Mount(api)
		}

		if opts.AuthMiddleware != nil {
			api.Group(func(protected chi.Router) {
				protected.Use(opts.AuthMiddleware.Middleware)
				protected.Use(ProfileLocaleMiddleware)
				mountProtectedRoutes(protected, opts)
			})
		} else {
			mountProtectedRoutes(api, opts)
		}
	})

	if opts.WebDir != "" {
		mountSPA(r, opts.WebDir, opts.Logger)
	}

	return r
}

func mountProtectedRoutes(r chi.Router, opts Options) {
	if opts.ProfileHandlers != nil {
		r.Group(func(r chi.Router) {
			r.Use(RequirePermission(auth.PermProfileEdit))
			opts.ProfileHandlers.Mount(r)
		})
	}
	if opts.WatchHandlers != nil {
		opts.WatchHandlers.Mount(r)
	}
	if opts.UserHandlers != nil {
		r.Group(func(r chi.Router) {
			r.Use(RequirePermission(auth.PermUsersManage))
			opts.UserHandlers.Mount(r)
		})
	}
	if opts.LibraryHandlers != nil {
		opts.LibraryHandlers.Mount(r)
	}
	if opts.MediaHandlers != nil {
		opts.MediaHandlers.Mount(r)
	}
	if opts.BrowseHandlers != nil {
		opts.BrowseHandlers.Mount(r)
	}
	if opts.StreamingHandlers != nil {
		opts.StreamingHandlers.Mount(r)
	}
	if opts.SettingsHandlers != nil {
		opts.SettingsHandlers.Mount(r)
	}
	if opts.OnboardingHandlers != nil {
		opts.OnboardingHandlers.MountProtected(r)
	}
	if opts.SubtitleHandlers != nil {
		opts.SubtitleHandlers.Mount(r)
	}
	if opts.RequestHandlers != nil {
		opts.RequestHandlers.Mount(r)
	}
	if opts.NotificationHandlers != nil {
		opts.NotificationHandlers.Mount(r)
	}
	if opts.PluginHandlers != nil {
		opts.PluginHandlers.Mount(r)
	}
	if opts.AutomationHandlers != nil {
		opts.AutomationHandlers.Mount(r)
	}
}

// mountSPA serves the built React SPA from dir. It serves static assets
// directly and falls back to index.html for any non-API path so that
// client-side routes resolve on hard reloads. The handler is silently
// skipped when dir does not exist, allowing the same image to be used in
// API-only deployments.
func mountSPA(r chi.Router, dir string, logger *slog.Logger) {
	clean := filepath.Clean(dir)
	info, err := os.Stat(clean)
	if err != nil || !info.IsDir() {
		if logger != nil {
			logger.Warn("spa directory not available",
				slog.String("dir", clean),
				slog.Any("error", err),
			)
		}
		return
	}

	indexPath := filepath.Join(clean, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		if logger != nil {
			logger.Warn("spa index.html missing",
				slog.String("dir", clean),
				slog.Any("error", err),
			)
		}
		return
	}

	fs := http.FileServer(http.Dir(clean))
	spaHandler := func(w http.ResponseWriter, req *http.Request) {
		urlPath := req.URL.Path
		if urlPath == "" || urlPath == "/" {
			serveIndex(w, req, indexPath)
			return
		}
		if strings.HasPrefix(urlPath, "/api/") {
			writeError(w, req, errNotFound)
			return
		}

		assetPath := filepath.Join(clean, filepath.FromSlash(strings.TrimPrefix(urlPath, "/")))
		if rel, err := filepath.Rel(clean, assetPath); err != nil || strings.HasPrefix(rel, "..") {
			writeError(w, req, errNotFound)
			return
		}
		if stat, err := os.Stat(assetPath); err == nil && !stat.IsDir() {
			fs.ServeHTTP(w, req)
			return
		}
		serveIndex(w, req, indexPath)
	}
	r.Get("/*", spaHandler)
	r.Head("/*", spaHandler)
}

func serveIndex(w http.ResponseWriter, r *http.Request, indexPath string) {
	w.Header().Set("Cache-Control", "no-cache")
	http.ServeFile(w, r, indexPath)
}

// corsMiddleware adapts the go-chi/cors module to our config struct.
func corsMiddleware(cfg config.CORSConfig) func(http.Handler) http.Handler {
	options := cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   cfg.AllowedMethods,
		AllowedHeaders:   cfg.AllowedHeaders,
		ExposedHeaders:   []string{requestIDHeader},
		AllowCredentials: true,
		MaxAge:           int(cfg.MaxAge.Seconds()),
	}
	return cors.Handler(options)
}
