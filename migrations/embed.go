// Package migrations carries the SQL migration files for Somra's data layer.
//
// The files are embedded at compile time so the single Somra binary can run
// migrations without relying on the file system layout of the host. The
// embed directive lives next to the SQL files because Go's go:embed cannot
// traverse parent directories; the embedded FS is then re-exported to the
// internal/platform/db package, which is responsible for applying migrations
// at startup.
package migrations

import "embed"

// FS holds every goose SQL migration shipped with this binary.
//
//go:embed *.sql
var FS embed.FS
