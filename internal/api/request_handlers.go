package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/somralab/somra-media/internal/auth"
	"github.com/somralab/somra-media/internal/notifications"
	"github.com/somralab/somra-media/internal/platform/db"
	platformerrors "github.com/somralab/somra-media/internal/platform/errors"
	"github.com/somralab/somra-media/internal/requests"
)

// RequestHandlers serves content request CRUD and workflow endpoints.
type RequestHandlers struct {
	Repo       *db.RequestRepo
	Users      *db.UserRepo
	Collision  *requests.CollisionChecker
	Policy     *requests.PolicyService
	Discoverer *requests.Discoverer
	Handoff    requests.AutomationHandoff
	Notify     *notifications.Dispatcher
	StatusPub  RequestStatusPublisher
	Locale     func(*http.Request) string
	Now        func() time.Time
}

// Mount registers /requests routes.
func (h *RequestHandlers) Mount(r chi.Router) {
	r.Route("/requests", func(r chi.Router) {
		r.With(RequirePermission(auth.PermRequestsRead)).Get("/", h.list)
		r.With(RequirePermission(auth.PermRequestsCreate)).Post("/", h.create)
		r.With(RequirePermission(auth.PermRequestsCreate)).Get("/discover", h.discover)
		r.With(RequirePermission(auth.PermRequestsRead)).Get("/policies", h.getPolicies)
		r.With(RequirePermission(auth.PermRequestsManage)).Patch("/policies", h.patchPolicies)

		r.Route("/{requestId}", func(r chi.Router) {
			r.With(RequirePermission(auth.PermRequestsRead)).Get("/", h.get)
			r.With(RequirePermission(auth.PermRequestsRead)).Patch("/", h.patch)
			r.With(RequirePermission(auth.PermRequestsManage)).Post("/approve", h.approve)
			r.With(RequirePermission(auth.PermRequestsManage)).Post("/reject", h.reject)
			r.With(RequirePermission(auth.PermRequestsRead)).Post("/cancel", h.cancel)
		})
	})
}

func (h *RequestHandlers) now() time.Time {
	if h != nil && h.Now != nil {
		return h.Now()
	}
	return time.Now().UTC()
}

func (h *RequestHandlers) locale(r *http.Request) string {
	if h != nil && h.Locale != nil {
		if loc := h.Locale(r); loc != "" {
			return loc
		}
	}
	return "en-US"
}

func (h *RequestHandlers) list(w http.ResponseWriter, r *http.Request) {
	ac, ok := auth.FromContext(r.Context())
	if !ok {
		writeError(w, r, platformerrors.New(http.StatusUnauthorized, platformerrors.CodeUnauthorized, "auth.errors.unauthorized"))
		return
	}
	filter := db.RequestListFilter{
		Limit:  queryInt(r, "limit", 50),
		Offset: queryInt(r, "offset", 0),
	}
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 50
	}
	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = db.RequestStatus(status)
	}
	if userFilter := r.URL.Query().Get("userId"); userFilter != "" {
		if !auth.HasPermission(ac, auth.PermRequestsManage) {
			writeError(w, r, platformerrors.New(http.StatusForbidden, platformerrors.CodeForbidden, "auth.errors.forbidden"))
			return
		}
		filter.UserID = userFilter
	} else if !auth.HasPermission(ac, auth.PermRequestsManage) {
		filter.UserID = ac.Claims.UserID
	}

	rows, err := h.Repo.List(r.Context(), filter)
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "requests.list.failed"))
		return
	}
	if rows == nil {
		rows = []db.Request{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"requests": rows})
}

type createRequestInput struct {
	MediaKind         string `json:"mediaKind"`
	Provider          string `json:"provider"`
	ExternalID        string `json:"externalId"`
	Title             string `json:"title"`
	PosterURL         string `json:"posterUrl"`
	QualityResolution string `json:"qualityResolution"`
	QualityProfile    string `json:"qualityProfile"`
}

func (h *RequestHandlers) create(w http.ResponseWriter, r *http.Request) {
	ac, ok := auth.FromContext(r.Context())
	if !ok {
		writeError(w, r, platformerrors.New(http.StatusUnauthorized, platformerrors.CodeUnauthorized, "auth.errors.unauthorized"))
		return
	}
	var in createRequestInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "requests.create.invalid"))
		return
	}
	in.Provider = strings.TrimSpace(in.Provider)
	in.ExternalID = strings.TrimSpace(in.ExternalID)
	in.Title = strings.TrimSpace(in.Title)
	if in.MediaKind == "" || in.Provider == "" || in.ExternalID == "" || in.Title == "" {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "requests.create.invalid"))
		return
	}

	ctx := r.Context()
	if h.Collision != nil {
		if err := h.Collision.ValidateCreation(ctx, in.Provider, in.ExternalID); err != nil {
			h.writeCollisionError(w, r, err)
			return
		}
	}
	if h.Policy != nil {
		if err := h.Policy.ValidateQuota(ctx, ac.Claims.UserID, ac.Claims.Roles); err != nil {
			if errors.Is(err, requests.ErrQuotaExceeded) {
				writeError(w, r, platformerrors.New(http.StatusForbidden, platformerrors.CodeForbidden, "requests.quota.exceeded"))
				return
			}
			writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "requests.create.failed"))
			return
		}
	}

	status := db.RequestStatusPending
	if h.Policy != nil {
		decision, err := h.Policy.Evaluate(ctx, ac.Claims.UserID, ac.Claims.Roles)
		if err != nil {
			writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "requests.create.failed"))
			return
		}
		if decision.AutoApprove {
			status = db.RequestStatusApproved
		}
	}

	resolution := db.RequestQualityResolution(in.QualityResolution)
	if resolution == "" {
		resolution = db.RequestQualityAny
	}

	id, err := h.Repo.Create(ctx, db.Request{
		UserID:            ac.Claims.UserID,
		MediaKind:         db.RequestMediaKind(in.MediaKind),
		Provider:          in.Provider,
		ExternalID:        in.ExternalID,
		Title:             in.Title,
		PosterURL:         in.PosterURL,
		QualityResolution: resolution,
		QualityProfile:    in.QualityProfile,
		Status:            status,
	})
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "requests.create.failed"))
		return
	}

	row, err := h.Repo.GetByID(ctx, id)
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "requests.create.failed"))
		return
	}

	domain := requests.FromDBRequest(row)
	if status == db.RequestStatusApproved {
		if err := h.runHandoff(ctx, domain); err != nil {
			writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "requests.handoff.failed"))
			return
		}
	}

	h.publishStatus(row)
	h.notifyCreated(ctx, domain)
	if status == db.RequestStatusApproved {
		h.notifyApproved(ctx, domain)
	}

	writeJSON(w, http.StatusCreated, row)
}

func (h *RequestHandlers) discover(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "requests.discover.query_required"))
		return
	}
	kind := requests.MediaKind(r.URL.Query().Get("kind"))
	if kind == "" {
		kind = requests.MediaKindMovie
	}
	if h.Discoverer == nil {
		writeError(w, r, platformerrors.New(http.StatusServiceUnavailable, platformerrors.CodeInternal, "requests.discover.unavailable"))
		return
	}
	hits, err := h.Discoverer.Search(r.Context(), requests.DiscoverSearchParams{
		Query:  q,
		Kind:   kind,
		Locale: h.locale(r),
		Limit:  20,
	})
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusBadGateway, platformerrors.CodeInternal, "requests.discover.failed"))
		return
	}
	results := make([]map[string]any, 0, len(hits))
	for _, hit := range hits {
		results = append(results, map[string]any{
			"mediaKind":  string(kind),
			"provider":   hit.Provider,
			"externalId": hit.ExternalID,
			"title":      hit.Title,
			"posterUrl":  hit.PosterURL,
			"inLibrary":  false,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"results": results})
}

func (h *RequestHandlers) getPolicies(w http.ResponseWriter, r *http.Request) {
	row, err := h.Repo.GetPolicy(r.Context())
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "requests.policies.get.failed"))
		return
	}
	writeJSON(w, http.StatusOK, policyResponse(row))
}

type requestPoliciesPatch struct {
	AutoApproveRoles  []string       `json:"autoApproveRoles"`
	UserQuotaPerMonth *int           `json:"userQuotaPerMonth"`
	AdminSettings     map[string]any `json:"adminSettings"`
}

func (h *RequestHandlers) patchPolicies(w http.ResponseWriter, r *http.Request) {
	var patch requestPoliciesPatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "requests.policies.patch.invalid"))
		return
	}
	cur, err := h.Repo.GetPolicy(r.Context())
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "requests.policies.get.failed"))
		return
	}
	if patch.AutoApproveRoles != nil {
		b, err := json.Marshal(patch.AutoApproveRoles)
		if err != nil {
			writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "requests.policies.patch.invalid"))
			return
		}
		cur.AutoApproveRoles = string(b)
	}
	if patch.UserQuotaPerMonth != nil {
		if *patch.UserQuotaPerMonth < 0 {
			writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "requests.policies.patch.invalid"))
			return
		}
		cur.UserQuotaPerMonth = *patch.UserQuotaPerMonth
	}
	if patch.AdminSettings != nil {
		b, err := json.Marshal(patch.AdminSettings)
		if err != nil {
			writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "requests.policies.patch.invalid"))
			return
		}
		cur.AdminSettings = string(b)
	}
	if err := h.Repo.UpsertPolicy(r.Context(), cur); err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "requests.policies.patch.failed"))
		return
	}
	updated, err := h.Repo.GetPolicy(r.Context())
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "requests.policies.get.failed"))
		return
	}
	writeJSON(w, http.StatusOK, policyResponse(updated))
}

func (h *RequestHandlers) get(w http.ResponseWriter, r *http.Request) {
	ac, ok := auth.FromContext(r.Context())
	if !ok {
		writeError(w, r, platformerrors.New(http.StatusUnauthorized, platformerrors.CodeUnauthorized, "auth.errors.unauthorized"))
		return
	}
	id, err := parseID(chi.URLParam(r, "requestId"))
	if err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "requests.id.invalid"))
		return
	}
	row, err := h.Repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, db.ErrRequestNotFound) {
			writeError(w, r, platformerrors.New(http.StatusNotFound, platformerrors.CodeNotFound, "requests.not_found"))
			return
		}
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "requests.get.failed"))
		return
	}
	if !canAccessRequest(ac, row) {
		writeError(w, r, platformerrors.New(http.StatusForbidden, platformerrors.CodeForbidden, "auth.errors.forbidden"))
		return
	}
	writeJSON(w, http.StatusOK, row)
}

type requestPatchBody struct {
	QualityResolution *string `json:"qualityResolution"`
	QualityProfile    *string `json:"qualityProfile"`
	AdminNote         *string `json:"adminNote"`
}

func (h *RequestHandlers) patch(w http.ResponseWriter, r *http.Request) {
	ac, ok := auth.FromContext(r.Context())
	if !ok {
		writeError(w, r, platformerrors.New(http.StatusUnauthorized, platformerrors.CodeUnauthorized, "auth.errors.unauthorized"))
		return
	}
	id, err := parseID(chi.URLParam(r, "requestId"))
	if err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "requests.id.invalid"))
		return
	}
	row, err := h.Repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, db.ErrRequestNotFound) {
			writeError(w, r, platformerrors.New(http.StatusNotFound, platformerrors.CodeNotFound, "requests.not_found"))
			return
		}
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "requests.patch.failed"))
		return
	}
	if !canAccessRequest(ac, row) {
		writeError(w, r, platformerrors.New(http.StatusForbidden, platformerrors.CodeForbidden, "auth.errors.forbidden"))
		return
	}
	var body requestPatchBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "requests.patch.invalid"))
		return
	}
	patch := db.RequestUpdate{}
	if body.QualityResolution != nil {
		res := db.RequestQualityResolution(*body.QualityResolution)
		patch.QualityResolution = &res
	}
	if body.QualityProfile != nil {
		patch.QualityProfile = body.QualityProfile
	}
	if body.AdminNote != nil {
		if !auth.HasPermission(ac, auth.PermRequestsManage) {
			writeError(w, r, platformerrors.New(http.StatusForbidden, platformerrors.CodeForbidden, "auth.errors.forbidden"))
			return
		}
		patch.AdminNote = body.AdminNote
	}
	if patch.QualityResolution == nil && patch.QualityProfile == nil && patch.AdminNote == nil {
		writeJSON(w, http.StatusOK, row)
		return
	}
	if err := h.Repo.Update(r.Context(), id, patch); err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "requests.patch.failed"))
		return
	}
	updated, err := h.Repo.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "requests.patch.failed"))
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

type requestActionInput struct {
	AdminNote string `json:"adminNote"`
}

func (h *RequestHandlers) approve(w http.ResponseWriter, r *http.Request) {
	h.transition(w, r, db.RequestStatusApproved, true)
}

func (h *RequestHandlers) reject(w http.ResponseWriter, r *http.Request) {
	h.transition(w, r, db.RequestStatusRejected, false)
}

func (h *RequestHandlers) cancel(w http.ResponseWriter, r *http.Request) {
	ac, ok := auth.FromContext(r.Context())
	if !ok {
		writeError(w, r, platformerrors.New(http.StatusUnauthorized, platformerrors.CodeUnauthorized, "auth.errors.unauthorized"))
		return
	}
	id, err := parseID(chi.URLParam(r, "requestId"))
	if err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "requests.id.invalid"))
		return
	}
	row, err := h.Repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, db.ErrRequestNotFound) {
			writeError(w, r, platformerrors.New(http.StatusNotFound, platformerrors.CodeNotFound, "requests.not_found"))
			return
		}
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "requests.cancel.failed"))
		return
	}
	if row.UserID != ac.Claims.UserID && !auth.HasPermission(ac, auth.PermRequestsManage) {
		writeError(w, r, platformerrors.New(http.StatusForbidden, platformerrors.CodeForbidden, "auth.errors.forbidden"))
		return
	}
	h.transition(w, r, db.RequestStatusCancelled, false)
}

func (h *RequestHandlers) transition(w http.ResponseWriter, r *http.Request, target db.RequestStatus, handoffOnApprove bool) {
	id, err := parseID(chi.URLParam(r, "requestId"))
	if err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "requests.id.invalid"))
		return
	}
	row, err := h.Repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, db.ErrRequestNotFound) {
			writeError(w, r, platformerrors.New(http.StatusNotFound, platformerrors.CodeNotFound, "requests.not_found"))
			return
		}
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "requests.transition.failed"))
		return
	}

	var note *string
	if r.ContentLength > 0 {
		var action requestActionInput
		if err := json.NewDecoder(r.Body).Decode(&action); err == nil && action.AdminNote != "" {
			note = &action.AdminNote
		}
	}

	domain := requests.FromDBRequest(row)
	if err := requests.TransitionTo(&domain, requests.Status(target), h.now()); err != nil {
		writeError(w, r, platformerrors.New(http.StatusConflict, platformerrors.CodeConflict, "requests.transition.invalid"))
		return
	}

	st := db.RequestStatus(domain.Status)
	patch := db.RequestUpdate{Status: &st}
	if note != nil {
		patch.AdminNote = note
	}
	if err := h.Repo.Update(r.Context(), id, patch); err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "requests.transition.failed"))
		return
	}
	updated, err := h.Repo.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "requests.transition.failed"))
		return
	}
	domain = requests.FromDBRequest(updated)

	if handoffOnApprove && target == db.RequestStatusApproved {
		if err := h.runHandoff(r.Context(), domain); err != nil {
			writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "requests.handoff.failed"))
			return
		}
	}

	h.publishStatus(updated)
	switch target {
	case db.RequestStatusApproved:
		h.notifyApproved(r.Context(), domain)
	case db.RequestStatusRejected:
		h.notifyRejected(r.Context(), domain)
	case db.RequestStatusCancelled:
		// no notification hook for cancel in scope
	case db.RequestStatusCompleted:
		h.notifyCompleted(r.Context(), domain)
	}

	writeJSON(w, http.StatusOK, updated)
}

func (h *RequestHandlers) runHandoff(ctx context.Context, req requests.Request) error {
	if h.Handoff == nil {
		h.Handoff = requests.NoOpAutomationHandoff{}
	}
	return h.Handoff.Handoff(ctx, req)
}

func (h *RequestHandlers) publishStatus(row db.Request) {
	h.StatusPub.PublishRequestStatus(RequestStatusEvent{
		RequestID: row.ID,
		Status:    string(row.Status),
		UpdatedAt: row.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

func (h *RequestHandlers) adminIDs(ctx context.Context) []string {
	if h.Users == nil {
		return nil
	}
	accounts, err := h.Users.List(ctx)
	if err != nil {
		return nil
	}
	var ids []string
	for _, u := range accounts {
		for _, role := range u.Roles {
			if role == auth.RoleAdmin {
				ids = append(ids, u.ID)
				break
			}
		}
	}
	return ids
}

func (h *RequestHandlers) notifyCreated(ctx context.Context, req requests.Request) {
	if h.Notify == nil {
		return
	}
	_ = h.Notify.HandleRequestCreated(ctx, notifications.Event{
		UserID:   req.UserID,
		AdminIDs: h.adminIDs(ctx),
		Title:    req.Title,
		Data: map[string]any{
			"requestId": req.ID,
			"status":    string(req.Status),
		},
	})
}

func (h *RequestHandlers) notifyApproved(ctx context.Context, req requests.Request) {
	if h.Notify == nil {
		return
	}
	_ = h.Notify.HandleRequestApproved(ctx, notifications.Event{
		UserID: req.UserID,
		Title:  req.Title,
		Data:   map[string]any{"requestId": req.ID},
	})
}

func (h *RequestHandlers) notifyRejected(ctx context.Context, req requests.Request) {
	if h.Notify == nil {
		return
	}
	_ = h.Notify.HandleRequestRejected(ctx, notifications.Event{
		UserID: req.UserID,
		Title:  req.Title,
		Detail: req.AdminNote,
		Data:   map[string]any{"requestId": req.ID},
	})
}

func (h *RequestHandlers) notifyCompleted(ctx context.Context, req requests.Request) {
	if h.Notify == nil {
		return
	}
	_ = h.Notify.HandleRequestCompleted(ctx, notifications.Event{
		UserID: req.UserID,
		Title:  req.Title,
		Data:   map[string]any{"requestId": req.ID},
	})
}

func (h *RequestHandlers) writeCollisionError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, requests.ErrCollisionInLibrary):
		writeError(w, r, platformerrors.New(http.StatusConflict, platformerrors.CodeConflict, "requests.collision.in_library"))
	case errors.Is(err, requests.ErrCollisionDuplicatePending):
		writeError(w, r, platformerrors.New(http.StatusConflict, platformerrors.CodeConflict, "requests.collision.duplicate_pending"))
	default:
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "requests.create.failed"))
	}
}

func canAccessRequest(ac auth.AuthContext, req db.Request) bool {
	if req.UserID == ac.Claims.UserID {
		return true
	}
	return auth.HasPermission(ac, auth.PermRequestsManage)
}

func policyResponse(row db.RequestPolicy) map[string]any {
	var roles []string
	if row.AutoApproveRoles != "" {
		_ = json.Unmarshal([]byte(row.AutoApproveRoles), &roles)
	}
	var admin map[string]any
	if row.AdminSettings != "" {
		_ = json.Unmarshal([]byte(row.AdminSettings), &admin)
	}
	if admin == nil {
		admin = map[string]any{}
	}
	return map[string]any{
		"autoApproveRoles":  roles,
		"userQuotaPerMonth": row.UserQuotaPerMonth,
		"adminSettings":     admin,
	}
}
