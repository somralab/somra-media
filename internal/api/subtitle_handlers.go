package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/somralab/somra-media/internal/auth"
	"github.com/somralab/somra-media/internal/platform/db"
	platformerrors "github.com/somralab/somra-media/internal/platform/errors"
	"github.com/somralab/somra-media/internal/subtitles"
)

// SubtitleHandlers serves subtitle search/download/upload endpoints.
type SubtitleHandlers struct {
	Service *subtitles.Service
}

func (h *SubtitleHandlers) Mount(r chi.Router) {
	r.Route("/subtitles", func(r chi.Router) {
		r.With(RequirePermission(auth.PermLibraryRead)).Get("/search", h.searchGet)
		r.With(RequirePermission(auth.PermLibraryRead)).Post("/search", h.searchPost)
		r.With(RequirePermission(auth.PermLibraryWrite)).Post("/download", h.download)
		r.With(RequirePermission(auth.PermLibraryWrite)).Post("/upload", h.upload)
	})
	r.With(RequirePermission(auth.PermLibraryRead)).Get("/media-items/{itemId}/subtitles", h.list)
}

func (h *SubtitleHandlers) searchGet(w http.ResponseWriter, r *http.Request) {
	itemID, err := strconv.ParseInt(r.URL.Query().Get("mediaItemId"), 10, 64)
	if err != nil || itemID <= 0 {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "subtitles.search.invalid"))
		return
	}
	results, err := h.Service.Search(r.Context(), itemID, r.URL.Query().Get("language"), r.URL.Query().Get("query"))
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusBadGateway, platformerrors.CodeInternal, "subtitles.search.failed"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"results": results})
}

func (h *SubtitleHandlers) searchPost(w http.ResponseWriter, r *http.Request) {
	var body struct {
		MediaItemID int64  `json:"mediaItemId"`
		Language    string `json:"language"`
		Query       string `json:"query"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "subtitles.search.invalid"))
		return
	}
	results, err := h.Service.Search(r.Context(), body.MediaItemID, body.Language, body.Query)
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusBadGateway, platformerrors.CodeInternal, "subtitles.search.failed"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"results": results})
}

func (h *SubtitleHandlers) download(w http.ResponseWriter, r *http.Request) {
	var body struct {
		MediaItemID int64  `json:"mediaItemId"`
		Provider    string `json:"provider"`
		ExternalID  string `json:"externalId"`
		Language    string `json:"language"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "subtitles.download.invalid"))
		return
	}
	file, err := h.Service.Download(r.Context(), body.MediaItemID, body.Provider, body.ExternalID, body.Language)
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusBadGateway, platformerrors.CodeInternal, "subtitles.download.failed"))
		return
	}
	writeJSON(w, http.StatusCreated, toManagedSubtitle(file))
}

func (h *SubtitleHandlers) upload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(4 << 20); err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "subtitles.upload.invalid"))
		return
	}
	itemID, err := strconv.ParseInt(r.FormValue("mediaItemId"), 10, 64)
	if err != nil || itemID <= 0 {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "subtitles.upload.invalid"))
		return
	}
	language := r.FormValue("language")
	if language == "" {
		language = "en"
	}
	f, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "subtitles.upload.invalid"))
		return
	}
	defer func() { _ = f.Close() }()
	content, err := io.ReadAll(io.LimitReader(f, 2<<20))
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusBadRequest, platformerrors.CodeValidation, "subtitles.upload.invalid"))
		return
	}
	file, err := h.Service.Upload(r.Context(), itemID, language, content)
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "subtitles.upload.failed"))
		return
	}
	writeJSON(w, http.StatusCreated, toManagedSubtitle(file))
}

func (h *SubtitleHandlers) list(w http.ResponseWriter, r *http.Request) {
	itemID, err := strconv.ParseInt(chi.URLParam(r, "itemId"), 10, 64)
	if err != nil || itemID <= 0 {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "subtitles.list.invalid"))
		return
	}
	files, err := h.Service.List(r.Context(), itemID)
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "subtitles.list.failed"))
		return
	}
	out := make([]map[string]any, 0, len(files))
	for _, f := range files {
		out = append(out, toManagedSubtitle(f))
	}
	writeJSON(w, http.StatusOK, map[string]any{"subtitles": out})
}

func toManagedSubtitle(f db.SubtitleFile) map[string]any {
	m := map[string]any{
		"id":          f.ID,
		"mediaItemId": f.MediaItemID,
		"language":    f.Language,
		"source":      string(f.Source),
	}
	if f.Path != "" {
		m["path"] = f.Path
	}
	if f.Provider != "" {
		m["provider"] = f.Provider
	}
	if !f.CreatedAt.IsZero() {
		m["createdAt"] = f.CreatedAt.UTC().Format("2006-01-02T15:04:05Z")
	}
	return m
}
