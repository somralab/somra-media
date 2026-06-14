package errors

import (
	"encoding/json"
	stderrors "errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIError_ErrorFormatting(t *testing.T) {
	err := New(http.StatusBadRequest, CodeBadRequest, "missing field")
	assert.Contains(t, err.Error(), "bad_request")
	assert.Contains(t, err.Error(), "missing field")

	wrapped := err.Wrap(stderrors.New("inner"))
	assert.Contains(t, wrapped.Error(), "inner")
	assert.ErrorIs(t, wrapped, wrapped.Wrapped)
}

func TestAPIError_WithDetail_CopyOnWrite(t *testing.T) {
	base := New(http.StatusBadRequest, CodeBadRequest, "bad")
	withDetail := base.WithDetail("field", "username")

	assert.Empty(t, base.Details)
	assert.Equal(t, "username", withDetail.Details["field"])
}

func TestToEnvelope_KnownError(t *testing.T) {
	err := New(http.StatusNotFound, CodeNotFound, "library not found")
	status, env := ToEnvelope(err, "req-1")

	assert.Equal(t, http.StatusNotFound, status)
	assert.Equal(t, CodeNotFound, env.Code)
	assert.Equal(t, "errors.not_found", env.MessageKey)
	assert.Equal(t, "library not found", env.Message)
	assert.Equal(t, "req-1", env.RequestID)
}

func TestToEnvelope_UnknownErrorBecomesInternal(t *testing.T) {
	status, env := ToEnvelope(stderrors.New("boom"), "req-2")

	assert.Equal(t, http.StatusInternalServerError, status)
	assert.Equal(t, CodeInternal, env.Code)
	assert.Equal(t, "errors.internal_error", env.MessageKey)
	assert.Equal(t, "req-2", env.RequestID)
}

func TestWriteJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	require.NoError(t, WriteJSON(rec, http.StatusTeapot, Envelope{
		Code:       CodeInternal,
		MessageKey: "errors.internal_error",
		Message:    "broken",
	}))

	assert.Equal(t, http.StatusTeapot, rec.Code)
	assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

	var got Envelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.Equal(t, CodeInternal, got.Code)
	assert.Equal(t, "broken", got.Message)
}
