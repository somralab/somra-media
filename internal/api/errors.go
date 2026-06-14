package api

import (
	"log/slog"
	"net/http"

	i18npkg "github.com/somralab/somra-media/internal/platform/i18n"

	platformerrors "github.com/somralab/somra-media/internal/platform/errors"
	platformlog "github.com/somralab/somra-media/internal/platform/log"
)

// Package-level sentinel errors keep handler code declarative and re-use
// the same MessageKey/Message defaults across every endpoint. Tests rely
// on the stable codes here, not on free-form strings.
var (
	errNotFound        = platformerrors.New(http.StatusNotFound, platformerrors.CodeNotFound, "errors.not_found")
	errMethodNotAllow  = platformerrors.New(http.StatusMethodNotAllowed, platformerrors.CodeMethodNotAllowed, "errors.method_not_allowed")
	errTooManyRequests = platformerrors.New(http.StatusTooManyRequests, platformerrors.CodeTooManyRequests, "errors.rate_limited")
)

// writeError converts err to the wire envelope, writes it on w and emits
// a structured log line carrying the request id for correlation. When
// the i18n middleware (api.Options.LocalizerMiddleware) is mounted the
// envelope's Message field is localized using the request's negotiated
// locale; otherwise the message key is echoed verbatim so the response
// is still well-formed.
func writeError(w http.ResponseWriter, r *http.Request, err error) {
	reqID := RequestID(r.Context())
	status, env := platformerrors.ToEnvelope(err, reqID)
	if loc := i18npkg.FromContext(r.Context()); loc != nil {
		env.Message = loc.Message(env.MessageKey, env.Details)
	}
	if status >= http.StatusInternalServerError {
		platformlog.FromContext(r.Context()).ErrorContext(r.Context(), "api error",
			slog.String("code", env.Code),
			slog.String("message", env.Message),
			slog.Any("error", err),
		)
	}
	_ = platformerrors.WriteEnvelope(w, status, env)
}
