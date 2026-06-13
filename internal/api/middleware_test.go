package api

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/platform/config"
	platformerrors "github.com/somralab/somra-media/internal/platform/errors"
)

func TestRequestIDMiddleware_GeneratesWhenAbsent(t *testing.T) {
	h := RequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := RequestID(r.Context())
		assert.NotEmpty(t, id)
		_, _ = w.Write([]byte(id))
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	headerID := rec.Header().Get(requestIDHeader)
	assert.NotEmpty(t, headerID)
	assert.Equal(t, headerID, rec.Body.String())
}

func TestRequestIDMiddleware_HonoursInbound(t *testing.T) {
	const sentID = "trace-abc-123"
	h := RequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, sentID, RequestID(r.Context()))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(requestIDHeader, sentID)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.Equal(t, sentID, rec.Header().Get(requestIDHeader))
}

func TestRequestIDMiddleware_RejectsAbusivelyLong(t *testing.T) {
	long := strings.Repeat("a", 256)
	h := RequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := RequestID(r.Context())
		assert.NotEqual(t, long, got)
		assert.NotEmpty(t, got)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(requestIDHeader, long)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
}

func TestRecoverMiddleware_ConvertsPanicTo500(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	chain := RequestIDMiddleware(LoggerMiddleware(logger)(RecoverMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		panic("boom")
	}))))

	rec := httptest.NewRecorder()
	chain.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	require.Equal(t, http.StatusInternalServerError, rec.Code)

	var env platformerrors.Envelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &env))
	assert.Equal(t, platformerrors.CodeInternal, env.Code)
	assert.Equal(t, "errors.internal_error", env.MessageKey)
	assert.NotEmpty(t, env.RequestID)
	assert.Contains(t, buf.String(), "panic recovered")
}

func TestCORSMiddleware_Preflight(t *testing.T) {
	cfg := config.CORSConfig{
		AllowedOrigins: []string{"https://app.example"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Authorization", "Content-Type"},
		MaxAge:         60,
	}

	r := chi.NewRouter()
	r.Use(corsMiddleware(cfg))
	r.Get("/api/v1/health", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/health", nil)
	req.Header.Set("Origin", "https://app.example")
	req.Header.Set("Access-Control-Request-Method", "GET")
	req.Header.Set("Access-Control-Request-Headers", "Authorization")

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, "https://app.example", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, rec.Header().Get("Access-Control-Allow-Methods"), "GET")
	assert.Contains(t, rec.Header().Get("Access-Control-Allow-Headers"), "Authorization")
}

func TestNotFoundReturnsEnvelope(t *testing.T) {
	h := newTestHandler(t, Options{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/does-not-exist", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
	var env platformerrors.Envelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &env))
	assert.Equal(t, platformerrors.CodeNotFound, env.Code)
	assert.NotEmpty(t, env.RequestID)
}

func TestMethodNotAllowedReturnsEnvelope(t *testing.T) {
	h := newTestHandler(t, Options{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	var env platformerrors.Envelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &env))
	assert.Equal(t, platformerrors.CodeMethodNotAllowed, env.Code)
}

type rejectAll struct{}

func (rejectAll) Allow(*http.Request) bool { return false }

func TestRateLimitMiddleware_RejectsWhenLimiterDenies(t *testing.T) {
	h := newTestHandler(t, Options{RateLimiter: rejectAll{}})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusTooManyRequests, rec.Code)
	var env platformerrors.Envelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &env))
	assert.Equal(t, platformerrors.CodeTooManyRequests, env.Code)
}

func TestContentTypeMiddleware_RejectsNonJSONBody(t *testing.T) {
	r := chi.NewRouter()
	r.Use(RequestIDMiddleware)
	r.Use(ContentTypeMiddleware)
	r.Post("/echo", func(w http.ResponseWriter, req *http.Request) {
		_, _ = io.Copy(w, req.Body)
	})

	req := httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader("plain"))
	req.Header.Set("Content-Type", "text/plain")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnsupportedMediaType, rec.Code)
}

func TestContentTypeMiddleware_AcceptsJSON(t *testing.T) {
	r := chi.NewRouter()
	r.Use(RequestIDMiddleware)
	r.Use(ContentTypeMiddleware)
	r.Post("/echo", func(w http.ResponseWriter, req *http.Request) {
		_, _ = io.Copy(w, req.Body)
	})

	req := httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader(`{"a":1}`))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, `{"a":1}`, rec.Body.String())
}
