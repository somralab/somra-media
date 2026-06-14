package notifications

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
)

// SMTPChannel sends email via SMTP using settings-backed configuration.
type SMTPChannel struct {
	host     string
	port     int
	username string
	password string
	from     string
	useTLS   bool
	dial     func(network, address string) (net.Conn, error)
}

// SMTPConfig configures an SMTPChannel.
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	UseTLS   bool
	Dial     func(network, address string) (net.Conn, error)
}

// NewSMTPChannel returns an SMTP channel. Host, port, and from are required.
func NewSMTPChannel(cfg SMTPConfig) (*SMTPChannel, error) {
	if strings.TrimSpace(cfg.Host) == "" {
		return nil, fmt.Errorf("notifications/smtp: host is required")
	}
	if cfg.Port < 1 || cfg.Port > 65535 {
		return nil, fmt.Errorf("notifications/smtp: invalid port %d", cfg.Port)
	}
	if strings.TrimSpace(cfg.From) == "" {
		return nil, fmt.Errorf("notifications/smtp: from address is required")
	}
	dial := cfg.Dial
	if dial == nil {
		dial = net.Dial
	}
	return &SMTPChannel{
		host:     cfg.Host,
		port:     cfg.Port,
		username: cfg.Username,
		password: cfg.Password,
		from:     cfg.From,
		useTLS:   cfg.UseTLS,
		dial:     dial,
	}, nil
}

// ID implements Channel.
func (c *SMTPChannel) ID() ChannelID { return ChannelSMTP }

// Send implements Channel. The recipient address is read from n.Data["email"]
// when present; otherwise Send is a no-op success so preference wiring can
// supply addresses later.
func (c *SMTPChannel) Send(ctx context.Context, n Notification) error {
	if c == nil {
		return fmt.Errorf("notifications/smtp: channel is nil")
	}
	to, _ := n.Data["email"].(string)
	if strings.TrimSpace(to) == "" {
		return nil
	}
	return c.sendMail(ctx, to, n.Subject, n.Body)
}

// SendTest delivers a test message to the given address.
func (c *SMTPChannel) SendTest(ctx context.Context, to string) error {
	if strings.TrimSpace(to) == "" {
		return fmt.Errorf("notifications/smtp: test recipient is required")
	}
	return c.sendMail(ctx, to, "Somra notification test", "This is a test notification from Somra.")
}

func (c *SMTPChannel) sendMail(ctx context.Context, to, subject, body string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	addr := fmt.Sprintf("%s:%d", c.host, c.port)
	msg := buildMIME(c.from, to, subject, body)

	if c.useTLS {
		return c.sendSTARTTLS(addr, to, msg)
	}
	auth := smtpPlainAuth(c.host, c.username, c.password)
	return smtp.SendMail(addr, auth, c.from, []string{to}, msg)
}

func (c *SMTPChannel) sendSTARTTLS(addr, to string, msg []byte) error {
	conn, err := c.dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("notifications/smtp: dial: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, c.host)
	if err != nil {
		return fmt.Errorf("notifications/smtp: client: %w", err)
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{ServerName: c.host, MinVersion: tls.VersionTLS12}); err != nil {
			return fmt.Errorf("notifications/smtp: starttls: %w", err)
		}
	}
	if auth := smtpPlainAuth(c.host, c.username, c.password); auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("notifications/smtp: auth: %w", err)
		}
	}
	if err := client.Mail(c.from); err != nil {
		return fmt.Errorf("notifications/smtp: mail from: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("notifications/smtp: rcpt: %w", err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("notifications/smtp: data: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("notifications/smtp: write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("notifications/smtp: close data: %w", err)
	}
	return client.Quit()
}

func smtpPlainAuth(host, username, password string) smtp.Auth {
	if username == "" && password == "" {
		return nil
	}
	return smtp.PlainAuth("", username, password, host)
}

func buildMIME(from, to, subject, body string) []byte {
	headers := []string{
		"From: " + from,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}
	return []byte(strings.Join(headers, "\r\n"))
}
