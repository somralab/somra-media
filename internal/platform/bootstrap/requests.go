package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/somralab/somra-media/internal/api"
	"github.com/somralab/somra-media/internal/metadata"
	"github.com/somralab/somra-media/internal/notifications"
	"github.com/somralab/somra-media/internal/platform/db"
	"github.com/somralab/somra-media/internal/requests"
)

// RequestsBundle groups Sprint 08 request and notification HTTP dependencies.
type RequestsBundle struct {
	Requests      *api.RequestHandlers
	Notifications *api.NotificationHandlers
	Dispatcher    *notifications.Dispatcher
}

// WireRequests constructs request workflow services and notification dispatch.
func WireRequests(c *Components, lib *LibraryBundle, localeFn func(*http.Request) string) (*RequestsBundle, error) {
	if c == nil || c.DB == nil {
		return nil, fmt.Errorf("bootstrap requests: db required")
	}
	q := c.DB.Querier()
	reqRepo := db.NewRequestRepo(q)
	users := db.NewUserRepo(q)
	chRepo := db.NewNotificationChannelRepo(q)
	prefRepo := db.NewNotificationPreferenceRepo(q)
	profiles := db.NewProfileRepo(q)

	prefStore := &notifications.DBPreferenceStore{
		Prefs:    prefRepo,
		Channels: chRepo,
		Profiles: profiles,
		Users:    users,
	}
	dispatcher := notifications.NewDispatcher(notifications.DispatcherConfig{
		Renderer: notifications.NewTemplateRenderer(c.I18n),
		Filter:   notifications.NewPreferenceFilter(prefStore),
		Debounce: notifications.NewDebouncer(60 * time.Second),
		Queue:    c.Queue,
		Logger:   c.Logger,
	})
	if err := registerDBChannels(context.Background(), chRepo, dispatcher); err != nil && c.Logger != nil {
		c.Logger.Warn("bootstrap requests: register notification channels", slog.Any("error", err))
	}

	collision := &requests.CollisionChecker{
		Library:  &requests.DBLibraryLookup{Q: q},
		Requests: &requests.DBPendingRequestLookup{Q: q},
	}
	policy := &requests.PolicyService{
		Policies: requests.DBPolicyStore{Repo: reqRepo},
		Counter:  requests.DBRequestCounter{Repo: reqRepo},
	}

	var discoverer *requests.Discoverer
	if lib != nil && lib.Metadata != nil && lib.Metadata.Registry != nil {
		discoverer = &requests.Discoverer{
			Registry: lib.Metadata.Registry,
			Library:  &requests.DBLibraryLookup{Q: q},
			Limiter:  metadata.NewRateLimiter(250 * time.Millisecond),
		}
	}

	statusPub := api.RequestStatusPublisher{}
	if lib != nil {
		statusPub.Bus = lib.EventBus
	}

	return &RequestsBundle{
		Dispatcher: dispatcher,
		Requests: &api.RequestHandlers{
			Repo:       reqRepo,
			Users:      users,
			Collision:  collision,
			Policy:     policy,
			Discoverer: discoverer,
			Handoff:    requests.NoOpAutomationHandoff{},
			Notify:     dispatcher,
			StatusPub:  statusPub,
			Locale:     localeFn,
		},
		Notifications: &api.NotificationHandlers{
			Channels:   chRepo,
			Prefs:      prefRepo,
			Dispatcher: dispatcher,
		},
	}, nil
}

func registerDBChannels(ctx context.Context, repo *db.NotificationChannelRepo, d *notifications.Dispatcher) error {
	if repo == nil || d == nil {
		return nil
	}
	rows, err := repo.List(ctx)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if !row.Enabled {
			continue
		}
		ch, err := notifications.ChannelFromDB(row)
		if err != nil {
			return fmt.Errorf("channel %d: %w", row.ID, err)
		}
		d.RegisterChannel(ch)
	}
	return nil
}
