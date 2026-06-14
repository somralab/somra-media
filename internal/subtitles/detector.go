package subtitles

import (
	"context"
	"strings"
)

// MissingLanguages returns preferred langs not present in existing tracks.
func MissingLanguages(preferred []string, existing []string) []string {
	have := make(map[string]struct{}, len(existing))
	for _, e := range existing {
		have[strings.ToLower(e)] = struct{}{}
	}
	var missing []string
	for _, p := range preferred {
		if _, ok := have[strings.ToLower(p)]; !ok {
			missing = append(missing, p)
		}
	}
	return missing
}

// LanguageLister lists subtitle languages for a media item.
type LanguageLister interface {
	ListLanguages(ctx context.Context, mediaItemID int64) ([]string, error)
}

// Detector finds missing preferred subtitle languages.
type Detector struct {
	Preferred func(ctx context.Context) ([]string, error)
	List      LanguageLister
}

// MissingForItem returns preferred languages missing for an item.
func (d *Detector) MissingForItem(ctx context.Context, itemID int64, embedded []string) ([]string, error) {
	pref, err := d.Preferred(ctx)
	if err != nil {
		return nil, err
	}
	existing := append([]string{}, embedded...)
	if d.List != nil {
		langs, err := d.List.ListLanguages(ctx, itemID)
		if err != nil {
			return nil, err
		}
		existing = append(existing, langs...)
	}
	return MissingLanguages(pref, existing), nil
}
