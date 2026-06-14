package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/somralab/somra-media/internal/auth"
	"github.com/somralab/somra-media/internal/notifications"
	"github.com/somralab/somra-media/internal/platform/db"
	platformerrors "github.com/somralab/somra-media/internal/platform/errors"
)

// NotificationHandlers serves notification preference and channel endpoints.
type NotificationHandlers struct {
	Channels   *db.NotificationChannelRepo
	Prefs      *db.NotificationPreferenceRepo
	Dispatcher *notifications.Dispatcher
}

// Mount registers /notifications routes.
func (h *NotificationHandlers) Mount(r chi.Router) {
	r.Route("/notifications", func(r chi.Router) {
		r.Get("/preferences", h.getPreferences)
		r.Patch("/preferences", h.patchPreferences)

		r.Route("/channels", func(r chi.Router) {
			r.With(RequirePermission(auth.PermNotificationsManage)).Get("/", h.listChannels)
			r.With(RequirePermission(auth.PermNotificationsManage)).Post("/", h.createChannel)
			r.With(RequirePermission(auth.PermNotificationsManage)).Post("/{channelId}/test", h.testChannel)
		})
	})
}

func (h *NotificationHandlers) getPreferences(w http.ResponseWriter, r *http.Request) {
	ac, ok := auth.FromContext(r.Context())
	if !ok {
		writeError(w, r, platformerrors.New(http.StatusUnauthorized, platformerrors.CodeUnauthorized, "auth.errors.unauthorized"))
		return
	}
	rows, err := h.Prefs.ListByUser(r.Context(), ac.Claims.UserID)
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "notifications.preferences.list.failed"))
		return
	}
	if rows == nil {
		rows = []db.NotificationPreference{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"preferences": rows})
}

type notificationPreferencesPatch struct {
	Preferences []notificationPreferenceInput `json:"preferences"`
}

type notificationPreferenceInput struct {
	ID              int64  `json:"id"`
	EventType       string `json:"eventType"`
	ChannelID       int64  `json:"channelId"`
	Enabled         *bool  `json:"enabled"`
	DebounceSeconds *int   `json:"debounceSeconds"`
}

func (h *NotificationHandlers) patchPreferences(w http.ResponseWriter, r *http.Request) {
	ac, ok := auth.FromContext(r.Context())
	if !ok {
		writeError(w, r, platformerrors.New(http.StatusUnauthorized, platformerrors.CodeUnauthorized, "auth.errors.unauthorized"))
		return
	}
	var body notificationPreferencesPatch
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.Preferences) == 0 {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "notifications.preferences.patch.invalid"))
		return
	}
	ctx := r.Context()
	for _, in := range body.Preferences {
		if in.EventType == "" || in.ChannelID == 0 {
			writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "notifications.preferences.patch.invalid"))
			return
		}
		if _, err := h.Channels.GetByID(ctx, in.ChannelID); err != nil {
			if errors.Is(err, db.ErrNotificationChannelNotFound) {
				writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "notifications.channel.not_found"))
				return
			}
			writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "notifications.preferences.patch.failed"))
			return
		}
		p := db.NotificationPreference{
			ID:        in.ID,
			UserID:    ac.Claims.UserID,
			EventType: in.EventType,
			ChannelID: in.ChannelID,
			Enabled:   true,
		}
		if in.Enabled != nil {
			p.Enabled = *in.Enabled
		}
		if in.DebounceSeconds != nil {
			if *in.DebounceSeconds < 0 {
				writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "notifications.preferences.patch.invalid"))
				return
			}
			p.DebounceSeconds = *in.DebounceSeconds
		}
		if _, err := h.Prefs.Upsert(ctx, p); err != nil {
			writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "notifications.preferences.patch.failed"))
			return
		}
	}
	rows, err := h.Prefs.ListByUser(ctx, ac.Claims.UserID)
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "notifications.preferences.list.failed"))
		return
	}
	if rows == nil {
		rows = []db.NotificationPreference{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"preferences": rows})
}

func (h *NotificationHandlers) listChannels(w http.ResponseWriter, r *http.Request) {
	rows, err := h.Channels.List(r.Context())
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "notifications.channels.list.failed"))
		return
	}
	out := make([]map[string]any, 0, len(rows))
	for _, ch := range rows {
		out = append(out, channelResponse(ch))
	}
	writeJSON(w, http.StatusOK, map[string]any{"channels": out})
}

type notificationChannelInput struct {
	ChannelType string         `json:"channelType"`
	Name        string         `json:"name"`
	Config      map[string]any `json:"config"`
	Enabled     *bool          `json:"enabled"`
}

func (h *NotificationHandlers) createChannel(w http.ResponseWriter, r *http.Request) {
	var in notificationChannelInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.ChannelType == "" || in.Config == nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "notifications.channels.create.invalid"))
		return
	}
	cfgBytes, err := json.Marshal(in.Config)
	if err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "notifications.channels.create.invalid"))
		return
	}
	enabled := true
	if in.Enabled != nil {
		enabled = *in.Enabled
	}
	ch := db.NotificationChannel{
		ChannelType: db.NotificationChannelType(in.ChannelType),
		Name:        in.Name,
		Config:      string(cfgBytes),
		Enabled:     enabled,
	}
	id, err := h.Channels.Create(r.Context(), ch)
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "notifications.channels.create.failed"))
		return
	}
	created, err := h.Channels.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "notifications.channels.create.failed"))
		return
	}
	if h.Dispatcher != nil && created.Enabled {
		if delivery, err := notifications.ChannelFromDB(created); err == nil {
			h.Dispatcher.RegisterChannel(delivery)
		}
	}
	writeJSON(w, http.StatusCreated, channelResponse(created))
}

type notificationTestInput struct {
	Message string `json:"message"`
}

func (h *NotificationHandlers) testChannel(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "channelId"))
	if err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "notifications.channel.id.invalid"))
		return
	}
	row, err := h.Channels.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, db.ErrNotificationChannelNotFound) {
			writeError(w, r, platformerrors.New(http.StatusNotFound, platformerrors.CodeNotFound, "notifications.channel.not_found"))
			return
		}
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "notifications.channel.test.failed"))
		return
	}
	delivery, err := notifications.ChannelFromDB(row)
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusBadRequest, platformerrors.CodeValidation, "notifications.channels.create.invalid"))
		return
	}
	var body notificationTestInput
	if r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}
	msg := body.Message
	if msg == "" {
		msg = "Somra notification channel test"
	}
	ac, _ := auth.FromContext(r.Context())
	userID := ""
	if ac.Claims.UserID != "" {
		userID = ac.Claims.UserID
	}
	if err := delivery.Send(r.Context(), notifications.Notification{
		EventType: notifications.EventSystemError,
		Subject:   "Test notification",
		Body:      msg,
		Locale:    "en-US",
		UserID:    userID,
		SentAt:    time.Now().UTC(),
	}); err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusBadGateway, platformerrors.CodeInternal, "notifications.channel.test.failed"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func channelResponse(ch db.NotificationChannel) map[string]any {
	var config map[string]any
	if ch.Config != "" {
		_ = json.Unmarshal([]byte(ch.Config), &config)
	}
	if config == nil {
		config = map[string]any{}
	}
	return map[string]any{
		"id":          ch.ID,
		"channelType": ch.ChannelType,
		"name":        ch.Name,
		"config":      config,
		"enabled":     ch.Enabled,
		"createdAt":   ch.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":   ch.UpdatedAt.UTC().Format(time.RFC3339),
	}
}
