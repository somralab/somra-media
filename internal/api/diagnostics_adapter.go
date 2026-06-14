package api

import (
	"context"
	"time"

	"github.com/somralab/somra-media/internal/platform/diagnostics"
)

// DiagnosticsAggregator is the production HealthAggregator: it bridges
// the platform diagnostics.Registry into the api package's per-check
// health view. Critical providers that return Down propagate as the
// overall response status; non-critical Down/Degraded surface as
// "degraded". The mapping keeps /api/v1/health behaviour aligned with
// diagnostics.Registry.Run while keeping the wire shape unchanged.
type DiagnosticsAggregator struct {
	Registry *diagnostics.Registry
}

// NewDiagnosticsAggregator returns an Aggregator backed by r.
func NewDiagnosticsAggregator(r *diagnostics.Registry) *DiagnosticsAggregator {
	return &DiagnosticsAggregator{Registry: r}
}

// Aggregate satisfies HealthAggregator.
func (a *DiagnosticsAggregator) Aggregate(ctx context.Context) (string, map[string]HealthStatus) {
	if a == nil || a.Registry == nil {
		return "ok", nil
	}
	snap := a.Registry.Run(ctx)
	out := make(map[string]HealthStatus, len(snap.Checks))
	for _, c := range snap.Checks {
		out[c.Name] = HealthStatus{
			Status:  string(c.Status),
			Detail:  c.Detail,
			Latency: time.Duration(c.LatencyMs) * time.Millisecond,
		}
	}
	return string(snap.Overall), out
}
