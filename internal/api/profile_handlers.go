package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/somralab/somra-media/internal/auth"
	"github.com/somralab/somra-media/internal/platform/db"
	platformerrors "github.com/somralab/somra-media/internal/platform/errors"
)

// ProfileHandlers serves profile CRUD for the authenticated user.
type ProfileHandlers struct {
	Profiles *db.ProfileRepo
}

type profileResponse struct {
	UserID           string  `json:"userId"`
	Locale           string  `json:"locale"`
	Theme            string  `json:"theme"`
	AvatarURL        string  `json:"avatarUrl,omitempty"`
	MaxContentRating *string `json:"maxContentRating,omitempty"`
	IsChild          bool    `json:"isChild"`
}

type profileUpdateRequest struct {
	Locale           *string `json:"locale"`
	Theme            *string `json:"theme"`
	AvatarURL        *string `json:"avatarUrl"`
	MaxContentRating *string `json:"maxContentRating"`
	IsChild          *bool   `json:"isChild"`
}

// Mount registers profile routes (requires auth).
func (h *ProfileHandlers) Mount(r chi.Router) {
	r.Route("/profile", func(r chi.Router) {
		r.Get("/", h.get)
		r.Put("/", h.update)
	})
}

func (h *ProfileHandlers) get(w http.ResponseWriter, r *http.Request) {
	ac, ok := auth.FromContext(r.Context())
	if !ok {
		writeError(w, r, platformerrors.New(http.StatusUnauthorized, platformerrors.CodeUnauthorized, "auth.errors.unauthorized"))
		return
	}
	p, err := h.Profiles.Get(r.Context(), ac.Claims.UserID)
	if errors.Is(err, db.ErrProfileNotFound) {
		writeError(w, r, platformerrors.New(http.StatusNotFound, platformerrors.CodeNotFound, "auth.profile.not_found"))
		return
	}
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "auth.profile.get_failed"))
		return
	}
	writeJSON(w, http.StatusOK, toProfileResponse(p))
}

func (h *ProfileHandlers) update(w http.ResponseWriter, r *http.Request) {
	ac, ok := auth.FromContext(r.Context())
	if !ok {
		writeError(w, r, platformerrors.New(http.StatusUnauthorized, platformerrors.CodeUnauthorized, "auth.errors.unauthorized"))
		return
	}
	p, err := h.Profiles.Get(r.Context(), ac.Claims.UserID)
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "auth.profile.get_failed"))
		return
	}
	var req profileUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "auth.profile.invalid"))
		return
	}
	if req.Locale != nil {
		p.Locale = *req.Locale
	}
	if req.Theme != nil {
		p.Theme = *req.Theme
	}
	if req.AvatarURL != nil {
		p.AvatarURL = *req.AvatarURL
	}
	if req.MaxContentRating != nil {
		p.MaxContentRating = req.MaxContentRating
	}
	if req.IsChild != nil {
		p.IsChild = *req.IsChild
	}
	if err := h.Profiles.Update(r.Context(), p); err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusBadRequest, platformerrors.CodeValidation, "auth.profile.update_failed"))
		return
	}
	writeJSON(w, http.StatusOK, toProfileResponse(p))
}

func toProfileResponse(p db.UserProfile) profileResponse {
	return profileResponse{
		UserID:           p.UserID,
		Locale:           p.Locale,
		Theme:            p.Theme,
		AvatarURL:        p.AvatarURL,
		MaxContentRating: p.MaxContentRating,
		IsChild:          p.IsChild,
	}
}
