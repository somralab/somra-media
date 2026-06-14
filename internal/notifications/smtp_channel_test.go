package notifications_test

import (
	"bufio"
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/notifications"
)

func TestSMTPChannelSendTest(t *testing.T) {
	t.Parallel()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })

	done := make(chan struct{}, 1)
	go func() {
		conn, acceptErr := ln.Accept()
		if acceptErr != nil {
			return
		}
		defer conn.Close()
		_ = conn.SetDeadline(time.Now().Add(3 * time.Second))
		reader := bufio.NewReader(conn)
		write := func(msg string) { _, _ = conn.Write([]byte(msg)) }
		write("220 localhost ESMTP\r\n")
		for {
			line, readErr := reader.ReadString('\n')
			if readErr != nil {
				return
			}
			cmd := strings.ToUpper(strings.TrimSpace(line))
			switch {
			case strings.HasPrefix(cmd, "EHLO"), strings.HasPrefix(cmd, "HELO"):
				write("250-localhost\r\n250 OK\r\n")
			case strings.HasPrefix(cmd, "MAIL FROM"):
				write("250 OK\r\n")
			case strings.HasPrefix(cmd, "RCPT TO"):
				write("250 OK\r\n")
			case cmd == "DATA":
				write("354 go\r\n")
			case cmd == "QUIT":
				write("221 bye\r\n")
				done <- struct{}{}
				return
			default:
				if strings.TrimSpace(line) == "." {
					write("250 queued\r\n")
				}
			}
		}
	}()

	_, portStr, splitErr := net.SplitHostPort(ln.Addr().String())
	require.NoError(t, splitErr)

	ch, err := notifications.NewSMTPChannel(notifications.SMTPConfig{
		Host:   "127.0.0.1",
		Port:   mustAtoi(t, portStr),
		From:   "somra@test.local",
		UseTLS: false,
		Dial:   net.Dial,
	})
	require.NoError(t, err)

	err = ch.SendTest(context.Background(), "user@test.local")
	require.NoError(t, err)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("smtp server did not complete session")
	}
}

func TestSMTPChannelSendWithoutEmailSkips(t *testing.T) {
	t.Parallel()
	ch, err := notifications.NewSMTPChannel(notifications.SMTPConfig{
		Host: "127.0.0.1", Port: 25, From: "a@b.c", UseTLS: false,
	})
	require.NoError(t, err)
	assert.Equal(t, notifications.ChannelSMTP, ch.ID())
	err = ch.Send(context.Background(), notifications.Notification{Subject: "x", Body: "y"})
	require.NoError(t, err)
}

func TestSMTPChannelSendWithSTARTTLS(t *testing.T) {
	t.Parallel()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })

	done := make(chan struct{}, 1)
	go func() {
		conn, acceptErr := ln.Accept()
		if acceptErr != nil {
			return
		}
		defer conn.Close()
		_ = conn.SetDeadline(time.Now().Add(3 * time.Second))
		reader := bufio.NewReader(conn)
		write := func(msg string) { _, _ = conn.Write([]byte(msg)) }
		write("220 localhost ESMTP\r\n")
		for {
			line, readErr := reader.ReadString('\n')
			if readErr != nil {
				return
			}
			cmd := strings.ToUpper(strings.TrimSpace(line))
			switch {
			case strings.HasPrefix(cmd, "EHLO"), strings.HasPrefix(cmd, "HELO"):
				write("250-STARTTLS\r\n250 OK\r\n")
			case cmd == "STARTTLS":
				write("220 ready\r\n")
			case strings.HasPrefix(cmd, "MAIL FROM"):
				write("250 OK\r\n")
			case strings.HasPrefix(cmd, "RCPT TO"):
				write("250 OK\r\n")
			case cmd == "DATA":
				write("354 go\r\n")
			case cmd == "QUIT":
				write("221 bye\r\n")
				done <- struct{}{}
				return
			default:
				if strings.TrimSpace(line) == "." {
					write("250 queued\r\n")
				}
			}
		}
	}()

	_, portStr, splitErr := net.SplitHostPort(ln.Addr().String())
	require.NoError(t, splitErr)

	ch, err := notifications.NewSMTPChannel(notifications.SMTPConfig{
		Host:   "127.0.0.1",
		Port:   mustAtoi(t, portStr),
		From:   "somra@test.local",
		UseTLS: true,
		Dial:   net.Dial,
	})
	require.NoError(t, err)

	err = ch.Send(context.Background(), notifications.Notification{
		Subject: "TLS test",
		Body:    "body",
		Data:    map[string]any{"email": "user@test.local"},
		SentAt:  time.Now(),
	})
	require.NoError(t, err)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("smtp STARTTLS server did not complete session")
	}
}

func TestSMTPChannelSendCancelledContext(t *testing.T) {
	t.Parallel()
	ch, err := notifications.NewSMTPChannel(notifications.SMTPConfig{
		Host: "127.0.0.1", Port: 25, From: "a@b.c", UseTLS: false,
	})
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = ch.Send(ctx, notifications.Notification{
		Subject: "x",
		Body:    "y",
		Data:    map[string]any{"email": "to@test.local"},
	})
	require.Error(t, err)
}

func TestDiscordChannelNilSend(t *testing.T) {
	t.Parallel()
	var ch *notifications.DiscordChannel
	err := ch.Send(context.Background(), notifications.Notification{Subject: "x", Body: "y"})
	require.Error(t, err)
}

func TestSMTPChannelNilSend(t *testing.T) {
	t.Parallel()
	var ch *notifications.SMTPChannel
	err := ch.Send(context.Background(), notifications.Notification{Subject: "x", Body: "y"})
	require.Error(t, err)
}

func mustAtoi(t *testing.T, s string) int {
	t.Helper()
	n := 0
	for _, c := range s {
		n = n*10 + int(c-'0')
	}
	return n
}
