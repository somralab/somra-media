package notifications

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/somralab/somra-media/internal/jobs"
)

// Dispatcher routes domain events to pluggable channels with preference
// filtering, i18n rendering, debouncing, and async delivery.
type Dispatcher struct {
	mu       sync.RWMutex
	channels map[ChannelID]Channel
	renderer *TemplateRenderer
	filter   *PreferenceFilter
	debounce *Debouncer
	queue    jobs.JobQueue
	logger   *slog.Logger
	now      func() time.Time
}

// DispatcherConfig configures a Dispatcher.
type DispatcherConfig struct {
	Renderer *TemplateRenderer
	Filter   *PreferenceFilter
	Debounce *Debouncer
	Queue    jobs.JobQueue
	Logger   *slog.Logger
}

// NewDispatcher returns a dispatcher with optional dependencies.
func NewDispatcher(cfg DispatcherConfig) *Dispatcher {
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}
	return &Dispatcher{
		channels: make(map[ChannelID]Channel),
		renderer: cfg.Renderer,
		filter:   cfg.Filter,
		debounce: cfg.Debounce,
		queue:    cfg.Queue,
		logger:   logger,
		now:      time.Now,
	}
}

// RegisterChannel adds or replaces a delivery channel.
func (d *Dispatcher) RegisterChannel(ch Channel) {
	if d == nil || ch == nil {
		return
	}
	d.mu.Lock()
	d.channels[ch.ID()] = ch
	d.mu.Unlock()
}

// Channel returns a registered channel by ID.
func (d *Dispatcher) Channel(id ChannelID) (Channel, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	ch, ok := d.channels[id]
	return ch, ok
}

// Dispatch enqueues notification delivery for each eligible recipient
// and channel. Rendering and preference checks run synchronously;
// channel Send calls run asynchronously via the job queue.
func (d *Dispatcher) Dispatch(ctx context.Context, ev Event) error {
	if d == nil {
		return fmt.Errorf("notifications: dispatcher is nil")
	}
	if d.renderer == nil {
		return fmt.Errorf("notifications: template renderer not configured")
	}
	if d.queue == nil {
		return fmt.Errorf("notifications: job queue not configured")
	}

	recipients := ev.Recipients()
	if len(recipients) == 0 {
		return nil
	}

	for _, userID := range recipients {
		if err := d.dispatchToUser(ctx, ev, userID); err != nil {
			return err
		}
	}
	return nil
}

func (d *Dispatcher) dispatchToUser(ctx context.Context, ev Event, userID string) error {
	filter := d.filter
	if filter == nil {
		filter = NewPreferenceFilter(nil)
	}
	ok, err := filter.ShouldNotify(ctx, userID, ev.Type)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	locale, err := filter.Locale(ctx, userID)
	if err != nil {
		return err
	}
	rendered, err := d.renderer.Render(ev.Type, locale, ev.TemplateData())
	if err != nil {
		return err
	}

	channelIDs, err := filter.AllowedChannels(ctx, userID, ev.Type)
	if err != nil {
		return err
	}

	n := Notification{
		EventType: ev.Type,
		Subject:   rendered.Subject,
		Body:      rendered.Body,
		Locale:    rendered.Locale,
		UserID:    userID,
		Data:      ev.TemplateData(),
		SentAt:    d.now(),
	}

	for _, chID := range channelIDs {
		ch, registered := d.Channel(chID)
		if !registered {
			continue
		}
		key := DebounceKey(userID, ev.Type, chID)
		window := d.debounceWindow(ctx, filter, userID)
		if d.debounce != nil && !d.debounce.AllowWithWindow(key, window) {
			d.logger.DebugContext(ctx, "notifications.debounced",
				slog.String("user", userID),
				slog.String("event", string(ev.Type)),
				slog.String("channel", string(chID)),
			)
			continue
		}
		if err := d.enqueueSend(ctx, ch, n); err != nil {
			return err
		}
	}
	return nil
}

func (d *Dispatcher) enqueueSend(ctx context.Context, ch Channel, n Notification) error {
	chCopy := ch
	nCopy := n
	_, err := d.queue.Enqueue(ctx, jobs.JobFunc(func(jobCtx context.Context) error {
		if err := chCopy.Send(jobCtx, nCopy); err != nil {
			d.logger.ErrorContext(jobCtx, "notifications.send.failed",
				slog.String("channel", string(chCopy.ID())),
				slog.String("event", string(nCopy.EventType)),
				slog.String("user", nCopy.UserID),
				slog.String("error", err.Error()),
			)
			return fmt.Errorf("notifications: send via %s: %w", chCopy.ID(), err)
		}
		d.logger.InfoContext(jobCtx, "notifications.send.ok",
			slog.String("channel", string(chCopy.ID())),
			slog.String("event", string(nCopy.EventType)),
			slog.String("user", nCopy.UserID),
		)
		return nil
	}), jobs.WithName("notification:"+string(ch.ID())))
	return err
}

func (d *Dispatcher) debounceWindow(ctx context.Context, filter *PreferenceFilter, userID string) time.Duration {
	if d.debounce == nil {
		return 0
	}
	if override, err := filter.DebounceSeconds(ctx, userID); err == nil && override != nil && *override > 0 {
		return time.Duration(*override) * time.Second
	}
	return d.debounce.windowForDefault()
}

// HandleRequestCreated dispatches a request.created notification.
func (d *Dispatcher) HandleRequestCreated(ctx context.Context, ev Event) error {
	ev.Type = EventRequestCreated
	return d.Dispatch(ctx, ev)
}

// HandleRequestApproved dispatches a request.approved notification.
func (d *Dispatcher) HandleRequestApproved(ctx context.Context, ev Event) error {
	ev.Type = EventRequestApproved
	return d.Dispatch(ctx, ev)
}

// HandleRequestRejected dispatches a request.rejected notification.
func (d *Dispatcher) HandleRequestRejected(ctx context.Context, ev Event) error {
	ev.Type = EventRequestRejected
	return d.Dispatch(ctx, ev)
}

// HandleRequestCompleted dispatches a request.completed notification.
func (d *Dispatcher) HandleRequestCompleted(ctx context.Context, ev Event) error {
	ev.Type = EventRequestCompleted
	return d.Dispatch(ctx, ev)
}

// HandleSystemError dispatches a system.error notification to admins.
func (d *Dispatcher) HandleSystemError(ctx context.Context, ev Event) error {
	ev.Type = EventSystemError
	return d.Dispatch(ctx, ev)
}
