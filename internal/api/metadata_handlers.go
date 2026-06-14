package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/somralab/somra-media/internal/metadata"
	"github.com/somralab/somra-media/internal/platform/db"
	platformerrors "github.com/somralab/somra-media/internal/platform/errors"
)

// MediaHandlers serves media item and metadata rematch endpoints.
type MediaHandlers struct {
	DB       *db.DB
	Metadata *metadata.Service
	Locale   func(*http.Request) string
}

type rematchRequest struct {
	Provider   string `json:"provider"`
	ExternalID string `json:"externalId"`
}

func (h *MediaHandlers) Mount(r chi.Router) {
	r.Get("/libraries/{libraryID}/items", h.listItems)
	r.Get("/media-items/{itemID}", h.getItem)
	r.Get("/media-items/{itemID}/match-candidates", h.matchCandidates)
	r.Post("/media-items/{itemID}/rematch", h.rematch)
	r.Post("/libraries/{libraryID}/match", h.autoMatch)
}

func (h *MediaHandlers) listItems(w http.ResponseWriter, r *http.Request) {
	libraryID, err := parseID(chi.URLParam(r, "libraryID"))
	if err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "library.id.invalid"))
		return
	}
	locale := "en-US"
	if h.Locale != nil {
		locale = h.Locale(r)
	}
	items, err := db.NewMediaRepo(h.DB.Querier()).ListItemsByLibrary(r.Context(), libraryID, locale, 100, 0)
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "media.list.failed"))
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *MediaHandlers) getItem(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "itemID"))
	if err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "media.id.invalid"))
		return
	}
	locale := "en-US"
	if h.Locale != nil {
		locale = h.Locale(r)
	}
	item, err := db.NewMediaRepo(h.DB.Querier()).GetItemByID(r.Context(), id, locale)
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusNotFound, platformerrors.CodeNotFound, "media.notFound"))
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *MediaHandlers) matchCandidates(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "itemID"))
	if err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "media.id.invalid"))
		return
	}
	locale := "en-US"
	if h.Locale != nil {
		locale = h.Locale(r)
	}
	results, err := h.Metadata.SearchCandidates(r.Context(), id, locale)
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "media.match.failed"))
		return
	}
	writeJSON(w, http.StatusOK, results)
}

func (h *MediaHandlers) rematch(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "itemID"))
	if err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "media.id.invalid"))
		return
	}
	var req rematchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Provider == "" || req.ExternalID == "" {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "media.rematch.invalid"))
		return
	}
	locale := "en-US"
	if h.Locale != nil {
		locale = h.Locale(r)
	}
	if err := h.Metadata.Rematch(r.Context(), id, req.Provider, req.ExternalID, locale); err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusBadRequest, platformerrors.CodeValidation, "media.rematch.failed"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *MediaHandlers) autoMatch(w http.ResponseWriter, r *http.Request) {
	libraryID, err := parseID(chi.URLParam(r, "libraryID"))
	if err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "library.id.invalid"))
		return
	}
	locale := "en-US"
	if h.Locale != nil {
		locale = h.Locale(r)
	}
	n, err := h.Metadata.AutoMatch(r.Context(), libraryID, locale, 100)
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "media.automatch.failed"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"matched": n})
}
