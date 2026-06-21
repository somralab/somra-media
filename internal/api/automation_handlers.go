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

// AutomationHandlers serves automation search, downloads, quality profiles, and monitors.
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
		r.With(RequirePermission(auth.PermPluginsManage)).Get("/quality-profiles/{profileId}", h.getQualityProfile)
		r.With(RequirePermission(auth.PermPluginsManage)).Patch("/quality-profiles/{profileId}", h.patchQualityProfile)
		r.With(RequirePermission(auth.PermPluginsManage)).Get("/monitors", h.listMonitors)
		r.With(RequirePermission(auth.PermPluginsManage)).Post("/monitors", h.createMonitor)
		r.With(RequirePermission(auth.PermPluginsManage)).Get("/monitors/{monitorId}", h.getMonitor)
		r.With(RequirePermission(auth.PermPluginsManage)).Patch("/monitors/{monitorId}", h.patchMonitor)
		r.With(RequirePermission(auth.PermPluginsManage)).Delete("/monitors/{monitorId}", h.deleteMonitor)
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

func (h *AutomationHandlers) getQualityProfile(w http.ResponseWriter, r *http.Request) {
	if h.AutoRepo == nil {
		writeError(w, r, platformerrors.New(http.StatusServiceUnavailable, platformerrors.CodeInternal, "automation.quality_profiles.get.failed"))
		return
	}
	id, err := parseID(chi.URLParam(r, "profileId"))
	if err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "automation.quality_profiles.id.invalid"))
		return
	}
	row, err := h.AutoRepo.GetQualityProfileByID(r.Context(), id)
	if err != nil {
		writeError(w, r, platformerrors.New(http.StatusNotFound, platformerrors.CodeNotFound, "automation.quality_profiles.not_found"))
		return
	}
	writeJSON(w, http.StatusOK, row)
}

type qualityProfilePatch struct {
	Name      *string `json:"name"`
	Spec      *string `json:"spec"`
	IsDefault *bool   `json:"isDefault"`
}

func (h *AutomationHandlers) patchQualityProfile(w http.ResponseWriter, r *http.Request) {
	if h.AutoRepo == nil {
		writeError(w, r, platformerrors.New(http.StatusServiceUnavailable, platformerrors.CodeInternal, "automation.quality_profiles.update.failed"))
		return
	}
	id, err := parseID(chi.URLParam(r, "profileId"))
	if err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "automation.quality_profiles.id.invalid"))
		return
	}
	var patch qualityProfilePatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "automation.quality_profiles.update.invalid"))
		return
	}
	cur, err := h.AutoRepo.GetQualityProfileByID(r.Context(), id)
	if err != nil {
		writeError(w, r, platformerrors.New(http.StatusNotFound, platformerrors.CodeNotFound, "automation.quality_profiles.not_found"))
		return
	}
	name := cur.Name
	if patch.Name != nil {
		name = *patch.Name
	}
	spec := cur.Spec
	if patch.Spec != nil {
		spec = *patch.Spec
	}
	if err := h.AutoRepo.UpdateQualityProfile(r.Context(), id, name, spec, patch.IsDefault); err != nil {
		if errors.Is(err, db.ErrQualityProfileDuplicate) {
			writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "automation.quality_profiles.update.invalid"))
			return
		}
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "automation.quality_profiles.update.failed"))
		return
	}
	row, err := h.AutoRepo.GetQualityProfileByID(r.Context(), id)
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "automation.quality_profiles.update.failed"))
		return
	}
	writeJSON(w, http.StatusOK, row)
}

func (h *AutomationHandlers) listMonitors(w http.ResponseWriter, r *http.Request) {
	if h.AutoRepo == nil {
		writeError(w, r, platformerrors.New(http.StatusServiceUnavailable, platformerrors.CodeInternal, "automation.monitors.list.failed"))
		return
	}
	rows, err := h.AutoRepo.ListMonitors(r.Context())
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "automation.monitors.list.failed"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"monitors": rows})
}

type monitorInput struct {
	Title          string `json:"title"`
	Provider       string `json:"provider"`
	ExternalID     string `json:"externalId"`
	QualityProfile string `json:"qualityProfile,omitempty"`
	Enabled        *bool  `json:"enabled,omitempty"`
}

func (h *AutomationHandlers) createMonitor(w http.ResponseWriter, r *http.Request) {
	if h.AutoRepo == nil {
		writeError(w, r, platformerrors.New(http.StatusServiceUnavailable, platformerrors.CodeInternal, "automation.monitors.create.failed"))
		return
	}
	ac, ok := auth.FromContext(r.Context())
	if !ok {
		writeError(w, r, platformerrors.New(http.StatusUnauthorized, platformerrors.CodeUnauthorized, "auth.unauthorized"))
		return
	}
	var in monitorInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Title == "" || in.ExternalID == "" {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "automation.monitors.create.invalid"))
		return
	}
	if in.Provider == "" {
		in.Provider = "tmdb"
	}
	enabled := true
	if in.Enabled != nil {
		enabled = *in.Enabled
	}
	id, err := h.AutoRepo.CreateMonitor(r.Context(), db.AutomationMonitor{
		UserID:         ac.Claims.UserID,
		Title:          in.Title,
		Provider:       in.Provider,
		ExternalID:     in.ExternalID,
		QualityProfile: in.QualityProfile,
		Enabled:        enabled,
	})
	if err != nil {
		if errors.Is(err, db.ErrAutomationMonitorDuplicate) {
			writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "automation.monitors.create.invalid"))
			return
		}
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "automation.monitors.create.failed"))
		return
	}
	row, err := h.AutoRepo.GetMonitorByID(r.Context(), id)
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "automation.monitors.create.failed"))
		return
	}
	writeJSON(w, http.StatusCreated, row)
}

func (h *AutomationHandlers) getMonitor(w http.ResponseWriter, r *http.Request) {
	if h.AutoRepo == nil {
		writeError(w, r, platformerrors.New(http.StatusServiceUnavailable, platformerrors.CodeInternal, "automation.monitors.get.failed"))
		return
	}
	id, err := parseID(chi.URLParam(r, "monitorId"))
	if err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "automation.monitors.id.invalid"))
		return
	}
	row, err := h.AutoRepo.GetMonitorByID(r.Context(), id)
	if err != nil {
		writeError(w, r, platformerrors.New(http.StatusNotFound, platformerrors.CodeNotFound, "automation.monitors.not_found"))
		return
	}
	writeJSON(w, http.StatusOK, row)
}

type monitorPatch struct {
	Title          *string `json:"title"`
	QualityProfile *string `json:"qualityProfile"`
	Enabled        *bool   `json:"enabled"`
}

func (h *AutomationHandlers) patchMonitor(w http.ResponseWriter, r *http.Request) {
	if h.AutoRepo == nil {
		writeError(w, r, platformerrors.New(http.StatusServiceUnavailable, platformerrors.CodeInternal, "automation.monitors.update.failed"))
		return
	}
	id, err := parseID(chi.URLParam(r, "monitorId"))
	if err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "automation.monitors.id.invalid"))
		return
	}
	var patch monitorPatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "automation.monitors.update.invalid"))
		return
	}
	if err := h.AutoRepo.PatchMonitor(r.Context(), id, patch.Title, patch.QualityProfile, patch.Enabled); err != nil {
		if errors.Is(err, db.ErrAutomationMonitorNotFound) {
			writeError(w, r, platformerrors.New(http.StatusNotFound, platformerrors.CodeNotFound, "automation.monitors.not_found"))
			return
		}
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "automation.monitors.update.failed"))
		return
	}
	row, err := h.AutoRepo.GetMonitorByID(r.Context(), id)
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "automation.monitors.update.failed"))
		return
	}
	writeJSON(w, http.StatusOK, row)
}

func (h *AutomationHandlers) deleteMonitor(w http.ResponseWriter, r *http.Request) {
	if h.AutoRepo == nil {
		writeError(w, r, platformerrors.New(http.StatusServiceUnavailable, platformerrors.CodeInternal, "automation.monitors.delete.failed"))
		return
	}
	id, err := parseID(chi.URLParam(r, "monitorId"))
	if err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "automation.monitors.id.invalid"))
		return
	}
	if err := h.AutoRepo.DeleteMonitor(r.Context(), id); err != nil {
		if errors.Is(err, db.ErrAutomationMonitorNotFound) {
			writeError(w, r, platformerrors.New(http.StatusNotFound, platformerrors.CodeNotFound, "automation.monitors.not_found"))
			return
		}
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "automation.monitors.delete.failed"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
