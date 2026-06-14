package requests

import "context"

// AutomationHandoff transfers an approved request to the automation module (Sprint 09).
type AutomationHandoff interface {
	Handoff(ctx context.Context, req Request) error
}

// HandoffQueue records handoff intent for later processing by automation.
type HandoffQueue interface {
	RecordHandoff(ctx context.Context, req Request) error
}

// NoOpAutomationHandoff satisfies [AutomationHandoff] without side effects.
type NoOpAutomationHandoff struct{}

// Handoff implements [AutomationHandoff].
func (NoOpAutomationHandoff) Handoff(_ context.Context, _ Request) error {
	return nil
}

// QueuingAutomationHandoff records handoff when a queue is configured.
type QueuingAutomationHandoff struct {
	Queue HandoffQueue
}

// Handoff implements [AutomationHandoff].
func (h QueuingAutomationHandoff) Handoff(ctx context.Context, req Request) error {
	if h.Queue == nil {
		return nil
	}
	return h.Queue.RecordHandoff(ctx, req)
}
