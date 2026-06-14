package api

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/platform/db"
	"github.com/somralab/somra-media/internal/subtitles"
)

func TestSubtitleHandlers_DirectValidation(t *testing.T) {
	h := &SubtitleHandlers{Service: &subtitles.Service{}}

	rec := httptest.NewRecorder()
	h.searchGet(rec, httptest.NewRequest(http.MethodGet, "/?mediaItemId=bad", nil))
	require.Equal(t, http.StatusBadRequest, rec.Code)

	rec = httptest.NewRecorder()
	h.searchPost(rec, httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(`{`))))
	require.Equal(t, http.StatusBadRequest, rec.Code)

	rec = httptest.NewRecorder()
	h.download(rec, httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(`{`))))
	require.Equal(t, http.StatusBadRequest, rec.Code)

	rec = httptest.NewRecorder()
	h.upload(rec, httptest.NewRequest(http.MethodPost, "/", nil))
	require.Equal(t, http.StatusBadRequest, rec.Code)

	rec = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("itemId", "nope")
	h.list(rec, req.WithContext(contextWithChi(req.Context(), rctx)))
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func contextWithChi(ctx context.Context, rctx *chi.Context) context.Context {
	return context.WithValue(ctx, chi.RouteCtxKey, rctx)
}

func TestToManagedSubtitle_Fields(t *testing.T) {
	out := toManagedSubtitle(db.SubtitleFile{
		ID: 1, MediaItemID: 2, Language: "en", Source: db.SubtitleUploaded,
		Path: "/tmp/s.srt", Provider: "manual",
	})
	require.EqualValues(t, 1, out["id"])
	require.Equal(t, "/tmp/s.srt", out["path"])
	require.Equal(t, "manual", out["provider"])
}
