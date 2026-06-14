// Package diagnostics provides a lightweight health-enrichment layer.
//
// Each diagnostics.Provider returns a Check describing the state of one
// subsystem (uptime, scheduler, DB, ...). The Registry aggregates
// providers and computes a single Overall status used by the
// /api/v1/health endpoint added in Sprint 01.
//
// Providers are cheap to add; later sprints can plug in cache,
// transcoder, indexer and notification checks behind the same contract.
package diagnostics

import (
	"context"
	"sync"
	"time"
)

// Status is the canonical health verdict for a single check or for the
// aggregate Registry.
type Status string

const (
	StatusOK       Status = "ok"
	StatusDegraded Status = "degraded"
	StatusDown     Status = "down"
)

// Check is the data returned by a provider.
type Check struct {
	Name      string `json:"name"`
	Status    Status `json:"status"`
	Detail    string `json:"detail,omitempty"`
	LatencyMs int64  `json:"latencyMs"`
	Critical  bool   `json:"critical,omitempty"`
}

// Provider produces a Check on demand. Implementations should be safe
// for concurrent use because Registry may invoke them in parallel.
type Provider interface {
	Name() string
	Check(ctx context.Context) Check
	Critical() bool
}

// Pinger is the minimal contract needed by DBProvider. Paket 4's *DB
// satisfies it through its existing Ping(ctx) method.
type Pinger interface {
	Ping(ctx context.Context) error
}

// Registry aggregates Providers into a single overall verdict.
type Registry struct {
	mu        sync.RWMutex
	providers []Provider
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry { return &Registry{} }

// Register adds a provider. Calling Register with the same provider
// pointer twice is a no-op.
func (r *Registry) Register(p Provider) {
	if p == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, existing := range r.providers {
		if existing == p {
			return
		}
	}
	r.providers = append(r.providers, p)
}

// Snapshot is the aggregated diagnostics output.
type Snapshot struct {
	Overall Status  `json:"overall"`
	Checks  []Check `json:"checks"`
}

// Run executes every registered provider and computes the overall
// verdict:
//
//   - Down       if any critical provider reports Down.
//   - Degraded   if any non-critical provider is not OK,
//     or any critical provider is Degraded.
//   - OK         otherwise.
//
// Providers are invoked sequentially today; if a future provider needs
// parallelism, this is the place to add it.
func (r *Registry) Run(ctx context.Context) Snapshot {
	r.mu.RLock()
	providers := make([]Provider, len(r.providers))
	copy(providers, r.providers)
	r.mu.RUnlock()

	checks := make([]Check, 0, len(providers))
	overall := StatusOK
	for _, p := range providers {
		start := time.Now()
		c := p.Check(ctx)
		if c.LatencyMs == 0 {
			c.LatencyMs = time.Since(start).Milliseconds()
		}
		if c.Name == "" {
			c.Name = p.Name()
		}
		c.Critical = p.Critical()
		checks = append(checks, c)

		switch c.Status {
		case StatusDown:
			if c.Critical {
				overall = StatusDown
			} else if overall != StatusDown {
				overall = StatusDegraded
			}
		case StatusDegraded:
			if overall == StatusOK {
				overall = StatusDegraded
			}
		}
	}
	return Snapshot{Overall: overall, Checks: checks}
}
