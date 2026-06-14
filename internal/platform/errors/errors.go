// Package errors defines the project-wide error envelope and a small set of
// canonical error codes. The envelope shape is fixed at the API surface so
// the frontend can render localized messages from messageKey and so future
// packets (auth, library, streaming) can map domain errors to a single
// transport contract.
//
// i18n: messageKey points at the future server-side catalog (owned by
// Sprint 01 Paket 5). Message is the resolved string in the negotiated
// locale; until Paket 5 wires the catalog, handlers fill Message with the
// English source text so clients still receive a non-empty body.
package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// Code is a stable, machine-readable identifier returned to API clients.
// The string value is part of the public contract; do not rename existing
// codes without bumping the API version.
type Code string

const (
	CodeBadRequest         Code = "bad_request"
	CodeUnauthorized       Code = "unauthorized"
	CodeForbidden          Code = "forbidden"
	CodeNotFound           Code = "not_found"
	CodeMethodNotAllowed   Code = "method_not_allowed"
	CodeConflict           Code = "conflict"
	CodeUnsupportedMedia   Code = "unsupported_media_type"
	CodeTooManyRequests    Code = "too_many_requests"
	CodeInternal           Code = "internal_error"
	CodeServiceUnavailable Code = "service_unavailable"
	CodeNotImplemented     Code = "not_implemented"
)

// Envelope is the JSON shape returned for every API error. Fields are flat
// so they survive being embedded in larger payloads (e.g. batch responses).
type Envelope struct {
	Code       Code              `json:"code"`
	MessageKey string            `json:"messageKey"`
	Message    string            `json:"message"`
	Details    map[string]string `json:"details,omitempty"`
	RequestID  string            `json:"requestId,omitempty"`
}

// APIError is the internal representation handlers raise. It implements error
// so it composes with errors.Is/As; convert to Envelope at the HTTP boundary
// via ToEnvelope.
type APIError struct {
	Status     int
	Code       Code
	MessageKey string
	Message    string
	Details    map[string]string
	Wrapped    error
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Wrapped != nil {
		return fmt.Sprintf("%s: %s: %s", e.Code, e.Message, e.Wrapped.Error())
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the wrapped cause so errors.Is/As traverse the chain.
func (e *APIError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Wrapped
}

// WithDetail returns a copy of e with key=value appended to Details. Allows
// handlers to add field-level context without mutating shared errors.
func (e *APIError) WithDetail(key, value string) *APIError {
	if e == nil {
		return nil
	}
	cp := *e
	cp.Details = make(map[string]string, len(e.Details)+1)
	for k, v := range e.Details {
		cp.Details[k] = v
	}
	cp.Details[key] = value
	return &cp
}

// Wrap returns a copy of e wrapping cause. Use when bubbling up an internal
// error you also want to surface to the client.
func (e *APIError) Wrap(cause error) *APIError {
	if e == nil {
		return nil
	}
	cp := *e
	cp.Wrapped = cause
	return &cp
}

// New constructs an APIError with the given status, code, and English source
// message. messageKey conventionally takes the form "errors.<code>".
func New(status int, code Code, message string) *APIError {
	return &APIError{
		Status:     status,
		Code:       code,
		MessageKey: defaultMessageKey(code),
		Message:    message,
	}
}

func defaultMessageKey(code Code) string {
	return "errors." + string(code)
}

// ToEnvelope projects err onto the wire envelope. requestID is optional and
// when non-empty is echoed so clients can correlate logs with API calls.
func ToEnvelope(err error, requestID string) (int, Envelope) {
	var api *APIError
	if errors.As(err, &api) && api != nil {
		return api.Status, Envelope{
			Code:       api.Code,
			MessageKey: api.MessageKey,
			Message:    api.Message,
			Details:    api.Details,
			RequestID:  requestID,
		}
	}
	return http.StatusInternalServerError, Envelope{
		Code:       CodeInternal,
		MessageKey: defaultMessageKey(CodeInternal),
		Message:    http.StatusText(http.StatusInternalServerError),
		RequestID:  requestID,
	}
}

// WriteJSON serialises the envelope onto w with the correct status and
// content type. Failures to encode are logged-and-ignored by the caller.
func WriteJSON(w http.ResponseWriter, status int, env Envelope) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(env); err != nil {
		return fmt.Errorf("errors: encode envelope: %w", err)
	}
	return nil
}
