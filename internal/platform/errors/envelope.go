// Package errors defines Somra's i18n-aware HTTP error envelope.
//
// Handlers create an *Error (or wrap an existing one with WithCause /
// WithDetails) and pass it to WriteJSON together with the per-request
// Localizer. The envelope is intentionally small: code (machine-stable
// short string), messageKey (i18n key), message (localized text),
// details (free-form, optional).
//
// The envelope shape is shared with the API gateway (Paket 1); this
// package's WriteJSON layers locale negotiation on top.
package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	i18npkg "github.com/somralab/somra-media/internal/platform/i18n"
)

func jsonEncode(w io.Writer, v any) error {
	return json.NewEncoder(w).Encode(v)
}

// Error is the canonical Somra error type. It implements the standard
// error interface and supports errors.Is / errors.Unwrap so callers can
// use the usual idioms.
type Error struct {
	Code       string
	MessageKey string
	HTTPStatus int
	Details    map[string]any
	Cause      error
}

// Error implements the error interface. The returned string is intended
// for logs, not end users — use WriteJSON for user-facing responses.
func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.MessageKey, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.MessageKey)
}

// Unwrap supports errors.Unwrap.
func (e *Error) Unwrap() error { return e.Cause }

// Is allows two *Error values to be compared by Code only, which is the
// stable contract callers depend on.
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	return e != nil && t != nil && e.Code == t.Code
}

// WithCause returns a copy of e with Cause set; the originating Error
// is left untouched so package-level sentinels stay immutable.
func (e *Error) WithCause(cause error) *Error {
	if e == nil {
		return nil
	}
	clone := *e
	clone.Cause = cause
	return &clone
}

// WithDetails returns a copy of e with Details merged in.
func (e *Error) WithDetails(details map[string]any) *Error {
	if e == nil {
		return nil
	}
	clone := *e
	if clone.Details == nil {
		clone.Details = make(map[string]any, len(details))
	} else {
		merged := make(map[string]any, len(clone.Details)+len(details))
		for k, v := range clone.Details {
			merged[k] = v
		}
		clone.Details = merged
	}
	for k, v := range details {
		clone.Details[k] = v
	}
	return &clone
}

// Predefined sentinel errors mapped to i18n keys. Use *.WithCause to
// attach the underlying error or *.WithDetails to attach validation
// information without losing the canonical Code/MessageKey pair.
var (
	ErrInternal = &Error{
		Code:       "internal",
		MessageKey: "errors.internal",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrBadRequest = &Error{
		Code:       "bad_request",
		MessageKey: "errors.bad_request",
		HTTPStatus: http.StatusBadRequest,
	}
	ErrNotFound = &Error{
		Code:       "not_found",
		MessageKey: "errors.not_found",
		HTTPStatus: http.StatusNotFound,
	}
	ErrUnauthorized = &Error{
		Code:       "unauthorized",
		MessageKey: "errors.unauthorized",
		HTTPStatus: http.StatusUnauthorized,
	}
	ErrForbidden = &Error{
		Code:       "forbidden",
		MessageKey: "errors.forbidden",
		HTTPStatus: http.StatusForbidden,
	}
	ErrRateLimited = &Error{
		Code:       "rate_limited",
		MessageKey: "errors.rate_limited",
		HTTPStatus: http.StatusTooManyRequests,
	}
)

// Envelope is the JSON shape returned by WriteJSON. The fields are
// public so other packages can unmarshal responses in tests without
// duplicating struct definitions.
type Envelope struct {
	Code       string         `json:"code"`
	MessageKey string         `json:"messageKey"`
	Message    string         `json:"message"`
	RequestID  string         `json:"requestId,omitempty"`
	Details    map[string]any `json:"details,omitempty"`
}

// WriteJSON serializes err as a JSON envelope on w. If err is nil the
// response is a generic 500 — callers should always provide an error.
//
// Localization order:
//   - localizer (caller-supplied) is preferred,
//   - then i18n.FromContext(r.Context()),
//   - finally the MessageKey is echoed verbatim so the response is
//     still well-formed even without an active locale.
func WriteJSON(w http.ResponseWriter, r *http.Request, err error, localizer *i18npkg.Localizer) {
	var ee *Error
	if !errors.As(err, &ee) {
		ee = ErrInternal.WithCause(err)
	}
	if localizer == nil && r != nil {
		localizer = i18npkg.FromContext(r.Context())
	}
	message := ee.MessageKey
	if localizer != nil {
		// Field name passed through Details; nil-safe.
		message = localizer.Message(ee.MessageKey, ee.Details)
	}

	env := Envelope{
		Code:       ee.Code,
		MessageKey: ee.MessageKey,
		Message:    message,
		Details:    ee.Details,
	}
	if r != nil {
		if rid := r.Header.Get("X-Request-Id"); rid != "" {
			env.RequestID = rid
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	status := ee.HTTPStatus
	if status == 0 {
		status = http.StatusInternalServerError
	}
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(env)
}
