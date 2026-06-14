// Command i18n-check enforces locale-bundle parity across the Somra codebase.
// It compares the set of message keys in every non-source locale against the
// source locale (en-US) for both the backend TOML bundles and the frontend
// JSON namespaces, and exits non-zero when any key is missing or extra.
//
// Usage:
//
//	i18n-check [--backend-dir DIR] [--frontend-dir DIR] [--source LOCALE]
//
// Default directories mirror the project layout described in AGENTS.md so
// the tool can be invoked without arguments from CI and the Makefile.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
)

const defaultSourceLocale = "en-US"

func main() {
	backendDir := flag.String("backend-dir", "internal/platform/i18n/locales", "directory of backend active.*.toml bundles")
	frontendDir := flag.String("frontend-dir", "web/src/i18n/locales", "directory of frontend locale subdirectories")
	source := flag.String("source", defaultSourceLocale, "source / fallback locale that defines the canonical key set")
	flag.Parse()

	ok := true

	if err := checkBackend(*backendDir, *source); err != nil {
		fmt.Fprintln(os.Stderr, err)
		ok = false
	}

	if err := checkFrontend(*frontendDir, *source); err != nil {
		fmt.Fprintln(os.Stderr, err)
		ok = false
	}

	if !ok {
		os.Exit(1)
	}
	fmt.Println("i18n-check: ok")
}

// checkBackend validates backend `active.<locale>.toml` files. Each file
// must define exactly the same set of message ids as the source locale.
func checkBackend(dir, source string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			fmt.Fprintf(os.Stderr, "i18n-check: backend dir %q missing, skipping\n", dir)
			return nil
		}
		return fmt.Errorf("read backend dir %q: %w", dir, err)
	}

	locales := map[string]map[string]struct{}{}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, "active.") || !strings.HasSuffix(name, ".toml") {
			continue
		}
		locale := strings.TrimSuffix(strings.TrimPrefix(name, "active."), ".toml")
		keys, err := loadTOMLKeys(filepath.Join(dir, name))
		if err != nil {
			return fmt.Errorf("backend %s: %w", locale, err)
		}
		locales[locale] = keys
	}

	if len(locales) == 0 {
		return nil
	}
	srcKeys, ok := locales[source]
	if !ok {
		return fmt.Errorf("backend: source locale %q missing in %s", source, dir)
	}

	var problems []string
	for _, locale := range sortedKeys(locales) {
		if locale == source {
			continue
		}
		if diff := compareKeys(srcKeys, locales[locale]); diff != "" {
			problems = append(problems, fmt.Sprintf("backend [%s] vs [%s]:\n%s", locale, source, diff))
		}
	}
	if len(problems) > 0 {
		return fmt.Errorf("backend locale parity failed:\n%s", strings.Join(problems, "\n"))
	}
	return nil
}

// checkFrontend validates frontend locale directories. Every locale folder
// must contain the same set of namespace files as the source locale and
// every namespace must expose the same flattened key set.
func checkFrontend(dir, source string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			fmt.Fprintf(os.Stderr, "i18n-check: frontend dir %q missing, skipping\n", dir)
			return nil
		}
		return fmt.Errorf("read frontend dir %q: %w", dir, err)
	}

	locales := map[string]map[string]struct{}{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		keys, err := loadJSONLocale(filepath.Join(dir, e.Name()))
		if err != nil {
			return fmt.Errorf("frontend %s: %w", e.Name(), err)
		}
		locales[e.Name()] = keys
	}

	if len(locales) == 0 {
		return nil
	}
	srcKeys, ok := locales[source]
	if !ok {
		return fmt.Errorf("frontend: source locale %q missing in %s", source, dir)
	}

	var problems []string
	for _, locale := range sortedKeys(locales) {
		if locale == source {
			continue
		}
		if diff := compareKeys(srcKeys, locales[locale]); diff != "" {
			problems = append(problems, fmt.Sprintf("frontend [%s] vs [%s]:\n%s", locale, source, diff))
		}
	}
	if len(problems) > 0 {
		return fmt.Errorf("frontend locale parity failed:\n%s", strings.Join(problems, "\n"))
	}
	return nil
}

// loadTOMLKeys returns the set of top-level message ids declared in a backend
// bundle. go-i18n v2 uses one table per message id; nested tables (e.g.
// pluralization "other") are intentionally ignored — we only care about
// which messages a locale defines.
func loadTOMLKeys(path string) (map[string]struct{}, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc map[string]any
	if err := toml.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	out := make(map[string]struct{}, len(doc))
	for k := range doc {
		out[k] = struct{}{}
	}
	return out, nil
}

// loadJSONLocale walks a frontend locale directory and returns a flat key
// set of `<namespace>.<dotted.path>` for every leaf value across every
// `*.json` namespace file in that locale.
func loadJSONLocale(dir string) (map[string]struct{}, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	keys := map[string]struct{}{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		ns := strings.TrimSuffix(e.Name(), ".json")
		raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		var doc map[string]any
		if err := json.Unmarshal(raw, &doc); err != nil {
			return nil, fmt.Errorf("parse %s/%s: %w", dir, e.Name(), err)
		}
		flatten(doc, ns, keys)
	}
	return keys, nil
}

func flatten(in any, prefix string, out map[string]struct{}) {
	switch v := in.(type) {
	case map[string]any:
		if len(v) == 0 {
			out[prefix] = struct{}{}
			return
		}
		for k, child := range v {
			next := prefix
			if next == "" {
				next = k
			} else {
				next = next + "." + k
			}
			flatten(child, next, out)
		}
	default:
		out[prefix] = struct{}{}
	}
}

// compareKeys returns a human-readable diff (missing + extra keys) or "" if
// the two sets are equal.
func compareKeys(src, other map[string]struct{}) string {
	var missing, extra []string
	for k := range src {
		if _, ok := other[k]; !ok {
			missing = append(missing, k)
		}
	}
	for k := range other {
		if _, ok := src[k]; !ok {
			extra = append(extra, k)
		}
	}
	if len(missing) == 0 && len(extra) == 0 {
		return ""
	}
	sort.Strings(missing)
	sort.Strings(extra)
	var b strings.Builder
	if len(missing) > 0 {
		b.WriteString("  missing keys (defined in source, absent here):\n")
		for _, k := range missing {
			fmt.Fprintf(&b, "    - %s\n", k)
		}
	}
	if len(extra) > 0 {
		b.WriteString("  extra keys (defined here, absent in source):\n")
		for _, k := range extra {
			fmt.Fprintf(&b, "    + %s\n", k)
		}
	}
	return b.String()
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
