// Package log builds the project-wide structured logger on top of the
// standard library's log/slog. Centralising logger construction keeps
// formatting, level parsing and contextual keys consistent across modules.
package log

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

// Format is the log output encoding.
type Format string

const (
	FormatJSON Format = "json"
	FormatText Format = "text"
)

// Options controls logger construction.
type Options struct {
	Level  string
	Format string
	Output io.Writer
}

// New returns a *slog.Logger configured per opts. Unknown level/format values
// fall back to info/json so a misconfigured environment still produces logs.
func New(opts Options) (*slog.Logger, error) {
	out := opts.Output
	if out == nil {
		out = os.Stdout
	}

	level, err := parseLevel(opts.Level)
	if err != nil {
		return nil, err
	}

	handlerOpts := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	switch strings.ToLower(opts.Format) {
	case "", string(FormatJSON):
		handler = slog.NewJSONHandler(out, handlerOpts)
	case string(FormatText):
		handler = slog.NewTextHandler(out, handlerOpts)
	default:
		return nil, fmt.Errorf("log: unknown format %q", opts.Format)
	}

	return slog.New(handler), nil
}

func parseLevel(s string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "info":
		return slog.LevelInfo, nil
	case "debug":
		return slog.LevelDebug, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("log: unknown level %q", s)
	}
}

type loggerCtxKey struct{}

// WithLogger returns a new context carrying logger. Handlers attach a
// per-request logger this way so middleware-applied fields (request id, etc.)
// flow through the call stack without additional plumbing.
func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	if logger == nil {
		return ctx
	}
	return context.WithValue(ctx, loggerCtxKey{}, logger)
}

// FromContext returns the logger stored on ctx, or slog.Default if none is set.
// Callers should treat the returned logger as safe to use immediately.
func FromContext(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return slog.Default()
	}
	if l, ok := ctx.Value(loggerCtxKey{}).(*slog.Logger); ok && l != nil {
		return l
	}
	return slog.Default()
}
