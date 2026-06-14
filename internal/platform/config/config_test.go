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
