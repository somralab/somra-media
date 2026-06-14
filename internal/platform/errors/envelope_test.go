package errors_test

import (
	"encoding/json"
	stderrors "errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"

	errpkg "github.com/somralab/somra-media/internal/platform/errors"
	"github.com/somralab/somra-media/internal/platform/i18n"
)

func newBundle(t *testing.T) *i18n.Bundle {
	t.Helper()
	b, err := i18n.NewBundle()
	require.NoError(t, err)
	return b
}

func TestErrorMethods(t *testing.T) {
	t.Parallel()
	cause := stderrors.New("io broken")
	e := errpkg.ErrInternal.WithCause(cause)
	require.NotSame(t, errpkg.ErrInternal, e, "WithCause must clone")
	require.True(t, stderrors.Is(e, errpkg.ErrInternal))
	require.ErrorIs(t, e, cause)
	require.Contains(t, e.Error(), "internal")
	require.Contains(t, e.Error(), "io broken")

	require.Equal(t, "<nil>", (*errpkg.Error)(nil).Error())

	withDetails := errpkg.ErrBadRequest.WithDetails(map[string]any{"field": "name"})
	withMore := withDetails.WithDetails(map[string]any{"hint": "non-empty"})
	require.Equal(t, "name", withMore.Details["field"])
	require.Equal(t, "non-empty", withMore.Details["hint"])
	require.Nil(t, errpkg.ErrBadRequest.Details, "sentinel must remain untouched")

	require.False(t, e.Is(stderrors.New("other")))
	require.Nil(t, (*errpkg.Error)(nil).WithCause(stderrors.New("x")))
	require.Nil(t, (*errpkg.Error)(nil).WithDetails(map[string]any{"a": 1}))
}

func TestWriteJSONEnglish(t *testing.T) {
	t.Parallel()
	b := newBundle(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	loc := b.Localize(language.AmericanEnglish)
	errpkg.WriteJSON(rec, req, errpkg.ErrNotFound, loc)

	require.Equal(t, http.StatusNotFound, rec.Code)
	require.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

	var env errpkg.Envelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &env))
	require.Equal(t, "not_found", env.Code)
	require.Equal(t, "errors.not_found", env.MessageKey)
	require.Equal(t, "The requested resource was not found.", env.Message)
}

func TestWriteJSONTurkishFromContext(t *testing.T) {
	t.Parallel()
	b := newBundle(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Language", "tr-TR")
	req.Header.Set("X-Request-Id", "req-42")

	b.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errpkg.WriteJSON(w, r, errpkg.ErrForbidden, nil)
	})).ServeHTTP(rec, req)

	require.Equal(t, http.StatusForbidden, rec.Code)
	var env errpkg.Envelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &env))
	require.Equal(t, "forbidden", env.Code)
	require.Equal(t, "Bu işlem için yetkiniz yok.", env.Message)
	require.Equal(t, "req-42", env.RequestID)
}

func TestWriteJSONWrapsUnknownError(t *testing.T) {
	t.Parallel()
	b := newBundle(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	loc := b.Localize(language.AmericanEnglish)

	errpkg.WriteJSON(rec, req, stderrors.New("oops"), loc)
	require.Equal(t, http.StatusInternalServerError, rec.Code)

	var env errpkg.Envelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &env))
	require.Equal(t, "internal", env.Code)
}

func TestWriteJSONTemplatedDetails(t *testing.T) {
	t.Parallel()
	b := newBundle(t)
	loc := b.Localize(language.MustParse("tr-TR"))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	err := (&errpkg.Error{
		Code:       "bad_request",
		MessageKey: "errors.validation.required",
		HTTPStatus: http.StatusBadRequest,
	}).WithDetails(map[string]any{"Field": "username"})

	errpkg.WriteJSON(rec, req, err, loc)
	var env errpkg.Envelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &env))
	require.Equal(t, "username alanı zorunludur.", env.Message)
	require.Equal(t, "username", env.Details["Field"])
}

func TestWriteJSONNoLocalizerEchoesKey(t *testing.T) {
	t.Parallel()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	errpkg.WriteJSON(rec, req, errpkg.ErrRateLimited, nil)
	require.Equal(t, http.StatusTooManyRequests, rec.Code)
	var env errpkg.Envelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &env))
	require.Equal(t, "errors.rate_limited", env.Message)
}
