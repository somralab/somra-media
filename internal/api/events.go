package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	platformerrors "github.com/somralab/somra-media/internal/platform/errors"
	platformlog "github.com/somralab/somra-media/internal/platform/log"
)

// defaultSSEHeartbeat is how often the SSE handler emits an SSE comment to
// keep proxies and load balancers from closing idle connections.
const defaultSSEHeartbeat = 15 * time.Second

// sseEventsHandler implements GET /api/v1/events/stream. The endpoint is a
// scaffold: it sends a single "hello" event on connect and then a heartbeat
// comment on a fixed interval until the client disconnects. Real business
// events are emitted in Sprint 05 (Paket 5 backend core / Sprint 02 library).
func sseEventsHandler(heartbeat time.Duration) http.HandlerFunc {
	if heartbeat <= 0 {
		heartbeat = defaultSSEHeartbeat
	}
	return func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			writeError(w, r, platformerrors.New(http.StatusInternalServerError, platformerrors.CodeInternal, "streaming not supported"))
			return
		}

		if !acceptsEventStream(r) {
			writeError(w, r, platformerrors.New(http.StatusNotAcceptable, platformerrors.CodeUnsupportedMedia, "Accept must include text/event-stream"))
			return
		}

		w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store, no-transform")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")
		w.WriteHeader(http.StatusOK)

		logger := platformlog.FromContext(r.Context())

		if err := writeSSEEvent(w, "hello", fmt.Sprintf(`{"requestId":%q}`, RequestID(r.Context()))); err != nil {
			logger.WarnContext(r.Context(), "sse: initial event write failed", slog.Any("error", err))
			return
		}
		flusher.Flush()

		ticker := time.NewTicker(heartbeat)
		defer ticker.Stop()

		ctx := r.Context()
		for {
			select {
			case <-ctx.Done():
				logger.DebugContext(ctx, "sse: client disconnected")
				return
			case <-ticker.C:
				if err := writeSSEComment(w, "ping"); err != nil {
					return
				}
				flusher.Flush()
			}
		}
	}
}

// acceptsEventStream returns true when the request's Accept header allows
// text/event-stream. A missing Accept header is treated as "accept anything"
// to keep curl-friendly defaults.
func acceptsEventStream(r *http.Request) bool {
	accept := strings.TrimSpace(r.Header.Get("Accept"))
	if accept == "" {
		return true
	}
	lower := strings.ToLower(accept)
	return strings.Contains(lower, "text/event-stream") || strings.Contains(lower, "*/*")
}

// writeSSEEvent writes a single named event with a JSON payload. Returns an
// error so the caller can stop the loop on broken pipe.
func writeSSEEvent(w http.ResponseWriter, name, data string) error {
	if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", name, data); err != nil {
		return fmt.Errorf("sse: write event %q: %w", name, err)
	}
	return nil
}

func writeSSEComment(w http.ResponseWriter, comment string) error {
	if _, err := fmt.Fprintf(w, ": %s\n\n", comment); err != nil {
		return fmt.Errorf("sse: write comment: %w", err)
	}
	return nil
}
