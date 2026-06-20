package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/somralab/somra-media/internal/auth"
	indexersearch "github.com/somralab/somra-media/internal/automation/indexer"
	"github.com/somralab/somra-media/internal/platform/db"
	platformerrors "github.com/somralab/somra-media/internal/platform/errors"
	"github.com/somralab/somra-media/internal/plugin"
)

// AutomationHandlers serves automation search, downloads, and quality profiles.
type AutomationHandlers struct {
	AutoRepo *db.AutomationRepo
	Search   *indexersearch.SearchService
}

// Mount registers /automation routes.
func (h *AutomationHandlers) Mount(r chi.Router) {
	r.Route("/automation", func(r chi.Router) {
		r.With(RequirePermission(auth.PermPluginsManage)).Post("/indexers/search", h.searchIndexers)
		r.With(RequirePermission(auth.PermPluginsManage)).Get("/downloads", h.listDownloads)
		r.With(RequirePermission(auth.PermPluginsManage)).Get("/downloads/{downloadId}", h.getDownload)
		r.With(RequirePermission(auth.PermPluginsManage)).Get("/quality-profiles", h.listQualityProfiles)
		r.With(RequirePermission(auth.PermPluginsManage)).Post("/quality-profiles", h.createQualityProfile)
	})
}

type indexerSearchInput struct {
	Title      string           `json:"title"`
	Year       *int             `json:"year,omitempty"`
	MediaKind  plugin.MediaKind `json:"mediaKind"`
	Season     *int             `json:"season,omitempty"`
	Episode    *int             `json:"episode,omitempty"`
	IndexerIDs []int64          `json:"indexerIds,omitempty"`
}

func (h *AutomationHandlers) searchIndexers(w http.ResponseWriter, r *http.Request) {
	if h.Search == nil {
		writeError(w, r, platformerrors.New(http.StatusServiceUnavailable, platformerrors.CodeInternal, "automation.indexers.search.unavailable"))
		return
	}
	var in indexerSearchInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Title == "" || in.MediaKind == "" {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "automation.indexers.search.invalid"))
		return
	}
	resp, err := h.Search.Search(r.Context(), indexersearch.SearchRequest{
		Query: plugin.SearchQuery{
			Title:     in.Title,
			Year:      in.Year,
			MediaKind: in.MediaKind,
			Season:    in.Season,
			Episode:   in.Episode,
		},
		IndexerIDs: in.IndexerIDs,
	})
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "automation.indexers.search.failed"))
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *AutomationHandlers) listDownloads(w http.ResponseWriter, r *http.Request) {
	if h.AutoRepo == nil {
		writeError(w, r, platformerrors.New(http.StatusServiceUnavailable, platformerrors.CodeInternal, "automation.downloads.list.failed"))
		return
	}
	rows, err := h.AutoRepo.ListDownloads(r.Context(), 100)
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "automation.downloads.list.failed"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"downloads": rows})
}

func (h *AutomationHandlers) getDownload(w http.ResponseWriter, r *http.Request) {
	if h.AutoRepo == nil {
		writeError(w, r, platformerrors.New(http.StatusServiceUnavailable, platformerrors.CodeInternal, "automation.downloads.get.failed"))
		return
	}
	id, err := parseID(chi.URLParam(r, "downloadId"))
	if err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "automation.downloads.id.invalid"))
		return
	}
	row, err := h.AutoRepo.GetDownloadByID(r.Context(), id)
	if err != nil {
		writeError(w, r, platformerrors.New(http.StatusNotFound, platformerrors.CodeNotFound, "automation.downloads.not_found"))
		return
	}
	writeJSON(w, http.StatusOK, row)
}

func (h *AutomationHandlers) listQualityProfiles(w http.ResponseWriter, r *http.Request) {
	if h.AutoRepo == nil {
		writeError(w, r, platformerrors.New(http.StatusServiceUnavailable, platformerrors.CodeInternal, "automation.quality_profiles.list.failed"))
		return
	}
	rows, err := h.AutoRepo.ListQualityProfiles(r.Context())
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "automation.quality_profiles.list.failed"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"profiles": rows})
}

type qualityProfileInput struct {
	Name      string `json:"name"`
	Spec      string `json:"spec"`
	IsDefault bool   `json:"isDefault"`
}

func (h *AutomationHandlers) createQualityProfile(w http.ResponseWriter, r *http.Request) {
	if h.AutoRepo == nil {
		writeError(w, r, platformerrors.New(http.StatusServiceUnavailable, platformerrors.CodeInternal, "automation.quality_profiles.create.failed"))
		return
	}
	var in qualityProfileInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Name == "" {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "automation.quality_profiles.create.invalid"))
		return
	}
	id, err := h.AutoRepo.CreateQualityProfile(r.Context(), in.Name, in.Spec, in.IsDefault)
	if err != nil {
		if errors.Is(err, db.ErrQualityProfileDuplicate) {
			writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "automation.quality_profiles.create.invalid"))
			return
		}
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "automation.quality_profiles.create.failed"))
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": id})
}
