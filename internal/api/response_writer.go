package api

import (
	"bufio"
	"errors"
	"net"
	"net/http"
)

// responseWriter wraps http.ResponseWriter to capture status code and byte
// count for access logging. Flusher and Hijacker are forwarded so that SSE
// and future WebSocket upgrades keep working when middleware wraps them.
type responseWriter struct {
	http.ResponseWriter
	status      int
	bytes       int
	wroteHeader bool
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, status: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}
	rw.status = code
	rw.wroteHeader = true
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.bytes += n
	return n, err
}

func (rw *responseWriter) Status() int { return rw.status }
func (rw *responseWriter) Bytes() int  { return rw.bytes }

// Flush forwards to the underlying writer when it supports flushing (used by
// the SSE handler for keep-alive heartbeats).
func (rw *responseWriter) Flush() {
	if f, ok := rw.ResponseWriter.(http.Flusher); ok {
		rw.wroteHeader = true
		f.Flush()
	}
}

// Hijack forwards to the underlying writer when it supports hijacking,
// returning an error otherwise. Required to keep WebSocket upgrades and
// long-lived stream handlers compatible with this wrapper.
func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := rw.ResponseWriter.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, errors.New("api: underlying writer does not support hijacking")
}
