package errors

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_DefaultsMessageKey(t *testing.T) {
	err := New(http.StatusBadRequest, CodeBadRequest, "")
	require.NotNil(t, err)
	assert.Equal(t, CodeBadRequest, err.Code)
	assert.Equal(t, "errors.bad_request", err.MessageKey)
	assert.Equal(t, http.StatusBadRequest, err.HTTPStatus)
}

func TestNew_HonoursExplicitMessageKey(t *testing.T) {
	err := New(http.StatusForbidden, CodeForbidden, "custom.key")
	assert.Equal(t, "custom.key", err.MessageKey)
}

func TestToEnvelope_FromSomraError(t *testing.T) {
	src := New(http.StatusConflict, CodeConflict, "errors.conflict")
	src.Details = map[string]any{"field": "name"}

	status, env := ToEnvelope(src, "req-1")
	assert.Equal(t, http.StatusConflict, status)
	assert.Equal(t, CodeConflict, env.Code)
	assert.Equal(t, "errors.conflict", env.MessageKey)
	assert.Equal(t, "errors.conflict", env.Message)
	assert.Equal(t, "req-1", env.RequestID)
	assert.Equal(t, "name", env.Details["field"])
}

func TestToEnvelope_FallsBackToInternal(t *testing.T) {
	status, env := ToEnvelope(errors.New("boom"), "")
	assert.Equal(t, http.StatusInternalServerError, status)
	assert.Equal(t, CodeInternal, env.Code)
}

func TestToEnvelope_NilStatusUses500(t *testing.T) {
	err := &Error{Code: CodeInternal, MessageKey: "errors.internal"}
	status, _ := ToEnvelope(err, "")
	assert.Equal(t, http.StatusInternalServerError, status)
}

func TestWriteEnvelope_EmitsJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	require.NoError(t, WriteEnvelope(rec, http.StatusTeapot, Envelope{
		Code:       CodeBadRequest,
		MessageKey: "errors.bad_request",
		Message:    "errors.bad_request",
	}))
	assert.Equal(t, http.StatusTeapot, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")

	var decoded Envelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &decoded))
	assert.Equal(t, CodeBadRequest, decoded.Code)
}

func TestJSONEncode_PropagatesError(t *testing.T) {
	err := jsonEncode(failingWriter{}, map[string]any{"x": 1})
	require.Error(t, err)
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("nope") }
