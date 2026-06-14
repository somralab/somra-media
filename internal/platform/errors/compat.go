package errors

import (
	"errors"
	"net/http"
)

// Code is a stable, machine-readable identifier for an error class. The
// string values are part of the public API contract and must not be
// renamed without bumping the API version. This typed alias was
// introduced in Paket 1 (API gateway); Paket 5's i18n-aware Error type
// is the canonical representation — Code constants below are kept as
// thin compatibility helpers so existing handlers remain idiomatic.
type Code = string

const (
	CodeBadRequest         Code = "bad_request"
	CodeValidation         Code = "validation_error"
	CodeUnauthorized       Code = "unauthorized"
	CodeForbidden          Code = "forbidden"
	CodeNotFound           Code = "not_found"
	CodeMethodNotAllowed   Code = "method_not_allowed"
	CodeConflict           Code = "conflict"
	CodeUnsupportedMedia   Code = "unsupported_media_type"
	CodeTooManyRequests    Code = "rate_limited"
	CodeInternal           Code = "internal"
	CodeServiceUnavailable Code = "service_unavailable"
	CodeNotImplemented     Code = "not_implemented"
)

// New constructs an i18n-aware *Error with the supplied HTTP status,
// machine code, and message key. The message key follows the
// "errors.<code>" convention by default when an empty key is passed.
func New(status int, code Code, messageKey string) *Error {
	if messageKey == "" {
		messageKey = "errors." + code
	}
	return &Error{
		Code:       code,
		MessageKey: messageKey,
		HTTPStatus: status,
	}
}

// Wrap attaches cause to a new Error with the given status, code and key.
func Wrap(cause error, status int, code Code, messageKey string) *Error {
	return New(status, code, messageKey).WithCause(cause)
}

// ToEnvelope projects err onto the Envelope wire shape without
// localization. Callers that have a Localizer should prefer WriteJSON,
// which performs locale negotiation as a side-effect of writing.
//
// requestID is optional; when non-empty it is echoed so clients can
// correlate logs with API calls.
func ToEnvelope(err error, requestID string) (int, Envelope) {
	var ee *Error
	if !errors.As(err, &ee) || ee == nil {
		ee = ErrInternal
	}
	status := ee.HTTPStatus
	if status == 0 {
		status = http.StatusInternalServerError
	}
	return status, Envelope{
		Code:       ee.Code,
		MessageKey: ee.MessageKey,
		Message:    ee.MessageKey,
		RequestID:  requestID,
		Details:    ee.Details,
	}
}

// WriteEnvelope writes a pre-built Envelope to w with the given status.
// This is the unlocalized counterpart of WriteJSON and matches the
// signature handlers in Paket 1 were originally written against.
func WriteEnvelope(w http.ResponseWriter, status int, env Envelope) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	return jsonEncode(w, env)
}
