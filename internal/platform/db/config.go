package db

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// envDataDir is the environment variable that overrides the data directory.
// It is intentionally kept here (not in a higher-level config package) so the
// db package can be bootstrapped independently before the rest of the binary
// is wired together.
const envDataDir = "SOMRA_DATA_DIR"

// defaultDataDir is the relative fallback location for the SQLite database
// file when no environment override is provided. It matches the volume layout
// documented in .plan/sprint-01-foundation/05-devops-tasks.md.
const defaultDataDir = "./data"

// defaultDBFile is the file name of the primary SQLite database.
const defaultDBFile = "somra.db"

// Config describes how the SQLite database should be located on disk.
//
// Both fields are intentionally exposed so higher layers can override them
// from their own configuration sources (env, CLI flags, integration tests).
type Config struct {
	// DataDir is the directory that holds the SQLite database file and any
	// auxiliary WAL/SHM files. The directory is created on demand by
	// EnsureDataDir.
	DataDir string

	// DBFile is the file name (not path) of the SQLite database stored inside
	// DataDir. Use DSN to obtain the connection string.
	DBFile string
}

// Default returns a Config rooted at ./data/somra.db, suitable for local
// development when no environment overrides are present.
func Default() Config {
	return Config{
		DataDir: defaultDataDir,
		DBFile:  defaultDBFile,
	}
}

// FromEnv returns a Config that honours the SOMRA_DATA_DIR environment
// variable. Missing or empty values fall back to Default.
func FromEnv() Config {
	cfg := Default()
	if v := strings.TrimSpace(os.Getenv(envDataDir)); v != "" {
		cfg.DataDir = v
	}
	return cfg
}

// Validate checks that the configuration is internally consistent. It does
// not touch the file system; use EnsureDataDir for that.
func (c Config) Validate() error {
	if strings.TrimSpace(c.DataDir) == "" {
		return fmt.Errorf("db config: data dir must not be empty")
	}
	if strings.TrimSpace(c.DBFile) == "" {
		return fmt.Errorf("db config: db file must not be empty")
	}
	if strings.ContainsAny(c.DBFile, string(os.PathSeparator)+"/\\") {
		return fmt.Errorf("db config: db file %q must not contain path separators", c.DBFile)
	}
	return nil
}

// Path returns the absolute path of the SQLite database file. The path is
// not guaranteed to exist; callers should run EnsureDataDir first.
func (c Config) Path() (string, error) {
	if err := c.Validate(); err != nil {
		return "", err
	}
	abs, err := filepath.Abs(filepath.Join(c.DataDir, c.DBFile))
	if err != nil {
		return "", fmt.Errorf("db config: resolve path: %w", err)
	}
	return abs, nil
}

// DSN returns the connection string passed to the modernc.org/sqlite driver.
// PRAGMAs are intentionally not encoded in the DSN because we apply them
// explicitly in Open so misconfiguration surfaces as a clear error.
func (c Config) DSN() (string, error) {
	p, err := c.Path()
	if err != nil {
		return "", err
	}
	return "file:" + p, nil
}

// EnsureDataDir creates the data directory if it does not already exist.
// It returns an error if the path exists but is not a directory, or if the
// directory cannot be created.
func (c Config) EnsureDataDir() error {
	if err := c.Validate(); err != nil {
		return err
	}
	info, err := os.Stat(c.DataDir)
	switch {
	case err == nil:
		if !info.IsDir() {
			return fmt.Errorf("db config: %q exists and is not a directory", c.DataDir)
		}
		return nil
	case os.IsNotExist(err):
		if mkErr := os.MkdirAll(c.DataDir, 0o755); mkErr != nil {
			return fmt.Errorf("db config: create data dir %q: %w", c.DataDir, mkErr)
		}
		return nil
	default:
		return fmt.Errorf("db config: stat data dir %q: %w", c.DataDir, err)
	}
}
