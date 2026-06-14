package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// Settings keys for notification channel configuration.
const (
	KeyWebhookEnabled     = "notifications.webhook.enabled"
	KeyWebhookURL         = "notifications.webhook.url"
	KeyDiscordEnabled     = "notifications.discord.enabled"
	KeyDiscordWebhookURL  = "notifications.discord.webhook_url"
	KeySMTPEnabled        = "notifications.smtp.enabled"
	KeySMTPHost           = "notifications.smtp.host"
	KeySMTPPort           = "notifications.smtp.port"
	KeySMTPUsername       = "notifications.smtp.username"
	KeySMTPPassword       = "notifications.smtp.password"
	KeySMTPFrom           = "notifications.smtp.from"
	KeySMTPUseTLS         = "notifications.smtp.use_tls"
	KeyDebounceSeconds    = "notifications.debounce_seconds"
	KeyDefaultAdminEmails = "notifications.admin_emails"
)

// SettingsReader loads persisted notification settings.
type SettingsReader interface {
	GetString(ctx context.Context, key, defaultVal string) (string, error)
}

// ChannelConfig holds parsed channel settings.
type ChannelConfig struct {
	WebhookEnabled    bool
	WebhookURL        string
	DiscordEnabled    bool
	DiscordWebhookURL string
	SMTPEnabled       bool
	SMTPHost          string
	SMTPPort          int
	SMTPUsername      string
	SMTPPassword      string
	SMTPFrom          string
	SMTPUseTLS        bool
	DebounceSeconds   int
	AdminEmails       []string
}

// LoadChannelConfig reads notification settings via reader.
func LoadChannelConfig(ctx context.Context, reader SettingsReader) (ChannelConfig, error) {
	if reader == nil {
		return ChannelConfig{}, fmt.Errorf("notifications: settings reader is required")
	}
	cfg := ChannelConfig{DebounceSeconds: 60, SMTPPort: 587, SMTPUseTLS: true}

	var err error
	cfg.WebhookEnabled, err = readBool(ctx, reader, KeyWebhookEnabled, false)
	if err != nil {
		return ChannelConfig{}, err
	}
	cfg.WebhookURL, err = reader.GetString(ctx, KeyWebhookURL, "")
	if err != nil {
		return ChannelConfig{}, err
	}
	cfg.DiscordEnabled, err = readBool(ctx, reader, KeyDiscordEnabled, false)
	if err != nil {
		return ChannelConfig{}, err
	}
	cfg.DiscordWebhookURL, err = reader.GetString(ctx, KeyDiscordWebhookURL, "")
	if err != nil {
		return ChannelConfig{}, err
	}
	cfg.SMTPEnabled, err = readBool(ctx, reader, KeySMTPEnabled, false)
	if err != nil {
		return ChannelConfig{}, err
	}
	cfg.SMTPHost, err = reader.GetString(ctx, KeySMTPHost, "")
	if err != nil {
		return ChannelConfig{}, err
	}
	portRaw, err := reader.GetString(ctx, KeySMTPPort, "587")
	if err != nil {
		return ChannelConfig{}, err
	}
	if portRaw != "" {
		n, convErr := strconv.Atoi(portRaw)
		if convErr != nil || n < 1 || n > 65535 {
			return ChannelConfig{}, fmt.Errorf("notifications: invalid smtp port %q", portRaw)
		}
		cfg.SMTPPort = n
	}
	cfg.SMTPUsername, err = reader.GetString(ctx, KeySMTPUsername, "")
	if err != nil {
		return ChannelConfig{}, err
	}
	cfg.SMTPPassword, err = reader.GetString(ctx, KeySMTPPassword, "")
	if err != nil {
		return ChannelConfig{}, err
	}
	cfg.SMTPFrom, err = reader.GetString(ctx, KeySMTPFrom, "")
	if err != nil {
		return ChannelConfig{}, err
	}
	cfg.SMTPUseTLS, err = readBool(ctx, reader, KeySMTPUseTLS, true)
	if err != nil {
		return ChannelConfig{}, err
	}
	debounceRaw, err := reader.GetString(ctx, KeyDebounceSeconds, "60")
	if err != nil {
		return ChannelConfig{}, err
	}
	if debounceRaw != "" {
		n, convErr := strconv.Atoi(debounceRaw)
		if convErr != nil || n < 0 {
			return ChannelConfig{}, fmt.Errorf("notifications: invalid debounce seconds %q", debounceRaw)
		}
		cfg.DebounceSeconds = n
	}
	adminRaw, err := reader.GetString(ctx, KeyDefaultAdminEmails, "[]")
	if err != nil {
		return ChannelConfig{}, err
	}
	cfg.AdminEmails = parseEmailList(adminRaw)
	return cfg, nil
}

// DebounceWindow returns the configured debounce duration.
func (c ChannelConfig) DebounceWindow() time.Duration {
	if c.DebounceSeconds <= 0 {
		return 0
	}
	return time.Duration(c.DebounceSeconds) * time.Second
}

func readBool(ctx context.Context, reader SettingsReader, key string, def bool) (bool, error) {
	defVal := "false"
	if def {
		defVal = "true"
	}
	raw, err := reader.GetString(ctx, key, defVal)
	if err != nil {
		return false, err
	}
	return raw == "true" || raw == "1", nil
}

func parseEmailList(raw string) []string {
	if raw == "" {
		return nil
	}
	var emails []string
	if err := json.Unmarshal([]byte(raw), &emails); err == nil {
		return emails
	}
	return nil
}
