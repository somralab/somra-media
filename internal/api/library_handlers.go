package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/somralab/somra-media/internal/jobs"
	"github.com/somralab/somra-media/internal/library"
	"github.com/somralab/somra-media/internal/platform/db"
	platformerrors "github.com/somralab/somra-media/internal/platform/errors"
)

// LibraryHandlers serves library CRUD and scan endpoints.
type LibraryHandlers struct {
	Service *library.Service
	Locale  func(*http.Request) string
}

type libraryRequest struct {
	Name         string   `json:"name"`
	Kind         string   `json:"kind"`
	Paths        []string `json:"paths"`
	WatchEnabled *bool    `json:"watchEnabled"`
}

type libraryResponse struct {
	ID           int64    `json:"id"`
	Name         string   `json:"name"`
	Kind         string   `json:"kind"`
	Paths        []string `json:"paths"`
	WatchEnabled bool     `json:"watchEnabled"`
}

func (h *LibraryHandlers) Mount(r chi.Router) {
	r.Route("/libraries", func(r chi.Router) {
		r.Get("/", h.list)
		r.Post("/", h.create)
		r.Route("/{libraryID}", func(r chi.Router) {
			r.Get("/", h.get)
			r.Put("/", h.update)
			r.Delete("/", h.delete)
			r.Post("/scan", h.triggerScan)
			r.Get("/scans", h.listScans)
		})
	})
}

func (h *LibraryHandlers) list(w http.ResponseWriter, r *http.Request) {
	libs, err := h.Service.ListLibraries(r.Context())
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "library.list.failed"))
		return
	}
	out := make([]libraryResponse, 0, len(libs))
	for _, l := range libs {
		out = append(out, toLibraryResponse(l))
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *LibraryHandlers) create(w http.ResponseWriter, r *http.Request) {
	var req libraryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "library.create.invalid"))
		return
	}
	watch := true
	if req.WatchEnabled != nil {
		watch = *req.WatchEnabled
	}
	lib, err := h.Service.CreateLibrary(r.Context(), req.Name, db.LibraryKind(req.Kind), req.Paths, watch)
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusBadRequest, platformerrors.CodeValidation, "library.create.failed"))
		return
	}
	writeJSON(w, http.StatusCreated, toLibraryResponse(lib))
}

func (h *LibraryHandlers) get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "libraryID"))
	if err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "library.id.invalid"))
		return
	}
	lib, err := h.Service.GetLibrary(r.Context(), id)
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusNotFound, platformerrors.CodeNotFound, "library.notFound"))
		return
	}
	writeJSON(w, http.StatusOK, toLibraryResponse(lib))
}

func (h *LibraryHandlers) update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "libraryID"))
	if err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "library.id.invalid"))
		return
	}
	var req libraryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "library.update.invalid"))
		return
	}
	watch := true
	if req.WatchEnabled != nil {
		watch = *req.WatchEnabled
	}
	lib, err := h.Service.UpdateLibrary(r.Context(), id, req.Name, req.Paths, watch)
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusBadRequest, platformerrors.CodeValidation, "library.update.failed"))
		return
	}
	writeJSON(w, http.StatusOK, toLibraryResponse(lib))
}

func (h *LibraryHandlers) delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "libraryID"))
	if err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "library.id.invalid"))
		return
	}
	if err := h.Service.DeleteLibrary(r.Context(), id); err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusNotFound, platformerrors.CodeNotFound, "library.delete.failed"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type scanRequest struct {
	Type string `json:"type"`
}

type scanResponse struct {
	ScanRunID int64       `json:"scanRunId"`
	TaskID    jobs.TaskID `json:"taskId"`
}

func (h *LibraryHandlers) triggerScan(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "libraryID"))
	if err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "library.id.invalid"))
		return
	}
	var req scanRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	scanType := db.ScanFull
	if req.Type == "incremental" {
		scanType = db.ScanIncremental
	}
	runID, taskID, err := h.Service.TriggerScan(r.Context(), id, scanType)
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusConflict, platformerrors.CodeConflict, "library.scan.failed"))
		return
	}
	writeJSON(w, http.StatusAccepted, scanResponse{ScanRunID: runID, TaskID: taskID})
}

func (h *LibraryHandlers) listScans(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "libraryID"))
	if err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "library.id.invalid"))
		return
	}
	runs, err := h.Service.ListScanHistory(r.Context(), id, 20)
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "library.scans.failed"))
		return
	}
	writeJSON(w, http.StatusOK, runs)
}

func toLibraryResponse(l db.Library) libraryResponse {
	return libraryResponse{
		ID: l.ID, Name: l.Name, Kind: string(l.Kind),
		Paths: l.Paths, WatchEnabled: l.WatchEnabled,
	}
}

func parseID(raw string) (int64, error) {
	return strconv.ParseInt(raw, 10, 64)
}
