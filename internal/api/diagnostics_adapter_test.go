package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/platform/diagnostics"
)

func TestDiagnosticsAggregator_Aggregate(t *testing.T) {
	reg := diagnostics.NewRegistry()
	reg.Register(diagnostics.NewUptimeProvider())
	agg := NewDiagnosticsAggregator(reg)
	status, checks := agg.Aggregate(context.Background())
	assert.NotEmpty(t, status)
	require.NotEmpty(t, checks)
}

func TestDiagnosticsAggregator_NilRegistry(t *testing.T) {
	var agg *DiagnosticsAggregator
	status, checks := agg.Aggregate(context.Background())
	assert.Equal(t, "ok", status)
	assert.Nil(t, checks)
}
