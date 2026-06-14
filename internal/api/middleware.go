package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	platformerrors "github.com/somralab/somra-media/internal/platform/errors"
	platformlog "github.com/somralab/somra-media/internal/platform/log"
)

// requestIDHeader is the canonical request-id header. The middleware honours
// an inbound value when supplied (useful for downstream reverse proxies that
// assign their own trace ids) and otherwise mints a fresh one.
const requestIDHeader = "X-Request-Id"

type requestIDCtxKey struct{}

// RequestID returns the request id carried on ctx, or "" if absent.
func RequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v, ok := ctx.Value(requestIDCtxKey{}).(string); ok {
		return v
	}
	return ""
}

// newRequestID returns a 16-byte hex token. crypto/rand failures are
// extremely rare; on failure we return an empty string so middleware can
// still proceed (downstream code treats "" as "unknown").
func newRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return ""
	}
	return hex.EncodeToString(b[:])
}

// RequestIDMiddleware injects a request id into the request context and
// echoes it on the response header. Honours any inbound X-Request-Id value
// when it looks reasonable (non-empty, length-bounded) to support upstream
// trace correlation.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSpace(r.Header.Get(requestIDHeader))
		if id == "" || len(id) > 128 {
			id = newRequestID()
		}
		ctx := context.WithValue(r.Context(), requestIDCtxKey{}, id)
		w.Header().Set(requestIDHeader, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// LoggerMiddleware attaches a per-request logger (with request_id field) to
// the context and emits a structured access log entry once the response has
// been written.
//
// Implementation notes:
//   - The logger is captured from the *http.Server's BaseContext if present,
//     otherwise via platformlog.FromContext fallback to slog.Default.
//   - We wrap ResponseWriter to capture status + bytes without affecting
//     streaming semantics (Flusher/Hijacker are preserved).
func LoggerMiddleware(base *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			reqID := RequestID(r.Context())

			logger := base
			if logger == nil {
				logger = platformlog.FromContext(r.Context())
			}
			logger = logger.With(
				slog.String("request_id", reqID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
			)

			ctx := platformlog.WithLogger(r.Context(), logger)

			rw := wrapResponseWriter(w)
			next.ServeHTTP(rw, r.WithContext(ctx))

			logger.LogAttrs(ctx, slog.LevelInfo, "http request",
				slog.Int("status", rw.Status()),
				slog.Int("bytes", rw.Bytes()),
				slog.Duration("duration", time.Since(start)),
			)
		})
	}
}

// RecoverMiddleware converts panics into a 500 JSON envelope and logs the
// stack trace. Without this a panic would terminate the goroutine and leave
// the client hanging on a closed connection.
func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}

			logger := platformlog.FromContext(r.Context())
			logger.ErrorContext(r.Context(), "panic recovered",
				slog.Any("panic", rec),
				slog.String("stack", string(debug.Stack())),
			)

			_, env := platformerrors.ToEnvelope(
				platformerrors.New(http.StatusInternalServerError, platformerrors.CodeInternal, ""),
				RequestID(r.Context()),
			)
			_ = platformerrors.WriteEnvelope(w, http.StatusInternalServerError, env)
		}()
		next.ServeHTTP(w, r)
	})
}

// ContentTypeMiddleware rejects requests whose Content-Type is not
// application/json on methods that carry a body. Bodyless methods (GET,
// HEAD, OPTIONS, DELETE) pass through untouched.
func ContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodDelete:
			next.ServeHTTP(w, r)
			return
		}
		if r.ContentLength == 0 {
			next.ServeHTTP(w, r)
			return
		}
		ct := r.Header.Get("Content-Type")
		if i := strings.IndexByte(ct, ';'); i >= 0 {
			ct = ct[:i]
		}
		ct = strings.TrimSpace(strings.ToLower(ct))
		if ct != "application/json" && ct != "" {
			apiErr := platformerrors.New(http.StatusUnsupportedMediaType, platformerrors.CodeUnsupportedMedia, "")
			status, env := platformerrors.ToEnvelope(apiErr, RequestID(r.Context()))
			_ = platformerrors.WriteEnvelope(w, status, env)
			return
		}
		next.ServeHTTP(w, r)
	})
}
