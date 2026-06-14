package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/auth"
	"github.com/somralab/somra-media/internal/jobs"
	"github.com/somralab/somra-media/internal/notifications"
	"github.com/somralab/somra-media/internal/platform/db"
	"github.com/somralab/somra-media/internal/platform/i18n"
)

func newDirectNotificationHandlers(t *testing.T) (*NotificationHandlers, auth.AuthContext) {
	t.Helper()
	d := openTestDB(t)
	q := d.Querier()
	bundle, err := i18n.NewBundle()
	require.NoError(t, err)
	queue := jobs.NewMemoryQueue(jobs.MemoryQueueConfig{Workers: 1, Buffer: 4})
	t.Cleanup(queue.Close)
	dispatcher := notifications.NewDispatcher(notifications.DispatcherConfig{
		Renderer: notifications.NewTemplateRenderer(bundle),
		Filter:   notifications.NewPreferenceFilter(nil),
		Queue:    queue,
	})
	ac := adminAuthContext(t, d)
	return &NotificationHandlers{
		Channels:   db.NewNotificationChannelRepo(q),
		Prefs:      db.NewNotificationPreferenceRepo(q),
		Dispatcher: dispatcher,
	}, ac
}

func serveNotificationRoute(t *testing.T, h *NotificationHandlers, method, path string, body []byte, ac *auth.AuthContext) *httptest.ResponseRecorder {
	t.Helper()
	r := chi.NewRouter()
	if ac != nil {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				next.ServeHTTP(w, req.WithContext(auth.WithAuthContext(req.Context(), *ac)))
			})
		})
	}
	h.Mount(r)
	var reader *bytes.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, reader)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestNotificationHandlers_DirectValidationPaths(t *testing.T) {
	h, ac := newDirectNotificationHandlers(t)

	rec := serveNotificationRoute(t, h, http.MethodGet, "/notifications/preferences", nil, nil)
	require.Equal(t, http.StatusUnauthorized, rec.Code)

	rec = serveNotificationRoute(t, h, http.MethodPatch, "/notifications/preferences", []byte(`{"preferences":[]}`), &ac)
	require.Equal(t, http.StatusBadRequest, rec.Code)

	rec = serveNotificationRoute(t, h, http.MethodPost, "/notifications/channels", []byte(`{`), &ac)
	require.Equal(t, http.StatusBadRequest, rec.Code)

	rec = serveNotificationRoute(t, h, http.MethodPost, "/notifications/channels/1/test", nil, &ac)
	require.Equal(t, http.StatusNotFound, rec.Code)

	rec = serveNotificationRoute(t, h, http.MethodGet, "/notifications/channels", nil, &ac)
	require.Equal(t, http.StatusOK, rec.Code)

	rec = serveNotificationRoute(t, h, http.MethodGet, "/notifications/preferences", nil, &ac)
	require.Equal(t, http.StatusOK, rec.Code)

	badPatch, _ := json.Marshal(map[string]any{
		"preferences": []map[string]any{{
			"eventType": "request.created",
			"channelId": 1,
		}},
	})
	rec = serveNotificationRoute(t, h, http.MethodPatch, "/notifications/preferences", badPatch, &ac)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}
