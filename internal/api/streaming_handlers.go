package api

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/somralab/somra-media/internal/auth"
	"github.com/somralab/somra-media/internal/library"
	"github.com/somralab/somra-media/internal/platform/db"
	platformerrors "github.com/somralab/somra-media/internal/platform/errors"
	"github.com/somralab/somra-media/internal/streaming"
)

// StreamingHandlers serves playback and manifest endpoints.
type StreamingHandlers struct {
	Streaming *streaming.Service
	Media     *db.MediaRepo
	Library   *db.LibraryRepo
	Playback  *db.PlaybackRepo
	CacheRoot string
}

type playRequest struct {
	Capabilities        streaming.ClientCapabilities `json:"capabilities"`
	AudioStreamIndex    *int                         `json:"audioStreamIndex"`
	SubtitleStreamIndex *int                         `json:"subtitleStreamIndex"`
	StartPositionMs     int64                        `json:"startPositionMs"`
}

type playResponse struct {
	SessionID   string         `json:"sessionId"`
	Mode        streaming.Mode `json:"mode"`
	ManifestURL string         `json:"manifestUrl"`
	ExpiresAt   string         `json:"expiresAt"`
	Reason      string         `json:"reason,omitempty"`
}

type sessionSummary struct {
	SessionID   string         `json:"sessionId"`
	MediaItemID int64          `json:"mediaItemId"`
	Mode        streaming.Mode `json:"mode"`
	Status      string         `json:"status"`
	ExpiresAt   string         `json:"expiresAt"`
}

// Mount registers streaming routes (requires auth + library:read).
func (h *StreamingHandlers) Mount(r chi.Router) {
	r.With(RequirePermission(auth.PermLibraryRead)).Post("/media-items/{itemID}/play", h.play)
	r.Route("/streaming/sessions", func(r chi.Router) {
		r.With(RequirePermission(auth.PermLibraryRead)).Get("/", h.listSessions)
		r.Route("/{sessionID}", func(r chi.Router) {
			r.With(RequirePermission(auth.PermLibraryRead)).Delete("/", h.stopSession)
			r.With(RequirePermission(auth.PermLibraryRead)).Get("/master.m3u8", h.serveMaster)
			r.With(RequirePermission(auth.PermLibraryRead)).Get("/{file}", h.serveSegment)
		})
	})
}

func (h *StreamingHandlers) play(w http.ResponseWriter, r *http.Request) {
	ac, ok := auth.FromContext(r.Context())
	if !ok {
		writeError(w, r, platformerrors.New(http.StatusUnauthorized, platformerrors.CodeUnauthorized, "auth.errors.unauthorized"))
		return
	}
	itemID, err := parseID(chi.URLParam(r, "itemID"))
	if err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "streaming.play.invalid_item"))
		return
	}
	if err := h.checkParental(r, itemID); err != nil {
		writeError(w, r, err)
		return
	}

	var req playRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	resp, err := h.Streaming.StartPlay(r.Context(), streaming.PlayRequest{
		UserID: ac.Claims.UserID, MediaItemID: itemID,
		Capabilities: req.Capabilities, AudioStreamIndex: req.AudioStreamIndex,
		SubtitleStreamIndex: req.SubtitleStreamIndex, StartPositionMs: req.StartPositionMs,
	})
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusBadRequest, platformerrors.CodeValidation, "streaming.play.failed"))
		return
	}
	writeJSON(w, http.StatusOK, playResponse{
		SessionID: resp.SessionID, Mode: resp.Mode, ManifestURL: resp.ManifestURL,
		ExpiresAt: resp.ExpiresAt.UTC().Format(time.RFC3339), Reason: resp.Decision.Reason,
	})
}

func (h *StreamingHandlers) listSessions(w http.ResponseWriter, r *http.Request) {
	ac, ok := auth.FromContext(r.Context())
	if !ok {
		writeError(w, r, platformerrors.New(http.StatusUnauthorized, platformerrors.CodeUnauthorized, "auth.errors.unauthorized"))
		return
	}
	sessions, err := h.Playback.ListActiveByUser(r.Context(), ac.Claims.UserID)
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "streaming.sessions.list_failed"))
		return
	}
	out := make([]sessionSummary, 0, len(sessions))
	for _, s := range sessions {
		out = append(out, sessionSummary{
			SessionID: s.ID, MediaItemID: s.MediaItemID,
			Mode: streaming.Mode(s.Mode), Status: string(s.Status),
			ExpiresAt: s.ExpiresAt.UTC().Format(time.RFC3339),
		})
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *StreamingHandlers) stopSession(w http.ResponseWriter, r *http.Request) {
	ac, ok := auth.FromContext(r.Context())
	if !ok {
		writeError(w, r, platformerrors.New(http.StatusUnauthorized, platformerrors.CodeUnauthorized, "auth.errors.unauthorized"))
		return
	}
	sessionID := chi.URLParam(r, "sessionID")
	if err := h.Streaming.StopSession(r.Context(), sessionID, ac.Claims.UserID); err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusNotFound, platformerrors.CodeNotFound, "streaming.session.not_found"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *StreamingHandlers) serveMaster(w http.ResponseWriter, r *http.Request) {
	h.serveCachedFile(w, r, "master.m3u8", "application/vnd.apple.mpegurl")
}

func (h *StreamingHandlers) serveSegment(w http.ResponseWriter, r *http.Request) {
	file := chi.URLParam(r, "file")
	if file == "master.m3u8" {
		h.serveMaster(w, r)
		return
	}
	ct := "application/octet-stream"
	switch {
	case strings.HasSuffix(file, ".m3u8"):
		ct = "application/vnd.apple.mpegurl"
	case strings.HasSuffix(file, ".mp4"), strings.HasSuffix(file, ".m4s"):
		ct = "video/mp4"
	}
	h.serveCachedFile(w, r, file, ct)
}

func (h *StreamingHandlers) serveCachedFile(w http.ResponseWriter, r *http.Request, rel, contentType string) {
	ac, ok := auth.FromContext(r.Context())
	if !ok {
		writeError(w, r, platformerrors.New(http.StatusUnauthorized, platformerrors.CodeUnauthorized, "auth.errors.unauthorized"))
		return
	}
	sessionID := chi.URLParam(r, "sessionID")
	sess, err := h.Streaming.GetSession(r.Context(), sessionID, ac.Claims.UserID)
	if err != nil {
		writeError(w, r, platformerrors.New(http.StatusNotFound, platformerrors.CodeNotFound, "streaming.session.not_found"))
		return
	}
	_ = h.Streaming.TouchSession(r.Context(), sessionID)

	cacheRoot := h.CacheRoot
	if cacheRoot == "" {
		cacheRoot = filepath.Dir(sess.CachePath)
	}
	path, err := streaming.ValidateCacheSegmentPath(cacheRoot, sessionID, rel)
	if err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "streaming.segment.invalid_path"))
		return
	}

	if rel == "source" {
		if err := h.validateSourceAccess(r, sess.MediaFileID); err != nil {
			writeError(w, r, err)
			return
		}
	}

	if _, err := os.Stat(path); err != nil {
		if rel == "master.m3u8" {
			alt := filepath.Join(sess.CachePath, "stream.m3u8")
			if _, err2 := os.Stat(alt); err2 == nil {
				path = alt
			}
		}
	}

	f, err := os.Open(path)
	if err != nil {
		writeError(w, r, platformerrors.New(http.StatusNotFound, platformerrors.CodeNotFound, "streaming.segment.not_found"))
		return
	}
	defer func() { _ = f.Close() }()
	info, err := f.Stat()
	if err != nil {
		writeError(w, r, platformerrors.New(http.StatusNotFound, platformerrors.CodeNotFound, "streaming.segment.not_found"))
		return
	}
	w.Header().Set("Content-Type", contentType)
	http.ServeContent(w, r, info.Name(), info.ModTime(), f)
}

func (h *StreamingHandlers) validateSourceAccess(r *http.Request, fileID int64) *platformerrors.Error {
	mediaFile, err := h.Media.GetFileByID(r.Context(), fileID)
	if err != nil {
		return platformerrors.New(http.StatusNotFound, platformerrors.CodeNotFound, "streaming.source.not_found")
	}
	roots, err := h.Library.ListPaths(r.Context(), mediaFile.LibraryID)
	if err != nil {
		return platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "streaming.source.invalid")
	}
	if err := library.ValidateMediaPath(mediaFile.Path, roots); err != nil {
		return platformerrors.New(http.StatusForbidden, platformerrors.CodeForbidden, "streaming.source.forbidden")
	}
	return nil
}

func (h *StreamingHandlers) checkParental(r *http.Request, itemID int64) *platformerrors.Error {
	ac, ok := auth.FromContext(r.Context())
	if !ok {
		return nil
	}
	item, err := h.Media.GetItemByID(r.Context(), itemID, "en-US")
	if err != nil {
		return platformerrors.New(http.StatusNotFound, platformerrors.CodeNotFound, "media.notFound")
	}
	maxRating := ac.Profile.MaxContentRating
	if ac.Profile.IsChild && maxRating == nil {
		g := "PG"
		maxRating = &g
	}
	if !auth.RatingAllowed(maxRating, item.ContentRating) {
		return platformerrors.New(http.StatusNotFound, platformerrors.CodeNotFound, "media.notFound")
	}
	return nil
}
