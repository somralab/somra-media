// Package config loads runtime configuration from environment variables with
// safe defaults. Convention over configuration: every setting has a usable
// default so the binary can start with zero ops input.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config aggregates all runtime configuration for the Somra binary.
//
// Fields are grouped per concern (HTTP, logging, CORS, data, shutdown) so that
// later packets (DB, jobs, streaming) extend rather than mutate the contract.
type Config struct {
	HTTP      HTTPConfig
	Log       LogConfig
	CORS      CORSConfig
	Data      DataConfig
	Web       WebConfig
	Auth      AuthConfig
	Streaming StreamingConfig
	Shutdown  ShutdownConfig
}

// HTTPConfig describes the API gateway listener.
type HTTPConfig struct {
	// Addr is the listen address, e.g. ":8080" or "127.0.0.1:8080".
	Addr string
	// ReadHeaderTimeout caps the time spent reading request headers; protects
	// the server from slowloris-style attacks.
	ReadHeaderTimeout time.Duration
	// ReadTimeout caps total request body read time.
	ReadTimeout time.Duration
	// WriteTimeout caps response write time. SSE streams use hijacking so this
	// value does not bound long-lived event connections.
	WriteTimeout time.Duration
	// IdleTimeout for keep-alive connections.
	IdleTimeout time.Duration
}

// LogConfig controls the structured logger.
type LogConfig struct {
	// Level is one of "debug", "info", "warn", "error".
	Level string
	// Format is one of "json", "text".
	Format string
}

// CORSConfig controls the CORS middleware. Defaults allow only the local SPA
// dev origin; production deployments override via environment.
type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
	MaxAge         time.Duration
}

// DataConfig points at the on-disk data directory (SQLite, caches). Used by
// later packets; declared here so all configuration lives in one struct.
type DataConfig struct {
	Dir      string
	CacheDir string
}

// StreamingConfig controls transcode session limits and idle cleanup.
type StreamingConfig struct {
	MaxConcurrentTranscodes int
	MaxTranscodeQueue       int
	SessionTTL              time.Duration
	IdleTimeout             time.Duration
	FFmpegBin               string
	FFprobeBin              string
}

// WebConfig points at an optional directory of pre-built SPA assets that the
// Go binary serves alongside the API. Empty means "no SPA"; the Docker image
// sets it to the location where the build pipeline copied the bundle.
type WebConfig struct {
	Dir string
}

// AuthConfig holds JWT and refresh-token settings.
type AuthConfig struct {
	JWTSecret     string
	RefreshPepper string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
	SecureCookie  bool
}

// ShutdownConfig bounds the graceful shutdown window.
type ShutdownConfig struct {
	Timeout time.Duration
}

// Default returns a Config populated with defaults suitable for local
// development and zero-configuration container start.
func Default() Config {
	return Config{
		HTTP: HTTPConfig{
			Addr:              ":8080",
			ReadHeaderTimeout: 10 * time.Second,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      0,
			IdleTimeout:       120 * time.Second,
		},
		Log: LogConfig{
			Level:  "info",
			Format: "json",
		},
		CORS: CORSConfig{
			AllowedOrigins: []string{"http://localhost:5173"},
			AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-Request-Id", "Accept-Language"},
			MaxAge:         5 * time.Minute,
		},
		Data: DataConfig{
			Dir:      "./data",
			CacheDir: "./cache",
		},
		Streaming: StreamingConfig{
			MaxConcurrentTranscodes: 2,
			MaxTranscodeQueue:       8,
			SessionTTL:              4 * time.Hour,
			IdleTimeout:             15 * time.Minute,
			FFmpegBin:               "ffmpeg",
			FFprobeBin:              "ffprobe",
		},
		Web: WebConfig{
			Dir: "",
		},
		Auth: AuthConfig{
			AccessTTL:  15 * time.Minute,
			RefreshTTL: 7 * 24 * time.Hour,
		},
		Shutdown: ShutdownConfig{
			Timeout: 10 * time.Second,
		},
	}
}

// Load builds a Config starting from Default and overlaying any SOMRA_*
// environment variables that are present. Invalid values return an error so
// misconfiguration is surfaced early at bootstrap rather than at request time.
func Load() (Config, error) {
	return loadFrom(os.LookupEnv)
}

// loadFrom is the testable seam behind Load; it accepts an arbitrary lookup
// function so callers can inject deterministic environments in unit tests.
func loadFrom(lookup func(string) (string, bool)) (Config, error) {
	cfg := Default()

	if v, ok := lookup("SOMRA_HTTP_ADDR"); ok {
		cfg.HTTP.Addr = v
	} else if v, ok := lookup("SOMRA_LISTEN_ADDR"); ok {
		cfg.HTTP.Addr = v
	}
	if err := durationFromEnv(lookup, "SOMRA_HTTP_READ_HEADER_TIMEOUT", &cfg.HTTP.ReadHeaderTimeout); err != nil {
		return Config{}, err
	}
	if err := durationFromEnv(lookup, "SOMRA_HTTP_READ_TIMEOUT", &cfg.HTTP.ReadTimeout); err != nil {
		return Config{}, err
	}
	if err := durationFromEnv(lookup, "SOMRA_HTTP_WRITE_TIMEOUT", &cfg.HTTP.WriteTimeout); err != nil {
		return Config{}, err
	}
	if err := durationFromEnv(lookup, "SOMRA_HTTP_IDLE_TIMEOUT", &cfg.HTTP.IdleTimeout); err != nil {
		return Config{}, err
	}

	if v, ok := lookup("SOMRA_LOG_LEVEL"); ok {
		cfg.Log.Level = strings.ToLower(v)
	}
	if v, ok := lookup("SOMRA_LOG_FORMAT"); ok {
		cfg.Log.Format = strings.ToLower(v)
	}

	if v, ok := lookup("SOMRA_CORS_ORIGINS"); ok {
		cfg.CORS.AllowedOrigins = splitCSV(v)
	}
	if err := durationFromEnv(lookup, "SOMRA_CORS_MAX_AGE", &cfg.CORS.MaxAge); err != nil {
		return Config{}, err
	}

	if v, ok := lookup("SOMRA_DATA_DIR"); ok {
		cfg.Data.Dir = v
	}
	if v, ok := lookup("SOMRA_CACHE_DIR"); ok {
		cfg.Data.CacheDir = v
	}

	if v, ok := lookup("SOMRA_STREAMING_MAX_CONCURRENT"); ok {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.Streaming.MaxConcurrentTranscodes = n
		}
	}
	if v, ok := lookup("SOMRA_STREAMING_MAX_QUEUE"); ok {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.Streaming.MaxTranscodeQueue = n
		}
	}
	if err := durationFromEnv(lookup, "SOMRA_STREAMING_SESSION_TTL", &cfg.Streaming.SessionTTL); err != nil {
		return Config{}, err
	}
	if err := durationFromEnv(lookup, "SOMRA_STREAMING_IDLE_TIMEOUT", &cfg.Streaming.IdleTimeout); err != nil {
		return Config{}, err
	}
	if v, ok := lookup("SOMRA_FFMPEG_BIN"); ok {
		cfg.Streaming.FFmpegBin = v
	}
	if v, ok := lookup("SOMRA_FFPROBE_BIN"); ok {
		cfg.Streaming.FFprobeBin = v
	}

	if v, ok := lookup("SOMRA_WEB_DIR"); ok {
		cfg.Web.Dir = v
	}

	if v, ok := lookup("SOMRA_JWT_SECRET"); ok {
		cfg.Auth.JWTSecret = v
	}
	if v, ok := lookup("SOMRA_REFRESH_PEPPER"); ok {
		cfg.Auth.RefreshPepper = v
	}
	if err := durationFromEnv(lookup, "SOMRA_JWT_ACCESS_TTL", &cfg.Auth.AccessTTL); err != nil {
		return Config{}, err
	}
	if err := durationFromEnv(lookup, "SOMRA_JWT_REFRESH_TTL", &cfg.Auth.RefreshTTL); err != nil {
		return Config{}, err
	}
	if v, ok := lookup("SOMRA_AUTH_SECURE_COOKIE"); ok {
		cfg.Auth.SecureCookie = strings.EqualFold(v, "true") || v == "1"
	}

	if err := durationFromEnv(lookup, "SOMRA_SHUTDOWN_TIMEOUT", &cfg.Shutdown.Timeout); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func durationFromEnv(lookup func(string) (string, bool), key string, dst *time.Duration) error {
	raw, ok := lookup(key)
	if !ok {
		return nil
	}
	if d, err := time.ParseDuration(raw); err == nil {
		*dst = d
		return nil
	}
	if secs, err := strconv.Atoi(raw); err == nil {
		*dst = time.Duration(secs) * time.Second
		return nil
	}
	return fmt.Errorf("config: invalid duration for %s=%q", key, raw)
}

func splitCSV(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}
