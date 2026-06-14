package notifications

import "context"

// Channel delivers a rendered notification to an external sink.
type Channel interface {
	ID() ChannelID
	Send(ctx context.Context, n Notification) error
}
