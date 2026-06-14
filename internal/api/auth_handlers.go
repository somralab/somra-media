package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/somralab/somra-media/internal/auth"
	platformerrors "github.com/somralab/somra-media/internal/platform/errors"
)

const refreshTokenCookie = "somra_refresh_token"

// AuthHandlers serves authentication endpoints.
type AuthHandlers struct {
	Service      *auth.Service
	SecureCookie bool
}

type loginRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	DeviceLabel string `json:"deviceLabel"`
}

type tokenResponse struct {
	AccessToken string       `json:"accessToken"`
	ExpiresAt   time.Time    `json:"expiresAt"`
	User        userResponse `json:"user"`
}

type userResponse struct {
	ID       string   `json:"id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	Disabled bool     `json:"disabled"`
}

type setupAdminRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Mount registers auth routes on r.
func (h *AuthHandlers) Mount(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", h.login)
		r.Post("/refresh", h.refresh)
		r.Post("/logout", h.logout)
	})
	r.Get("/setup/status", h.setupStatus)
	r.Post("/setup/admin", h.setupAdmin)
	r.Group(func(r chi.Router) {
		r.Use(h.requireAuth().Middleware)
		r.Get("/auth/sessions", h.listSessions)
		r.Delete("/auth/sessions/{sessionID}", h.revokeSession)
	})
}

func (h *AuthHandlers) requireAuth() *AuthMiddleware {
	return &AuthMiddleware{Service: h.Service}
}

func (h *AuthHandlers) setupStatus(w http.ResponseWriter, r *http.Request) {
	required, err := h.Service.SetupRequired(r.Context())
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "auth.setup.status_failed"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"setupRequired": required})
}

func (h *AuthHandlers) setupAdmin(w http.ResponseWriter, r *http.Request) {
	var req setupAdminRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "auth.setup.invalid"))
		return
	}
	user, pair, err := h.Service.CreateAdmin(r.Context(), req.Username, req.Password)
	if errors.Is(err, auth.ErrSetupComplete) {
		writeError(w, r, platformerrors.New(http.StatusConflict, platformerrors.CodeConflict, "auth.setup.already_complete"))
		return
	}
	if errors.Is(err, auth.ErrWeakPassword) {
		writeError(w, r, platformerrors.Wrap(err, http.StatusBadRequest, platformerrors.CodeValidation, "auth.password.weak"))
		return
	}
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusBadRequest, platformerrors.CodeValidation, "auth.setup.failed"))
		return
	}
	h.setRefreshCookie(w, pair.RefreshToken)
	writeJSON(w, http.StatusCreated, tokenResponse{
		AccessToken: pair.AccessToken,
		ExpiresAt:   pair.ExpiresAt,
		User:        toUserResponse(user),
	})
}

func (h *AuthHandlers) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "auth.login.invalid"))
		return
	}
	user, pair, err := h.Service.Login(r.Context(), req.Username, req.Password, req.DeviceLabel, auth.ClientIP(r))
	if errors.Is(err, auth.ErrAccountLocked) {
		writeError(w, r, platformerrors.New(http.StatusTooManyRequests, platformerrors.CodeTooManyRequests, "auth.login.locked"))
		return
	}
	if errors.Is(err, auth.ErrInvalidCredentials) {
		writeError(w, r, platformerrors.New(http.StatusUnauthorized, platformerrors.CodeUnauthorized, "auth.login.invalid_credentials"))
		return
	}
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "auth.login.failed"))
		return
	}
	h.setRefreshCookie(w, pair.RefreshToken)
	writeJSON(w, http.StatusOK, tokenResponse{
		AccessToken: pair.AccessToken,
		ExpiresAt:   pair.ExpiresAt,
		User:        toUserResponse(user),
	})
}

func (h *AuthHandlers) refresh(w http.ResponseWriter, r *http.Request) {
	secret := h.refreshFromRequest(r)
	if secret == "" {
		writeError(w, r, platformerrors.New(http.StatusUnauthorized, platformerrors.CodeUnauthorized, "auth.refresh.missing"))
		return
	}
	pair, err := h.Service.Refresh(r.Context(), secret)
	if errors.Is(err, auth.ErrRevokedToken) || errors.Is(err, auth.ErrTokenNotFound) || errors.Is(err, auth.ErrInvalidToken) {
		h.clearRefreshCookie(w)
		writeError(w, r, platformerrors.New(http.StatusUnauthorized, platformerrors.CodeUnauthorized, "auth.refresh.invalid"))
		return
	}
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "auth.refresh.failed"))
		return
	}
	h.setRefreshCookie(w, pair.RefreshToken)
	writeJSON(w, http.StatusOK, map[string]any{
		"accessToken": pair.AccessToken,
		"expiresAt":   pair.ExpiresAt,
	})
}

func (h *AuthHandlers) logout(w http.ResponseWriter, r *http.Request) {
	secret := h.refreshFromRequest(r)
	if secret != "" {
		_ = h.Service.Logout(r.Context(), secret)
	}
	h.clearRefreshCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandlers) listSessions(w http.ResponseWriter, r *http.Request) {
	ac, ok := auth.FromContext(r.Context())
	if !ok {
		writeError(w, r, platformerrors.New(http.StatusUnauthorized, platformerrors.CodeUnauthorized, "auth.errors.unauthorized"))
		return
	}
	sessions, err := h.Service.ListSessions(r.Context(), ac.Claims.UserID)
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "auth.sessions.list_failed"))
		return
	}
	out := make([]map[string]any, 0, len(sessions))
	for _, s := range sessions {
		out = append(out, map[string]any{
			"id":          s.ID,
			"deviceLabel": s.DeviceLabel,
			"createdAt":   s.CreatedAt,
			"lastUsedAt":  s.LastUsedAt,
			"revokedAt":   s.RevokedAt,
			"expiresAt":   s.ExpiresAt,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *AuthHandlers) revokeSession(w http.ResponseWriter, r *http.Request) {
	ac, ok := auth.FromContext(r.Context())
	if !ok {
		writeError(w, r, platformerrors.New(http.StatusUnauthorized, platformerrors.CodeUnauthorized, "auth.errors.unauthorized"))
		return
	}
	sessionID := chi.URLParam(r, "sessionID")
	if err := h.Service.RevokeSession(r.Context(), ac.Claims.UserID, sessionID); err != nil {
		if errors.Is(err, auth.ErrForbidden) {
			writeError(w, r, platformerrors.New(http.StatusForbidden, platformerrors.CodeForbidden, "auth.errors.forbidden"))
			return
		}
		writeError(w, r, platformerrors.Wrap(err, http.StatusNotFound, platformerrors.CodeNotFound, "auth.sessions.not_found"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandlers) refreshFromRequest(r *http.Request) string {
	if c, err := r.Cookie(refreshTokenCookie); err == nil && c.Value != "" {
		return c.Value
	}
	var body struct {
		RefreshToken string `json:"refreshToken"`
	}
	if r.Body != nil && r.ContentLength != 0 {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}
	return body.RefreshToken
}

func (h *AuthHandlers) setRefreshCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenCookie,
		Value:    token,
		Path:     "/api/v1/auth",
		HttpOnly: true,
		Secure:   h.SecureCookie,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int((7 * 24 * time.Hour).Seconds()),
	})
}

func (h *AuthHandlers) clearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenCookie,
		Value:    "",
		Path:     "/api/v1/auth",
		HttpOnly: true,
		Secure:   h.SecureCookie,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}

func toUserResponse(u auth.UserAccount) userResponse {
	return userResponse{ID: u.ID, Username: u.Username, Roles: u.Roles, Disabled: u.Disabled}
}
