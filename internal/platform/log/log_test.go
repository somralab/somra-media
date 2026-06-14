package log

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_JSONOutput(t *testing.T) {
	var buf bytes.Buffer
	logger, err := New(Options{Level: "info", Format: "json", Output: &buf})
	require.NoError(t, err)
	logger.Info("hello", slog.String("key", "value"))

	var entry map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &entry))
	assert.Equal(t, "hello", entry["msg"])
	assert.Equal(t, "value", entry["key"])
}

func TestNew_UnknownFormat(t *testing.T) {
	_, err := New(Options{Level: "info", Format: "yaml"})
	require.Error(t, err)
}

func TestNew_UnknownLevel(t *testing.T) {
	_, err := New(Options{Level: "spammy", Format: "json"})
	require.Error(t, err)
}

func TestFromContext_FallsBackToDefault(t *testing.T) {
	got := FromContext(context.Background())
	assert.NotNil(t, got)
}

func TestWithLogger_RoundTrip(t *testing.T) {
	var buf bytes.Buffer
	logger, err := New(Options{Level: "debug", Format: "json", Output: &buf})
	require.NoError(t, err)
	ctx := WithLogger(context.Background(), logger)
	assert.Same(t, logger, FromContext(ctx))
}

func TestWithLogger_NilNoOp(t *testing.T) {
	ctx := WithLogger(context.Background(), nil)
	assert.NotNil(t, FromContext(ctx))
}
