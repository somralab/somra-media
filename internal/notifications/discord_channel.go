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

// DiscordChannel posts embed-style messages to a Discord webhook URL.
type DiscordChannel struct {
	webhookURL string
	httpClient *http.Client
}

// DiscordConfig configures a DiscordChannel.
type DiscordConfig struct {
	WebhookURL string
	HTTPClient *http.Client
}

// NewDiscordChannel returns a Discord webhook channel.
func NewDiscordChannel(cfg DiscordConfig) (*DiscordChannel, error) {
	if cfg.WebhookURL == "" {
		return nil, fmt.Errorf("notifications/discord: webhook url is required")
	}
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	return &DiscordChannel{webhookURL: cfg.WebhookURL, httpClient: client}, nil
}

// ID implements Channel.
func (c *DiscordChannel) ID() ChannelID { return ChannelDiscord }

type discordPayload struct {
	Content string         `json:"content,omitempty"`
	Embeds  []discordEmbed `json:"embeds,omitempty"`
}

type discordEmbed struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Color       int    `json:"color"`
}

// Send implements Channel.
func (c *DiscordChannel) Send(ctx context.Context, n Notification) error {
	if c == nil {
		return fmt.Errorf("notifications/discord: channel is nil")
	}
	payload := discordPayload{
		Embeds: []discordEmbed{{
			Title:       n.Subject,
			Description: n.Body,
			Color:       discordColorFor(n.EventType),
		}},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("notifications/discord: marshal payload: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("notifications/discord: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("notifications/discord: post: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("notifications/discord: unexpected status %d: %s", resp.StatusCode, string(snippet))
	}
	return nil
}

func discordColorFor(eventType EventType) int {
	switch eventType {
	case EventRequestApproved, EventRequestCompleted:
		return 0x57F287
	case EventRequestRejected, EventSystemError:
		return 0xED4245
	default:
		return 0x5865F2
	}
}
