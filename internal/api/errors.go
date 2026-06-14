package api

import (
	"log/slog"
	"net/http"

	platformerrors "github.com/somralab/somra-media/internal/platform/errors"
	platformlog "github.com/somralab/somra-media/internal/platform/log"
)

// Package-level sentinel errors keep handler code declarative and re-use the
// same MessageKey/Message defaults across every endpoint. Tests rely on the
// stable codes here, not on free-form strings.
var (
	errNotFound        = platformerrors.New(http.StatusNotFound, platformerrors.CodeNotFound, "Resource not found")
	errMethodNotAllow  = platformerrors.New(http.StatusMethodNotAllowed, platformerrors.CodeMethodNotAllowed, "Method not allowed")
	errTooManyRequests = platformerrors.New(http.StatusTooManyRequests, platformerrors.CodeTooManyRequests, "Too many requests")
)

// writeError converts err to the wire envelope, writes it on w and emits a
// structured log line carrying the request id for correlation.
func writeError(w http.ResponseWriter, r *http.Request, err error) {
	reqID := RequestID(r.Context())
	status, env := platformerrors.ToEnvelope(err, reqID)
	if status >= http.StatusInternalServerError {
		platformlog.FromContext(r.Context()).ErrorContext(r.Context(), "api error",
			slog.String("code", string(env.Code)),
			slog.String("message", env.Message),
			slog.Any("error", err),
		)
	}
	_ = platformerrors.WriteJSON(w, status, env)
}
