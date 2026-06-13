package api

import (
	"encoding/json"
	"net/http"
	"time"
)

// HealthCheck is the contract diagnostic providers register against. Sprint 05
// (Paket 5 / Sprint 02 DB) extends the health endpoint by registering DB,
// scheduler and storage checks; this skeleton merely returns "ok" for the
// process itself.
type HealthCheck interface {
	Name() string
	Check() HealthStatus
}

// HealthStatus is the per-check outcome surfaced by /api/v1/health.
type HealthStatus struct {
	Status  string         `json:"status"`
	Detail  string         `json:"detail,omitempty"`
	Latency time.Duration  `json:"-"`
	Extra   map[string]any `json:"extra,omitempty"`
}

// HealthResponse is the JSON body returned by /api/v1/health.
type HealthResponse struct {
	Status string                  `json:"status"`
	Time   string                  `json:"time"`
	Checks map[string]HealthStatus `json:"checks,omitempty"`
}

// healthHandler builds /api/v1/health. checks is optional; nil yields the
// minimal {status: ok, time} payload defined in the Sprint 01 acceptance
// criteria.
func healthHandler(now func() time.Time, checks []HealthCheck) http.HandlerFunc {
	if now == nil {
		now = time.Now
	}
	return func(w http.ResponseWriter, r *http.Request) {
		resp := HealthResponse{
			Status: "ok",
			Time:   now().UTC().Format(time.RFC3339Nano),
		}
		if len(checks) > 0 {
			resp.Checks = make(map[string]HealthStatus, len(checks))
			for _, c := range checks {
				if c == nil {
					continue
				}
				s := c.Check()
				resp.Checks[c.Name()] = s
				if s.Status != "ok" {
					resp.Status = "degraded"
				}
			}
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(resp)
	}
}
