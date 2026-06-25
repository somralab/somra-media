package bootstrap

import (
	"context"
	"fmt"
	"runtime"

	"github.com/somralab/somra-media/internal/api"
	"github.com/somralab/somra-media/internal/automation/download"
	"github.com/somralab/somra-media/internal/automation/handoff"
	"github.com/somralab/somra-media/internal/automation/importsvc"
	indexersearch "github.com/somralab/somra-media/internal/automation/indexer"
	"github.com/somralab/somra-media/internal/automation/seriesmonitor"
	"github.com/somralab/somra-media/internal/automation/worker"
	"github.com/somralab/somra-media/internal/jobs"
	"github.com/somralab/somra-media/internal/platform/db"
	"github.com/somralab/somra-media/internal/plugin"
	"github.com/somralab/somra-media/internal/requests"
)

// AutomationBundle groups Sprint 09 automation dependencies.
type AutomationBundle struct {
	Handlers *api.AutomationHandlers
	Queue    *handoff.Queue
}

// WireAutomation constructs automation services, schedules jobs, and handoff queue.
func WireAutomation(c *Components, lib *LibraryBundle, plugins *PluginsBundle, req *RequestsBundle) (*AutomationBundle, error) {
	if c == nil || c.DB == nil {
		return nil, fmt.Errorf("bootstrap automation: db required")
	}
	autoRepo := db.NewAutomationRepo(c.DB.Querier())
	reqRepo := db.NewRequestRepo(c.DB.Querier())
	queue := &handoff.Queue{Repo: autoRepo}

	var mgr *plugin.Manager
	if plugins != nil {
		mgr = plugins.Manager
	}
	searchSvc := &indexersearch.SearchService{Manager: mgr, Logger: c.Logger}
	processor := &worker.Processor{
		AutoRepo: autoRepo,
		Requests: reqRepo,
		Search:   searchSvc,
		Manager:  mgr,
		Logger:   c.Logger,
	}
	var importSvc *importsvc.Service
	if lib != nil && lib.Library != nil {
		importSvc = &importsvc.Service{Library: lib.Library}
	}
	monitor := &download.Monitor{
		AutoRepo: autoRepo,
		Requests: reqRepo,
		Manager:  mgr,
		Import:   importSvc,
		Logger:   c.Logger,
	}

	if c.Scheduler != nil {
		seriesScanner := &seriesmonitor.Scanner{
			AutoRepo: autoRepo,
			Requests: reqRepo,
			Search:   searchSvc,
			Logger:   c.Logger,
		}
		scheduleAutomationJobs(c, processor, monitor, seriesScanner)
	}

	if req != nil && req.Requests != nil {
		req.Requests.Handoff = requests.QueuingAutomationHandoff{Queue: queue}
	}

	return &AutomationBundle{
		Handlers: &api.AutomationHandlers{
			AutoRepo: autoRepo,
			Search:   searchSvc,
		},
		Queue: queue,
	}, nil
}

func scheduleAutomationJobs(c *Components, processor *worker.Processor, monitor *download.Monitor, seriesScanner *seriesmonitor.Scanner) {
	if c == nil || c.Scheduler == nil {
		return
	}
	pollCron := "0 */1 * * * *"
	if runtime.NumCPU() <= 4 {
		pollCron = "0 */2 * * * *"
	}
	_, _ = c.Scheduler.Schedule(pollCron, "automation-handoff", jobs.JobFunc(func(ctx context.Context) error {
		return processor.ProcessPending(ctx)
	}))
	_, _ = c.Scheduler.Schedule(pollCron, "automation-download-monitor", jobs.JobFunc(func(ctx context.Context) error {
		return monitor.Poll(ctx)
	}))
	if seriesScanner != nil {
		_, _ = c.Scheduler.Schedule("0 0 */15 * * *", "automation-series-monitor", jobs.JobFunc(func(ctx context.Context) error {
			return seriesScanner.ScanEnabledMonitors(ctx)
		}))
	}
	if c.Logger != nil {
		c.Logger.Info("automation jobs scheduled")
	}
}
