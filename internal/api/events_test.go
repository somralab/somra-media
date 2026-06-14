package api

import (
	"bufio"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSSE_EmitsHelloEventAndHeartbeat(t *testing.T) {
	h := newTestHandler(t, Options{SSEHeartbeat: 25 * time.Millisecond})
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/api/v1/events/stream", nil)
	require.NoError(t, err)
	req.Header.Set("Accept", "text/event-stream")

	resp, err := srv.Client().Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "text/event-stream")
	assert.Equal(t, "no-store, no-transform", resp.Header.Get("Cache-Control"))

	reader := bufio.NewReader(resp.Body)
	helloFound := false
	heartbeatFound := false

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			line = strings.TrimRight(line, "\r\n")
			if strings.HasPrefix(line, "event: hello") {
				helloFound = true
			}
			if strings.HasPrefix(line, ": ping") {
				heartbeatFound = true
			}
			if helloFound && heartbeatFound {
				return
			}
		}
	}()

	select {
	case <-done:
	case <-time.After(1500 * time.Millisecond):
		t.Fatalf("timed out reading SSE stream (hello=%v heartbeat=%v)", helloFound, heartbeatFound)
	}

	assert.True(t, helloFound, "hello event must be emitted on connect")
	assert.True(t, heartbeatFound, "heartbeat comment must be emitted")
}

func TestSSE_RejectsWrongAccept(t *testing.T) {
	h := newTestHandler(t, Options{SSEHeartbeat: 100 * time.Millisecond})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/events/stream", nil)
	req.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotAcceptable, rec.Code)
}

func TestSSE_DefaultAcceptHeaderAllowed(t *testing.T) {
	h := newTestHandler(t, Options{SSEHeartbeat: 25 * time.Millisecond})
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 750*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/api/v1/events/stream", nil)
	require.NoError(t, err)
	req.Header.Del("Accept")

	resp, err := srv.Client().Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	require.Equal(t, http.StatusOK, resp.StatusCode)
	chunk := make([]byte, 128)
	n, _ := io.ReadAtLeast(resp.Body, chunk, 1)
	assert.Greater(t, n, 0)
}
