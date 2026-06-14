package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func lookupFromMap(m map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		v, ok := m[key]
		return v, ok
	}
}

func TestDefault(t *testing.T) {
	cfg := Default()
	assert.Equal(t, ":8080", cfg.HTTP.Addr)
	assert.Equal(t, "info", cfg.Log.Level)
	assert.Equal(t, "json", cfg.Log.Format)
	assert.Equal(t, 10*time.Second, cfg.Shutdown.Timeout)
	assert.Equal(t, "./data", cfg.Data.Dir)
	assert.Empty(t, cfg.Web.Dir)
	assert.NotEmpty(t, cfg.CORS.AllowedOrigins)
}

func TestLoadFrom_Overrides(t *testing.T) {
	env := map[string]string{
		"SOMRA_HTTP_ADDR":        "127.0.0.1:9090",
		"SOMRA_LOG_LEVEL":        "DEBUG",
		"SOMRA_LOG_FORMAT":       "TEXT",
		"SOMRA_CORS_ORIGINS":     "https://a.example, https://b.example",
		"SOMRA_CORS_MAX_AGE":     "60s",
		"SOMRA_DATA_DIR":         "/var/lib/somra",
		"SOMRA_WEB_DIR":          "/web/static",
		"SOMRA_SHUTDOWN_TIMEOUT": "15",
	}
	cfg, err := loadFrom(lookupFromMap(env))
	require.NoError(t, err)
	assert.Equal(t, "127.0.0.1:9090", cfg.HTTP.Addr)
	assert.Equal(t, "debug", cfg.Log.Level)
	assert.Equal(t, "text", cfg.Log.Format)
	assert.Equal(t, []string{"https://a.example", "https://b.example"}, cfg.CORS.AllowedOrigins)
	assert.Equal(t, 60*time.Second, cfg.CORS.MaxAge)
	assert.Equal(t, "/var/lib/somra", cfg.Data.Dir)
	assert.Equal(t, "/web/static", cfg.Web.Dir)
	assert.Equal(t, 15*time.Second, cfg.Shutdown.Timeout)
}

func TestLoadFrom_InvalidDuration(t *testing.T) {
	env := map[string]string{"SOMRA_SHUTDOWN_TIMEOUT": "not-a-duration"}
	_, err := loadFrom(lookupFromMap(env))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SOMRA_SHUTDOWN_TIMEOUT")
}

func TestSplitCSV_TrimsAndDropsEmpty(t *testing.T) {
	got := splitCSV(" a , , b ,c ")
	assert.Equal(t, []string{"a", "b", "c"}, got)
}

func TestLoad(t *testing.T) {
	t.Setenv("SOMRA_HTTP_ADDR", "127.0.0.1:9090")
	t.Setenv("SOMRA_DATA_DIR", "/var/lib/somra")
	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "127.0.0.1:9090", cfg.HTTP.Addr)
	assert.Equal(t, "/var/lib/somra", cfg.Data.Dir)
}

func TestLoadFrom_StreamingAndAuth(t *testing.T) {
	env := map[string]string{
		"SOMRA_CACHE_DIR":                "/tmp/cache",
		"SOMRA_STREAMING_MAX_CONCURRENT": "2",
		"SOMRA_STREAMING_MAX_QUEUE":      "5",
		"SOMRA_STREAMING_SESSION_TTL":    "1h",
		"SOMRA_STREAMING_IDLE_TIMEOUT":   "30m",
		"SOMRA_FFMPEG_BIN":               "/usr/bin/ffmpeg",
		"SOMRA_FFPROBE_BIN":              "/usr/bin/ffprobe",
		"SOMRA_JWT_SECRET":               "secret",
		"SOMRA_REFRESH_PEPPER":           "pepper",
		"SOMRA_JWT_ACCESS_TTL":           "15m",
		"SOMRA_JWT_REFRESH_TTL":          "168h",
		"SOMRA_AUTH_SECURE_COOKIE":       "true",
		"SOMRA_HTTP_READ_HEADER_TIMEOUT": "5s",
		"SOMRA_HTTP_READ_TIMEOUT":        "10s",
		"SOMRA_HTTP_WRITE_TIMEOUT":       "20s",
		"SOMRA_HTTP_IDLE_TIMEOUT":        "30s",
	}
	cfg, err := loadFrom(lookupFromMap(env))
	require.NoError(t, err)
	assert.Equal(t, "/tmp/cache", cfg.Data.CacheDir)
	assert.Equal(t, 2, cfg.Streaming.MaxConcurrentTranscodes)
	assert.Equal(t, 5, cfg.Streaming.MaxTranscodeQueue)
	assert.Equal(t, time.Hour, cfg.Streaming.SessionTTL)
	assert.Equal(t, 30*time.Minute, cfg.Streaming.IdleTimeout)
	assert.Equal(t, "/usr/bin/ffmpeg", cfg.Streaming.FFmpegBin)
	assert.Equal(t, "/usr/bin/ffprobe", cfg.Streaming.FFprobeBin)
	assert.Equal(t, "secret", cfg.Auth.JWTSecret)
	assert.Equal(t, "pepper", cfg.Auth.RefreshPepper)
	assert.Equal(t, 15*time.Minute, cfg.Auth.AccessTTL)
	assert.True(t, cfg.Auth.SecureCookie)
	assert.Equal(t, 5*time.Second, cfg.HTTP.ReadHeaderTimeout)
}
