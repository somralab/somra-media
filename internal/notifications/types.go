package notifications

import "time"

// EventType identifies a notification-triggering domain event.
type EventType string

const (
	EventRequestCreated   EventType = "request.created"
	EventRequestApproved  EventType = "request.approved"
	EventRequestRejected  EventType = "request.rejected"
	EventRequestCompleted EventType = "request.completed"
	EventSystemError      EventType = "system.error"
)

// ChannelID names a delivery channel implementation.
type ChannelID string

const (
	ChannelWebhook ChannelID = "webhook"
	ChannelDiscord ChannelID = "discord"
	ChannelSMTP    ChannelID = "smtp"
)

// Notification is the rendered payload passed to a Channel.
type Notification struct {
	EventType EventType
	Subject   string
	Body      string
	Locale    string
	UserID    string
	Data      map[string]any
	SentAt    time.Time
}
