package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// WebhookChannel posts JSON payloads to a configured URL.
type WebhookChannel struct {
	url        string
	httpClient *http.Client
}

// WebhookConfig configures a WebhookChannel.
type WebhookConfig struct {
	URL        string
	HTTPClient *http.Client
}

// NewWebhookChannel returns a webhook channel. URL must be non-empty.
func NewWebhookChannel(cfg WebhookConfig) (*WebhookChannel, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("notifications/webhook: url is required")
	}
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	return &WebhookChannel{url: cfg.URL, httpClient: client}, nil
}

// ID implements Channel.
func (c *WebhookChannel) ID() ChannelID { return ChannelWebhook }

type webhookPayload struct {
	EventType string         `json:"eventType"`
	Subject   string         `json:"subject"`
	Body      string         `json:"body"`
	Locale    string         `json:"locale"`
	UserID    string         `json:"userId"`
	SentAt    time.Time      `json:"sentAt"`
	Data      map[string]any `json:"data,omitempty"`
}

// Send implements Channel.
func (c *WebhookChannel) Send(ctx context.Context, n Notification) error {
	if c == nil {
		return fmt.Errorf("notifications/webhook: channel is nil")
	}
	payload := webhookPayload{
		EventType: string(n.EventType),
		Subject:   n.Subject,
		Body:      n.Body,
		Locale:    n.Locale,
		UserID:    n.UserID,
		SentAt:    n.SentAt,
		Data:      n.Data,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("notifications/webhook: marshal payload: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("notifications/webhook: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Somra-Notifications/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("notifications/webhook: post: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("notifications/webhook: unexpected status %d: %s", resp.StatusCode, string(snippet))
	}
	return nil
}
