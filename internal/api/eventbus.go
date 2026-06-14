package api

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/somralab/somra-media/internal/library"
)

// EventBus broadcasts SSE-compatible events to subscribers.
type EventBus struct {
	mu          sync.RWMutex
	subscribers map[chan []byte]struct{}
}

// NewEventBus returns an empty event bus.
func NewEventBus() *EventBus {
	return &EventBus{subscribers: make(map[chan []byte]struct{})}
}

// Subscribe registers a channel for broadcast events.
func (b *EventBus) Subscribe() chan []byte {
	ch := make(chan []byte, 16)
	b.mu.Lock()
	b.subscribers[ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

// Unsubscribe removes a subscriber channel.
func (b *EventBus) Unsubscribe(ch chan []byte) {
	b.mu.Lock()
	delete(b.subscribers, ch)
	b.mu.Unlock()
	close(ch)
}

// Publish sends a named SSE payload to all subscribers.
func (b *EventBus) Publish(eventName string, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	frame, err := json.Marshal(map[string]json.RawMessage{
		"event": json.RawMessage(mustJSON(eventName)),
		"data":  json.RawMessage(data),
	})
	if err != nil {
		return
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.subscribers {
		select {
		case ch <- frame:
		default:
		}
	}
}

func mustJSON(s string) []byte {
	b, _ := json.Marshal(s)
	return b
}

// ScanProgressPublisher adapts EventBus to library.ProgressPublisher.
type ScanProgressPublisher struct {
	Bus *EventBus
}

// PublishScanProgress emits scan progress over SSE.
func (p ScanProgressPublisher) PublishScanProgress(_ context.Context, ev library.ProgressEvent) {
	if p.Bus == nil {
		return
	}
	p.Bus.Publish("scan.progress", ev)
}

// RequestStatusEvent is the SSE payload for request lifecycle updates.
type RequestStatusEvent struct {
	RequestID int64  `json:"requestId"`
	Status    string `json:"status"`
	UpdatedAt string `json:"updatedAt"`
}

// RequestStatusPublisher emits request.status over SSE.
type RequestStatusPublisher struct {
	Bus *EventBus
}

// PublishRequestStatus broadcasts a request status change.
func (p RequestStatusPublisher) PublishRequestStatus(ev RequestStatusEvent) {
	if p.Bus == nil {
		return
	}
	p.Bus.Publish("request.status", ev)
}
