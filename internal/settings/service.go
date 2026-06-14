package settings

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/somralab/somra-media/internal/platform/db"
)

// Service manages persisted settings with validation.
type Service struct {
	repo *db.SettingsRepo
}

// NewService returns a settings service bound to repo.
func NewService(repo *db.SettingsRepo) *Service {
	return &Service{repo: repo}
}

// Snapshot is the grouped settings response.
type Snapshot map[string]map[string]any

// GetAll returns all settings grouped by category.
func (s *Service) GetAll(ctx context.Context) (Snapshot, error) {
	general, err := s.getGeneral(ctx)
	if err != nil {
		return nil, err
	}
	library, err := s.getLibrary(ctx)
	if err != nil {
		return nil, err
	}
	playback, err := s.getPlayback(ctx)
	if err != nil {
		return nil, err
	}
	subtitles, err := s.getSubtitles(ctx)
	if err != nil {
		return nil, err
	}
	streaming, err := s.getStreaming(ctx)
	if err != nil {
		return nil, err
	}
	return Snapshot{
		CategoryGeneral:   general,
		CategoryLibrary:   library,
		CategoryPlayback:  playback,
		CategorySubtitles: subtitles,
		CategoryStreaming: streaming,
	}, nil
}

// PatchCategory validates and persists updates for a single category.
func (s *Service) PatchCategory(ctx context.Context, category string, patch map[string]any) (map[string]any, error) {
	switch category {
	case CategoryGeneral:
		return s.patchGeneral(ctx, patch)
	case CategoryLibrary:
		return s.patchLibrary(ctx, patch)
	case CategoryPlayback:
		return s.patchPlayback(ctx, patch)
	case CategorySubtitles:
		return s.patchSubtitles(ctx, patch)
	case CategoryStreaming:
		return s.patchStreaming(ctx, patch)
	default:
		return nil, fmt.Errorf("settings: unknown category %q", category)
	}
}

// ApplySmartDefaults writes recommended values from detection.
func (s *Service) ApplySmartDefaults(ctx context.Context, defaults SmartDefaults) error {
	if defaults.MaxConcurrentTranscodes < 1 {
		return fmt.Errorf("settings: invalid max concurrent transcodes")
	}
	if err := s.repo.Set(ctx, KeyStreamingMaxConcurrent, strconv.Itoa(defaults.MaxConcurrentTranscodes)); err != nil {
		return err
	}
	if defaults.MaxHWTranscodes > 0 {
		if err := s.repo.Set(ctx, KeyStreamingMaxHW, strconv.Itoa(defaults.MaxHWTranscodes)); err != nil {
			return err
		}
	}
	if defaults.HWMode != "" {
		if err := s.repo.Set(ctx, KeyStreamingHWMode, string(defaults.HWMode)); err != nil {
			return err
		}
	}
	if defaults.RecommendedAccelerator != "" {
		if err := s.repo.Set(ctx, KeyStreamingHWAccelerator, string(defaults.RecommendedAccelerator)); err != nil {
			return err
		}
	} else if defaults.HWMode == HWModeOff {
		if err := s.repo.Set(ctx, KeyStreamingHWAccelerator, "auto"); err != nil {
			return err
		}
	}
	if defaults.ScanCron != "" {
		if err := s.repo.Set(ctx, KeyLibraryScanCron, defaults.ScanCron); err != nil {
			return err
		}
	}
	if defaults.DefaultLocale != "" {
		if err := s.repo.Set(ctx, KeyDefaultLocale, defaults.DefaultLocale); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) getGeneral(ctx context.Context) (map[string]any, error) {
	locale, _ := s.repo.Get(ctx, KeyDefaultLocale)
	if locale == "" {
		locale = "en-US"
	}
	return map[string]any{"defaultLocale": locale}, nil
}

func (s *Service) getLibrary(ctx context.Context) (map[string]any, error) {
	cron, _ := s.repo.Get(ctx, KeyLibraryScanCron)
	if cron == "" {
		cron = defaultScanCron
	}
	return map[string]any{"scanCron": cron}, nil
}

func (s *Service) getPlayback(ctx context.Context) (map[string]any, error) {
	streaming, err := s.getStreaming(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"maxConcurrentTranscodes": streaming["maxConcurrentTranscodes"],
		"maxHWTranscodes":         streaming["maxHWTranscodes"],
		"hwMode":                  streaming["hwMode"],
		"hwAccelerator":           streaming["hwAccelerator"],
		"availableAccelerators":   streaming["availableAccelerators"],
	}, nil
}

func (s *Service) getStreaming(ctx context.Context) (map[string]any, error) {
	raw, err := s.repo.Get(ctx, KeyStreamingMaxConcurrent)
	if errors.Is(err, db.ErrSettingNotFound) || raw == "" {
		raw = "2"
	} else if err != nil {
		return nil, err
	}
	n, convErr := strconv.Atoi(raw)
	if convErr != nil {
		n = 2
	}
	hwRaw, err := s.GetString(ctx, KeyStreamingMaxHW, "2")
	if err != nil {
		return nil, err
	}
	hwN, _ := strconv.Atoi(hwRaw)
	if hwN < 1 {
		hwN = 2
	}
	mode, err := s.GetString(ctx, KeyStreamingHWMode, string(HWModeAuto))
	if err != nil {
		return nil, err
	}
	accel, err := s.GetString(ctx, KeyStreamingHWAccelerator, "auto")
	if err != nil {
		return nil, err
	}
	profile := DetectSystem(nil)
	available := make([]string, 0)
	for _, a := range profile.Accelerators {
		if a.Available {
			available = append(available, string(a.ID))
		}
	}
	return map[string]any{
		"maxConcurrentTranscodes": n,
		"maxHWTranscodes":         hwN,
		"hwMode":                  mode,
		"hwAccelerator":           accel,
		"availableAccelerators":   available,
	}, nil
}

func (s *Service) getSubtitles(ctx context.Context) (map[string]any, error) {
	autoDL, _ := s.repo.Get(ctx, KeySubtitlesAutoDownload)
	langs, _ := s.repo.Get(ctx, KeySubtitlesPreferredLangs)
	apiKey, _ := s.repo.Get(ctx, KeySubtitlesAPIKey)
	return map[string]any{
		"autoDownload":       autoDL == "true" || autoDL == "1",
		"preferredLanguages": parseLangList(langs),
		"apiKeySet":          apiKey != "",
	}, nil
}

func (s *Service) patchGeneral(ctx context.Context, patch map[string]any) (map[string]any, error) {
	if v, ok := patch["defaultLocale"].(string); ok {
		if v != "en-US" && v != "tr-TR" {
			return nil, fmt.Errorf("settings: invalid locale %q", v)
		}
		if err := s.repo.Set(ctx, KeyDefaultLocale, v); err != nil {
			return nil, err
		}
	}
	return s.getGeneral(ctx)
}

func (s *Service) patchLibrary(ctx context.Context, patch map[string]any) (map[string]any, error) {
	if v, ok := patch["scanCron"].(string); ok {
		if strings.TrimSpace(v) == "" {
			return nil, fmt.Errorf("settings: scanCron must not be empty")
		}
		if err := s.repo.Set(ctx, KeyLibraryScanCron, v); err != nil {
			return nil, err
		}
	}
	return s.getLibrary(ctx)
}

func (s *Service) patchPlayback(ctx context.Context, patch map[string]any) (map[string]any, error) {
	if v, ok := patch["maxConcurrentTranscodes"]; ok {
		n, err := toInt(v)
		if err != nil || n < 1 || n > 8 {
			return nil, fmt.Errorf("settings: invalid maxConcurrentTranscodes")
		}
		if err := s.repo.Set(ctx, KeyStreamingMaxConcurrent, strconv.Itoa(n)); err != nil {
			return nil, err
		}
	}
	if v, ok := patch["maxHWTranscodes"]; ok {
		n, err := toInt(v)
		if err != nil || n < 1 || n > 4 {
			return nil, fmt.Errorf("settings: invalid maxHWTranscodes")
		}
		if err := s.repo.Set(ctx, KeyStreamingMaxHW, strconv.Itoa(n)); err != nil {
			return nil, err
		}
	}
	if v, ok := patch["hwMode"].(string); ok {
		if v != string(HWModeAuto) && v != string(HWModeOff) && v != string(HWModeForce) {
			return nil, fmt.Errorf("settings: invalid hwMode")
		}
		if err := s.repo.Set(ctx, KeyStreamingHWMode, v); err != nil {
			return nil, err
		}
	}
	if v, ok := patch["hwAccelerator"].(string); ok {
		valid := v == "auto" || v == string(AcceleratorQSV) || v == string(AcceleratorNVENC) ||
			v == string(AcceleratorVAAPI) || v == string(AcceleratorAMF)
		if !valid {
			return nil, fmt.Errorf("settings: invalid hwAccelerator")
		}
		if err := s.repo.Set(ctx, KeyStreamingHWAccelerator, v); err != nil {
			return nil, err
		}
	}
	return s.getPlayback(ctx)
}

func (s *Service) patchStreaming(ctx context.Context, patch map[string]any) (map[string]any, error) {
	return s.patchPlayback(ctx, patch)
}

func (s *Service) patchSubtitles(ctx context.Context, patch map[string]any) (map[string]any, error) {
	if v, ok := patch["autoDownload"].(bool); ok {
		val := "false"
		if v {
			val = "true"
		}
		if err := s.repo.Set(ctx, KeySubtitlesAutoDownload, val); err != nil {
			return nil, err
		}
	}
	if v, ok := patch["preferredLanguages"]; ok {
		langs, err := normalizeLangList(v)
		if err != nil {
			return nil, err
		}
		b, _ := json.Marshal(langs)
		if err := s.repo.Set(ctx, KeySubtitlesPreferredLangs, string(b)); err != nil {
			return nil, err
		}
	}
	if v, ok := patch["apiKey"].(string); ok && v != "" {
		if err := s.repo.Set(ctx, KeySubtitlesAPIKey, v); err != nil {
			return nil, err
		}
	}
	return s.getSubtitles(ctx)
}

// GetString returns a setting value or defaultVal when missing.
func (s *Service) GetString(ctx context.Context, key, defaultVal string) (string, error) {
	v, err := s.repo.Get(ctx, key)
	if errors.Is(err, db.ErrSettingNotFound) {
		return defaultVal, nil
	}
	if err != nil {
		return "", err
	}
	if v == "" {
		return defaultVal, nil
	}
	return v, nil
}

// PreferredLanguages returns parsed subtitle language preferences.
func (s *Service) PreferredLanguages(ctx context.Context) ([]string, error) {
	raw, err := s.GetString(ctx, KeySubtitlesPreferredLangs, "[]")
	if err != nil {
		return nil, err
	}
	return parseLangList(raw), nil
}

// AutoDownloadSubtitles reports whether auto-download is enabled.
func (s *Service) AutoDownloadSubtitles(ctx context.Context) (bool, error) {
	raw, err := s.GetString(ctx, KeySubtitlesAutoDownload, "false")
	if err != nil {
		return false, err
	}
	return raw == "true" || raw == "1", nil
}

// OpenSubtitlesAPIKey returns the stored API key.
func (s *Service) OpenSubtitlesAPIKey(ctx context.Context) (string, error) {
	return s.GetString(ctx, KeySubtitlesAPIKey, "")
}

func parseLangList(raw string) []string {
	if raw == "" {
		return []string{}
	}
	var langs []string
	if err := json.Unmarshal([]byte(raw), &langs); err == nil {
		return langs
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func normalizeLangList(v any) ([]string, error) {
	switch t := v.(type) {
	case []any:
		out := make([]string, 0, len(t))
		for _, item := range t {
			s, ok := item.(string)
			if !ok || len(s) < 2 {
				return nil, fmt.Errorf("settings: invalid language code")
			}
			out = append(out, strings.ToLower(s))
		}
		return out, nil
	case []string:
		out := make([]string, 0, len(t))
		for _, s := range t {
			if len(s) < 2 {
				return nil, fmt.Errorf("settings: invalid language code")
			}
			out = append(out, strings.ToLower(s))
		}
		return out, nil
	case string:
		return parseLangList(t), nil
	default:
		return nil, fmt.Errorf("settings: invalid preferredLanguages type")
	}
}

func toInt(v any) (int, error) {
	switch t := v.(type) {
	case float64:
		return int(t), nil
	case int:
		return t, nil
	case json.Number:
		n, err := strconv.ParseInt(string(t), 10, 64)
		return int(n), err
	default:
		return 0, fmt.Errorf("not an int")
	}
}
