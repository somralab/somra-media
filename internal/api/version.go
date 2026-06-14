package api

import (
	"encoding/json"
	"net/http"
	"time"
)

// BuildInfo carries build-time identifiers. Values are populated from ldflags
// at link time; defaults keep the handler usable during local dev.
type BuildInfo struct {
	Version string
	Commit  string
	BuiltAt string
}

// VersionResponse is the JSON shape of /api/v1/version.
type VersionResponse struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	BuiltAt string `json:"builtAt"`
}

func versionHandler(info BuildInfo, now func() time.Time) http.HandlerFunc {
	if now == nil {
		now = time.Now
	}
	if info.Version == "" {
		info.Version = "0.1.0-dev"
	}
	if info.BuiltAt == "" {
		info.BuiltAt = now().UTC().Format(time.RFC3339)
	}
	resp := VersionResponse(info)
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(resp)
	}
}
