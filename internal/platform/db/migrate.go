package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/pressly/goose/v3"

	"github.com/somralab/somra-media/migrations"
)

// gooseInitOnce guards the global goose registration so multiple Open /
// MigrateUp calls in the same process do not race. Goose configures itself
// via package-level state, which is unavoidable until upstream exposes an
// instance-scoped API.
var gooseInitOnce sync.Once

// gooseMu serialises every goose entry point we call. Goose reads its
// package-level logger / dialect during migration execution, so concurrent
// callers in the same process must take this lock for the race detector to
// stay quiet. The lock is only held for the duration of the goose call so
// it does not block other database work.
var gooseMu sync.Mutex

const gooseDialect = "sqlite3"

// gooseLogger adapts slog.Logger to the goose.Logger interface so migration
// output flows through the application's structured logger rather than the
// default stderr printer. Goose only emits operational messages so we map
// them all to Info / Error.
type gooseLogger struct {
	logger *slog.Logger
}

func newGooseLogger(l *slog.Logger) goose.Logger {
	if l == nil {
		l = slog.Default()
	}
	return &gooseLogger{logger: l}
}

func (g *gooseLogger) Fatalf(format string, v ...any) {
	g.logger.Error("db migrate: " + strings.TrimSuffix(fmt.Sprintf(format, v...), "\n"))
}

func (g *gooseLogger) Printf(format string, v ...any) {
	g.logger.Info("db migrate: " + strings.TrimSuffix(fmt.Sprintf(format, v...), "\n"))
}

// configureGoose wires the embedded FS, dialect and logger into goose. The
// first caller wins for the logger because goose stores it in package-level
// state and mutating it across goroutines races with goose's own reads.
// Callers must hold gooseMu while running any goose entry point.
func configureGoose(logger *slog.Logger) error {
	var setErr error
	gooseInitOnce.Do(func() {
		goose.SetBaseFS(migrations.FS)
		if err := goose.SetDialect(gooseDialect); err != nil {
			setErr = fmt.Errorf("db migrate: set dialect: %w", err)
			return
		}
		goose.SetLogger(newGooseLogger(logger))
	})
	return setErr
}

// MigrateUp applies every pending goose migration in lexical order.
//
// The function is idempotent: running it against an up-to-date database is a
// no-op that returns nil. The provided logger receives goose's operational
// messages; pass nil to use slog.Default().
func MigrateUp(ctx context.Context, d *DB, logger *slog.Logger) error {
	if d == nil || d.sqlDB == nil {
		return fmt.Errorf("db migrate: nil database handle")
	}
	if err := configureGoose(logger); err != nil {
		return err
	}
	gooseMu.Lock()
	defer gooseMu.Unlock()
	if err := goose.UpContext(ctx, d.sqlDB, "."); err != nil {
		return fmt.Errorf("db migrate up: %w", err)
	}
	return nil
}

// MigrateStatus returns the current and target migration versions. "Current"
// is the latest version recorded in goose's bookkeeping table; "target" is
// the highest version present in the embedded FS. Callers can compare the
// two to decide whether MigrateUp is needed.
func MigrateStatus(ctx context.Context, d *DB) (current int64, target int64, err error) {
	if d == nil || d.sqlDB == nil {
		return 0, 0, fmt.Errorf("db migrate: nil database handle")
	}
	if err := configureGoose(nil); err != nil {
		return 0, 0, err
	}
	gooseMu.Lock()
	defer gooseMu.Unlock()
	current, err = goose.GetDBVersionContext(ctx, d.sqlDB)
	if err != nil {
		return 0, 0, fmt.Errorf("db migrate status: current: %w", err)
	}
	target, err = latestEmbeddedVersion()
	if err != nil {
		return 0, 0, fmt.Errorf("db migrate status: target: %w", err)
	}
	return current, target, nil
}

// IntegrityCheck runs PRAGMA integrity_check. A healthy database returns the
// literal string "ok"; anything else is a list of problems suitable for
// surfacing in diagnostics.
func IntegrityCheck(ctx context.Context, d *DB) (string, error) {
	if d == nil || d.sqlDB == nil {
		return "", fmt.Errorf("db integrity check: nil database handle")
	}
	rows, err := d.sqlDB.QueryContext(ctx, "PRAGMA integrity_check")
	if err != nil {
		return "", fmt.Errorf("db integrity check: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var lines []string
	for rows.Next() {
		var line sql.NullString
		if err := rows.Scan(&line); err != nil {
			return "", fmt.Errorf("db integrity check: scan: %w", err)
		}
		if line.Valid {
			lines = append(lines, line.String)
		}
	}
	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("db integrity check: rows: %w", err)
	}
	return strings.Join(lines, "\n"), nil
}

// latestEmbeddedVersion inspects the embedded migration files and returns
// the highest goose version present. It uses goose's own helpers so any
// future naming-rule changes upstream are reflected automatically.
func latestEmbeddedVersion() (int64, error) {
	versions, err := ListEmbeddedVersions()
	if err != nil {
		return 0, err
	}
	if len(versions) == 0 {
		return 0, fmt.Errorf("read embedded migrations: no migrations found")
	}
	return versions[len(versions)-1], nil
}

// ListEmbeddedVersions returns sorted goose version numbers from the
// embedded migration FS. Callers use this for upgrade tests that stop at
// the penultimate schema revision.
func ListEmbeddedVersions() ([]int64, error) {
	entries, err := migrations.FS.ReadDir(".")
	if err != nil {
		return nil, fmt.Errorf("read embedded migrations: %w", err)
	}
	var versions []int64
	for _, e := range entries {
		v, err := goose.NumericComponent(e.Name())
		if err != nil {
			return nil, fmt.Errorf("parse migration name %q: %w", e.Name(), err)
		}
		versions = append(versions, v)
	}
	sortVersions(versions)
	return versions, nil
}

func sortVersions(v []int64) {
	for i := 0; i < len(v); i++ {
		for j := i + 1; j < len(v); j++ {
			if v[j] < v[i] {
				v[i], v[j] = v[j], v[i]
			}
		}
	}
}

// MigrateUpTo applies embedded migrations through version inclusive. When
// version is zero, no migrations run. When version is already applied,
// goose returns without error.
func MigrateUpTo(ctx context.Context, d *DB, version int64, logger *slog.Logger) error {
	if d == nil || d.sqlDB == nil {
		return fmt.Errorf("db migrate: nil database handle")
	}
	if err := configureGoose(logger); err != nil {
		return err
	}
	gooseMu.Lock()
	defer gooseMu.Unlock()
	if err := goose.UpToContext(ctx, d.sqlDB, ".", version); err != nil {
		return fmt.Errorf("db migrate up to %d: %w", version, err)
	}
	return nil
}
