package diagnostics

import (
	"context"
	"fmt"
	"time"

	"github.com/somralab/somra-media/internal/jobs"
)

// UptimeProvider records the process start time and reports it as a
// non-critical informational check.
type UptimeProvider struct {
	name      string
	startedAt time.Time
}

// NewUptimeProvider builds a provider whose start time is captured at
// construction.
func NewUptimeProvider() *UptimeProvider {
	return &UptimeProvider{name: "uptime", startedAt: time.Now()}
}

// Name implements Provider.
func (p *UptimeProvider) Name() string { return p.name }

// Critical implements Provider; uptime is informational.
func (p *UptimeProvider) Critical() bool { return false }

// Check implements Provider.
func (p *UptimeProvider) Check(_ context.Context) Check {
	dur := time.Since(p.startedAt)
	return Check{
		Name:   p.name,
		Status: StatusOK,
		Detail: fmt.Sprintf("up %s", dur.Round(time.Second)),
	}
}

// StartedAt returns the captured start time. Useful for tests and the
// HTTP /version handler.
func (p *UptimeProvider) StartedAt() time.Time { return p.startedAt }

// SchedulerProvider reflects whether the supplied jobs.Scheduler is
// running. It is non-critical: a stopped scheduler degrades the system
// but does not break user-facing requests.
type SchedulerProvider struct {
	scheduler *jobs.Scheduler
}

// NewSchedulerProvider wires the diagnostics view of the scheduler.
func NewSchedulerProvider(s *jobs.Scheduler) *SchedulerProvider {
	return &SchedulerProvider{scheduler: s}
}

// Name implements Provider.
func (p *SchedulerProvider) Name() string { return "scheduler" }

// Critical implements Provider.
func (p *SchedulerProvider) Critical() bool { return false }

// Check implements Provider.
func (p *SchedulerProvider) Check(_ context.Context) Check {
	if p.scheduler == nil {
		return Check{Name: "scheduler", Status: StatusDown, Detail: "scheduler missing"}
	}
	if p.scheduler.Running() {
		return Check{Name: "scheduler", Status: StatusOK, Detail: "running"}
	}
	return Check{Name: "scheduler", Status: StatusDegraded, Detail: "idle"}
}

// DBProvider pings the underlying database. It is critical: a down
// database means the system cannot serve requests.
type DBProvider struct {
	name   string
	pinger Pinger
}

// NewDBProvider builds a provider that pings p. A nil pinger reports
// Down so misconfigured wiring is visible immediately.
func NewDBProvider(p Pinger) *DBProvider {
	return &DBProvider{name: "database", pinger: p}
}

// Name implements Provider.
func (d *DBProvider) Name() string { return d.name }

// Critical implements Provider.
func (d *DBProvider) Critical() bool { return true }

// Check implements Provider.
func (d *DBProvider) Check(ctx context.Context) Check {
	if d.pinger == nil {
		return Check{Name: d.name, Status: StatusDown, Detail: "pinger not configured"}
	}
	start := time.Now()
	if err := d.pinger.Ping(ctx); err != nil {
		return Check{
			Name:      d.name,
			Status:    StatusDown,
			Detail:    err.Error(),
			LatencyMs: time.Since(start).Milliseconds(),
		}
	}
	return Check{
		Name:      d.name,
		Status:    StatusOK,
		LatencyMs: time.Since(start).Milliseconds(),
	}
}
