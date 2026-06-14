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

// UserHandlers serves admin user management.
type UserHandlers struct {
	Service *auth.Service
	Users   *db.UserRepo
}

type createUserRequest struct {
	Username string   `json:"username"`
	Password string   `json:"password"`
	Roles    []string `json:"roles"`
}

type updateUserRequest struct {
	Disabled *bool    `json:"disabled"`
	Roles    []string `json:"roles"`
	Password string   `json:"password"`
}

// Mount registers user admin routes (requires auth+admin).
func (h *UserHandlers) Mount(r chi.Router) {
	r.Route("/users", func(r chi.Router) {
		r.Get("/", h.list)
		r.Post("/", h.create)
		r.Route("/{userID}", func(r chi.Router) {
			r.Get("/", h.get)
			r.Put("/", h.update)
		})
	})
}

func (h *UserHandlers) list(w http.ResponseWriter, r *http.Request) {
	users, err := h.Users.List(r.Context())
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "auth.users.list_failed"))
		return
	}
	out := make([]userResponse, 0, len(users))
	for _, u := range users {
		out = append(out, userResponse{ID: u.ID, Username: u.Username, Roles: u.Roles, Disabled: u.Disabled})
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *UserHandlers) create(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "auth.users.invalid"))
		return
	}
	if len(req.Roles) == 0 {
		req.Roles = []string{auth.RoleUser}
	}
	user, err := h.Service.Register(r.Context(), req.Username, req.Password, req.Roles)
	if errors.Is(err, auth.ErrWeakPassword) {
		writeError(w, r, platformerrors.Wrap(err, http.StatusBadRequest, platformerrors.CodeValidation, "auth.password.weak"))
		return
	}
	if errors.Is(err, db.ErrUserAlreadyExists) {
		writeError(w, r, platformerrors.New(http.StatusConflict, platformerrors.CodeConflict, "auth.users.exists"))
		return
	}
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusBadRequest, platformerrors.CodeValidation, "auth.users.create_failed"))
		return
	}
	writeJSON(w, http.StatusCreated, toUserResponse(user))
}

func (h *UserHandlers) get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "userID")
	u, err := h.Users.GetByID(r.Context(), id)
	if errors.Is(err, db.ErrUserNotFound) {
		writeError(w, r, platformerrors.New(http.StatusNotFound, platformerrors.CodeNotFound, "auth.users.not_found"))
		return
	}
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "auth.users.get_failed"))
		return
	}
	writeJSON(w, http.StatusOK, userResponse{ID: u.ID, Username: u.Username, Roles: u.Roles, Disabled: u.Disabled})
}

func (h *UserHandlers) update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "userID")
	var req updateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "auth.users.invalid"))
		return
	}
	if req.Disabled != nil {
		if err := h.Users.SetDisabled(r.Context(), id, *req.Disabled); err != nil {
			writeError(w, r, platformerrors.Wrap(err, http.StatusNotFound, platformerrors.CodeNotFound, "auth.users.not_found"))
			return
		}
	}
	if len(req.Roles) > 0 {
		if err := h.Users.SetRoles(r.Context(), id, req.Roles); err != nil {
			writeError(w, r, platformerrors.Wrap(err, http.StatusBadRequest, platformerrors.CodeValidation, "auth.users.update_failed"))
			return
		}
	}
	if req.Password != "" {
		hash, err := h.Service.HashPassword(req.Password)
		if err != nil {
			writeError(w, r, platformerrors.Wrap(err, http.StatusBadRequest, platformerrors.CodeValidation, "auth.password.weak"))
			return
		}
		if err := h.Users.UpdatePassword(r.Context(), id, hash); err != nil {
			writeError(w, r, platformerrors.Wrap(err, http.StatusNotFound, platformerrors.CodeNotFound, "auth.users.not_found"))
			return
		}
	}
	u, err := h.Users.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "auth.users.get_failed"))
		return
	}
	writeJSON(w, http.StatusOK, userResponse{ID: u.ID, Username: u.Username, Roles: u.Roles, Disabled: u.Disabled})
}
