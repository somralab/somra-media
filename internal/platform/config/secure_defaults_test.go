package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecureDefaults_CORSRestrictsLocalDevOrigin(t *testing.T) {
	cfg := Default()
	require.NotEmpty(t, cfg.CORS.AllowedOrigins)
	for _, origin := range cfg.CORS.AllowedOrigins {
		assert.NotEqual(t, "*", origin, "wildcard CORS must not be the default")
	}
	assert.Contains(t, cfg.CORS.AllowedOrigins, "http://localhost:5173")
}

func TestSecureDefaults_CORSMethodsAreExplicit(t *testing.T) {
	cfg := Default()
	assert.Contains(t, cfg.CORS.AllowedMethods, "GET")
	assert.Contains(t, cfg.CORS.AllowedMethods, "OPTIONS")
}

func TestSecureDefaults_HTTPTimeoutsConfigured(t *testing.T) {
	cfg := Default()
	assert.Positive(t, cfg.HTTP.ReadHeaderTimeout)
	assert.Positive(t, cfg.HTTP.ReadTimeout)
	assert.Positive(t, cfg.HTTP.IdleTimeout)
}

func TestSecureDefaults_JWTSecretNotHardcoded(t *testing.T) {
	cfg := Default()
	assert.Empty(t, cfg.Auth.JWTSecret, "JWT secret must come from env in production")
}

func TestSecureDefaults_ProductionCORSOverride(t *testing.T) {
	env := map[string]string{
		"SOMRA_CORS_ORIGINS": "https://somra.example",
	}
	cfg, err := loadFrom(lookupFromMap(env))
	require.NoError(t, err)
	assert.Equal(t, []string{"https://somra.example"}, cfg.CORS.AllowedOrigins)
}
