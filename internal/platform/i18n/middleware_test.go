package i18n_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/platform/i18n"
)

func TestMiddlewareAttachesLocalizer(t *testing.T) {
	t.Parallel()
	b, err := i18n.NewBundle()
	require.NoError(t, err)

	captured := make(chan string, 1)
	handler := b.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loc := i18n.FromContext(r.Context())
		require.NotNil(t, loc)
		captured <- loc.Message("errors.internal", nil)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Language", "tr-TR")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "Beklenmeyen bir sunucu hatası oluştu.", <-captured)
}

func TestMiddlewareFallsBackToEnglish(t *testing.T) {
	t.Parallel()
	b, err := i18n.NewBundle()
	require.NoError(t, err)

	captured := make(chan string, 1)
	handler := b.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loc := i18n.FromContext(r.Context())
		captured <- loc.Message("errors.internal", nil)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, "An internal error occurred.", <-captured)
}

func TestMiddlewareHonoursOverrides(t *testing.T) {
	t.Parallel()
	b, err := i18n.NewBundle()
	require.NoError(t, err)

	captured := make(chan string, 1)
	handler := b.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loc := i18n.FromContext(r.Context())
		captured <- loc.Message("errors.internal", nil)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Language", "en-US")
	ctx := i18n.WithLocaleOverrides(req.Context(), "tr-TR", "")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, "Beklenmeyen bir sunucu hatası oluştu.", <-captured)
}

func TestFromContextEmpty(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	require.Nil(t, i18n.FromContext(req.Context()))
	require.Equal(t, i18n.SourceLanguage, i18n.TagFromContext(req.Context()))
}

func TestTagFromContext(t *testing.T) {
	t.Parallel()
	b, err := i18n.NewBundle()
	require.NoError(t, err)

	var seen string
	handler := b.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = i18n.TagFromContext(r.Context()).String()
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Language", "tr-TR")
	handler.ServeHTTP(httptest.NewRecorder(), req)
	require.Equal(t, "tr-TR", seen)
}
