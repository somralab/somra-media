package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/platform/db"
)

func TestNotificationHandlers_PreferencesAndChannels(t *testing.T) {
	h, _, token, _ := newRequestTestRouter(t)

	req := authRequest(http.MethodGet, "/api/v1/notifications/preferences", token, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var prefs map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &prefs))
	prefItems, _ := prefs["preferences"].([]any)
	assert.NotNil(t, prefItems)

	chBody, _ := json.Marshal(map[string]any{
		"channelType": "webhook",
		"name":        "Test Hook",
		"config":      map[string]any{"url": "https://example.com/hook"},
		"enabled":     true,
	})
	req = authRequest(http.MethodPost, "/api/v1/notifications/channels", token, chBody)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var created map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))
	chID := int64(created["id"].(float64))

	req = authRequest(http.MethodGet, "/api/v1/notifications/channels", token, nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	patchBody, _ := json.Marshal(map[string]any{
		"preferences": []map[string]any{{
			"eventType":       "request.created",
			"channelId":       chID,
			"enabled":         true,
			"debounceSeconds": 10,
		}},
	})
	req = authRequest(http.MethodPatch, "/api/v1/notifications/preferences", token, patchBody)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var updated map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &updated))
	items := updated["preferences"].([]any)
	require.Len(t, items, 1)
}

func TestNotificationHandlers_CreateInvalidChannel(t *testing.T) {
	h, _, token, _ := newRequestTestRouter(t)
	req := authRequest(http.MethodPost, "/api/v1/notifications/channels", token, []byte(`{"channelType":"webhook"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestNotificationHandlers_InvalidPreferencesPatch(t *testing.T) {
	h, _, token, _ := newRequestTestRouter(t)

	req := authRequest(http.MethodPatch, "/api/v1/notifications/preferences", token, []byte(`{}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)

	patchBody, _ := json.Marshal(map[string]any{
		"preferences": []map[string]any{{
			"eventType":       "request.created",
			"channelId":       99999,
			"debounceSeconds": -1,
		}},
	})
	req = authRequest(http.MethodPatch, "/api/v1/notifications/preferences", token, patchBody)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestNotificationHandlers_TestChannelWebhook(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	h, _, token, _ := newRequestTestRouter(t)
	chBody, _ := json.Marshal(map[string]any{
		"channelType": "webhook",
		"name":        "Local Hook",
		"config":      map[string]any{"url": srv.URL},
		"enabled":     true,
	})
	req := authRequest(http.MethodPost, "/api/v1/notifications/channels", token, chBody)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var created map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))
	chID := int64(created["id"].(float64))

	testBody, _ := json.Marshal(map[string]string{"message": "ping"})
	req = authRequest(http.MethodPost, fmt.Sprintf("/api/v1/notifications/channels/%d/test", chID), token, testBody)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNoContent, rec.Code)
}

func TestNotificationHandlers_ChannelTestNotFound(t *testing.T) {
	h, _, token, _ := newRequestTestRouter(t)
	req := authRequest(http.MethodPost, "/api/v1/notifications/channels/99999/test", token, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestNotificationChannelResponseShape(t *testing.T) {
	row := db.NotificationChannel{
		ID: 1, ChannelType: db.NotificationChannelWebhook,
		Name: "x", Config: `{"url":"https://example.com"}`, Enabled: true,
	}
	resp := channelResponse(row)
	cfg := resp["config"].(map[string]any)
	assert.Equal(t, "https://example.com", cfg["url"])
}
