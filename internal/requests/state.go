package requests

import (
	"fmt"
	"time"
)

var allowedTransitions = map[Status]map[Status]struct{}{
	StatusPending: {
		StatusApproved:  {},
		StatusRejected:  {},
		StatusCancelled: {},
	},
	StatusApproved: {
		StatusCompleted: {},
		StatusCancelled: {},
	},
}

// CanTransition reports whether current may move to next.
func CanTransition(current, next Status) bool {
	targets, ok := allowedTransitions[current]
	if !ok {
		return false
	}
	_, ok = targets[next]
	return ok
}

// AllowedTargets returns valid next statuses from current.
func AllowedTargets(current Status) []Status {
	targets := allowedTransitions[current]
	out := make([]Status, 0, len(targets))
	for s := range targets {
		out = append(out, s)
	}
	return out
}

// IsTerminal reports whether no further transitions are permitted.
func IsTerminal(status Status) bool {
	return len(allowedTransitions[status]) == 0
}

// TransitionTo applies next to req when permitted.
func TransitionTo(req *Request, next Status, now time.Time) error {
	if req == nil {
		return fmt.Errorf("requests transition: %w", ErrInvalidTransition)
	}
	if !CanTransition(req.Status, next) {
		return fmt.Errorf("requests transition %q -> %q: %w", req.Status, next, ErrInvalidTransition)
	}
	req.Status = next
	req.UpdatedAt = now
	return nil
}
