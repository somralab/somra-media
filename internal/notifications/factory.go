package notifications

import (
	"context"
	"fmt"
	"net/http"
)

// BuildChannels constructs enabled channels from settings-backed config.
func BuildChannels(ctx context.Context, reader SettingsReader, httpClient *http.Client) ([]Channel, error) {
	cfg, err := LoadChannelConfig(ctx, reader)
	if err != nil {
		return nil, err
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	out := make([]Channel, 0, 3)
	if cfg.WebhookEnabled && cfg.WebhookURL != "" {
		ch, err := NewWebhookChannel(WebhookConfig{
			URL:        cfg.WebhookURL,
			HTTPClient: httpClient,
		})
		if err != nil {
			return nil, fmt.Errorf("notifications: webhook channel: %w", err)
		}
		out = append(out, ch)
	}
	if cfg.DiscordEnabled && cfg.DiscordWebhookURL != "" {
		ch, err := NewDiscordChannel(DiscordConfig{
			WebhookURL: cfg.DiscordWebhookURL,
			HTTPClient: httpClient,
		})
		if err != nil {
			return nil, fmt.Errorf("notifications: discord channel: %w", err)
		}
		out = append(out, ch)
	}
	if cfg.SMTPEnabled && cfg.SMTPHost != "" && cfg.SMTPFrom != "" {
		ch, err := NewSMTPChannel(SMTPConfig{
			Host:     cfg.SMTPHost,
			Port:     cfg.SMTPPort,
			Username: cfg.SMTPUsername,
			Password: cfg.SMTPPassword,
			From:     cfg.SMTPFrom,
			UseTLS:   cfg.SMTPUseTLS,
		})
		if err != nil {
			return nil, fmt.Errorf("notifications: smtp channel: %w", err)
		}
		out = append(out, ch)
	}
	return out, nil
}
