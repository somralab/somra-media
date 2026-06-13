// Package api owns the HTTP gateway: chi router, middleware chain, and the
// skeleton handlers (health, version, SSE) used by the M1 milestone. Real
// business endpoints land in later sprints behind the same router.
package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	"github.com/somralab/somra-media/internal/platform/config"
)

// Options configures the router. All fields are optional; sensible defaults
// keep the router usable from tests without wiring every dependency.
type Options struct {
	Logger      *slog.Logger
	Build       BuildInfo
	CORS        config.CORSConfig
	RateLimiter RateLimiter
	HealthCheck []HealthCheck

	// Now is the clock used by handlers; nil falls back to time.Now. Tests
	// inject a fixed time to assert on payload shape.
	Now func() time.Time

	// SSEHeartbeat overrides the default heartbeat interval. Useful for
	// tests that want a faster keep-alive cadence.
	SSEHeartbeat time.Duration
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

	r.NotFound(func(w http.ResponseWriter, req *http.Request) {
		writeError(w, req, errNotFound)
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, req *http.Request) {
		writeError(w, req, errMethodNotAllow)
	})

	r.Route("/api/v1", func(api chi.Router) {
		api.Get("/health", healthHandler(opts.Now, opts.HealthCheck))
		api.Get("/version", versionHandler(opts.Build, opts.Now))
		api.Get("/events/stream", sseEventsHandler(opts.SSEHeartbeat))
	})

	return r
}

// corsMiddleware adapts the go-chi/cors module to our config struct.
func corsMiddleware(cfg config.CORSConfig) func(http.Handler) http.Handler {
	options := cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   cfg.AllowedMethods,
		AllowedHeaders:   cfg.AllowedHeaders,
		ExposedHeaders:   []string{requestIDHeader},
		AllowCredentials: false,
		MaxAge:           int(cfg.MaxAge.Seconds()),
	}
	return cors.Handler(options)
}
