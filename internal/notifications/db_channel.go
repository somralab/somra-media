package notifications

import (
	"encoding/json"
	"fmt"

	"github.com/somralab/somra-media/internal/platform/db"
)

// ChannelFromDB constructs a delivery channel from a persisted row.
func ChannelFromDB(ch db.NotificationChannel) (Channel, error) {
	switch ch.ChannelType {
	case db.NotificationChannelWebhook:
		var cfg struct {
			URL string `json:"url"`
		}
		if err := json.Unmarshal([]byte(ch.Config), &cfg); err != nil {
			return nil, fmt.Errorf("notifications: decode webhook config: %w", err)
		}
		return NewWebhookChannel(WebhookConfig{URL: cfg.URL})
	case db.NotificationChannelDiscord:
		var cfg struct {
			WebhookURL string `json:"webhookUrl"`
			URL        string `json:"url"`
		}
		if err := json.Unmarshal([]byte(ch.Config), &cfg); err != nil {
			return nil, fmt.Errorf("notifications: decode discord config: %w", err)
		}
		url := cfg.WebhookURL
		if url == "" {
			url = cfg.URL
		}
		return NewDiscordChannel(DiscordConfig{WebhookURL: url})
	case db.NotificationChannelEmail:
		var cfg struct {
			Host     string `json:"host"`
			Port     int    `json:"port"`
			Username string `json:"username"`
			Password string `json:"password"`
			From     string `json:"from"`
			UseTLS   bool   `json:"useTls"`
		}
		if err := json.Unmarshal([]byte(ch.Config), &cfg); err != nil {
			return nil, fmt.Errorf("notifications: decode email config: %w", err)
		}
		if cfg.Port == 0 {
			cfg.Port = 587
		}
		return NewSMTPChannel(SMTPConfig{
			Host:     cfg.Host,
			Port:     cfg.Port,
			Username: cfg.Username,
			Password: cfg.Password,
			From:     cfg.From,
			UseTLS:   cfg.UseTLS,
		})
	default:
		return nil, fmt.Errorf("notifications: unsupported channel type %q", ch.ChannelType)
	}
}
