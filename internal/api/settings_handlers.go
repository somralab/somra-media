package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/somralab/somra-media/internal/auth"
	platformerrors "github.com/somralab/somra-media/internal/platform/errors"
	"github.com/somralab/somra-media/internal/settings"
)

// SystemHandlers serves system detection endpoints.
type SystemHandlers struct {
	DataDir   string
	CacheDir  string
	FFmpegBin string
}

// Mount registers public system routes.
func (h *SystemHandlers) Mount(r chi.Router) {
	r.Get("/system/detect", h.detect)
}

func (h *SystemHandlers) detect(w http.ResponseWriter, r *http.Request) {
	paths := []string{h.DataDir, h.CacheDir}
	if raw := r.URL.Query().Get("paths"); raw != "" {
		paths = strings.Split(raw, ",")
	}
	profile := settings.DetectSystemWithFFmpeg(paths, h.FFmpegBin)
	writeJSON(w, http.StatusOK, profile)
}

// SettingsHandlers serves central settings API.
type SettingsHandlers struct {
	Service   *settings.Service
	OnPatched func(ctx context.Context, category string) error
}

func (h *SettingsHandlers) Mount(r chi.Router) {
	r.Route("/settings", func(r chi.Router) {
		r.With(RequirePermission(auth.PermUsersManage)).Get("/", h.getAll)
		r.With(RequirePermission(auth.PermUsersManage)).Patch("/{category}", h.patchCategory)
	})
}

func (h *SettingsHandlers) getAll(w http.ResponseWriter, r *http.Request) {
	snap, err := h.Service.GetAll(r.Context())
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "settings.get.failed"))
		return
	}
	writeJSON(w, http.StatusOK, snap)
}

func (h *SettingsHandlers) patchCategory(w http.ResponseWriter, r *http.Request) {
	category := chi.URLParam(r, "category")
	var patch map[string]any
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "settings.patch.invalid"))
		return
	}
	out, err := h.Service.PatchCategory(r.Context(), category, patch)
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusBadRequest, platformerrors.CodeValidation, "settings.patch.failed"))
		return
	}
	if h.OnPatched != nil {
		if syncErr := h.OnPatched(r.Context(), category); syncErr != nil {
			writeError(w, r, platformerrors.Wrap(syncErr, http.StatusInternalServerError, platformerrors.CodeInternal, "settings.sync.failed"))
			return
		}
	}
	writeJSON(w, http.StatusOK, out)
}

// OnboardingHandlers serves wizard state endpoints.
type OnboardingHandlers struct {
	Onboarding *settings.Onboarding
}

func (h *OnboardingHandlers) Mount(r chi.Router) {
	r.Get("/onboarding/status", h.status)
	r.Post("/onboarding/step", h.step)
}

// MountProtected registers onboarding routes that require authentication.
func (h *OnboardingHandlers) MountProtected(r chi.Router) {
	r.With(RequirePermission(auth.PermUsersManage)).Post("/onboarding/complete", h.complete)
}

func (h *OnboardingHandlers) status(w http.ResponseWriter, r *http.Request) {
	state, err := h.Onboarding.Status(r.Context())
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "onboarding.status.failed"))
		return
	}
	writeJSON(w, http.StatusOK, state)
}

type onboardingStepBody struct {
	Phase         string `json:"phase"`
	Locale        string `json:"locale"`
	ApplyDefaults *bool  `json:"applyDefaults"`
	LibraryID     int64  `json:"libraryId"`
}

func (h *OnboardingHandlers) step(w http.ResponseWriter, r *http.Request) {
	var body onboardingStepBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "onboarding.step.invalid"))
		return
	}
	apply := false
	if body.ApplyDefaults != nil {
		apply = *body.ApplyDefaults
	}
	state, err := h.Onboarding.AdvanceStep(r.Context(), settings.StepRequest{
		Phase:         body.Phase,
		Locale:        body.Locale,
		ApplyDefaults: apply,
		LibraryID:     body.LibraryID,
	})
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusBadRequest, platformerrors.CodeValidation, "onboarding.step.failed"))
		return
	}
	writeJSON(w, http.StatusOK, state)
}

func (h *OnboardingHandlers) complete(w http.ResponseWriter, r *http.Request) {
	if err := h.Onboarding.Complete(r.Context()); err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "onboarding.complete.failed"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
