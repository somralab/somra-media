package plugin

import (
	"encoding/json"
	"strings"
)

// SecretFieldProvider optionally declares implementation-specific secret keys.
type SecretFieldProvider interface {
	SecretFields() []string
}

var defaultSecretKeys = map[string]struct{}{
	"apikey":   {},
	"password": {},
	"token":    {},
	"secret":   {},
}

// IsSecretKey reports whether key should be stored encrypted and redacted from API responses.
func IsSecretKey(key string) bool {
	lower := strings.ToLower(key)
	if _, ok := defaultSecretKeys[lower]; ok {
		return true
	}
	for suffix := range defaultSecretKeys {
		if strings.HasSuffix(lower, suffix) && len(key) > len(suffix) {
			return true
		}
	}
	return false
}

func extraSecretKeys(extra []string) map[string]struct{} {
	if len(extra) == 0 {
		return nil
	}
	out := make(map[string]struct{}, len(extra))
	for _, k := range extra {
		out[strings.ToLower(k)] = struct{}{}
	}
	return out
}

func isSecretKeyWithExtra(key string, extra map[string]struct{}) bool {
	if IsSecretKey(key) {
		return true
	}
	if extra != nil {
		_, ok := extra[strings.ToLower(key)]
		return ok
	}
	return false
}

// SplitConfig separates public fields from secret values in a config object.
func SplitConfig(raw json.RawMessage, extraFields []string) (json.RawMessage, map[string]string, error) {
	if len(raw) == 0 {
		return json.RawMessage("{}"), map[string]string{}, nil
	}
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, nil, err
	}
	extra := extraSecretKeys(extraFields)
	public := make(map[string]any, len(obj))
	secrets := make(map[string]string)
	for k, v := range obj {
		if !isSecretKeyWithExtra(k, extra) {
			public[k] = v
			continue
		}
		if s, ok := v.(string); ok && s != "" {
			secrets[k] = s
		}
	}
	pubBytes, err := json.Marshal(public)
	if err != nil {
		return nil, nil, err
	}
	return pubBytes, secrets, nil
}

// MergeConfig merges public config with decrypted secrets for plugin activation.
func MergeConfig(public json.RawMessage, secrets map[string]string) (json.RawMessage, error) {
	obj := map[string]any{}
	if len(public) > 0 {
		if err := json.Unmarshal(public, &obj); err != nil {
			return nil, err
		}
	}
	for k, v := range secrets {
		obj[k] = v
	}
	return json.Marshal(obj)
}

// RedactConfig returns an API-safe config map with secret values replaced by *Set booleans.
func RedactConfig(public json.RawMessage, secrets map[string]string) (map[string]any, error) {
	obj := map[string]any{}
	if len(public) > 0 {
		if err := json.Unmarshal(public, &obj); err != nil {
			return nil, err
		}
	}
	for k := range secrets {
		delete(obj, k)
		obj[secretSetKey(k)] = true
	}
	return obj, nil
}

// MergeConfigPatch applies a patch to existing config, preserving secrets when patch omits or clears them.
func MergeConfigPatch(existingPublic json.RawMessage, existingSecrets map[string]string, patch json.RawMessage, extraFields []string) (json.RawMessage, map[string]string, error) {
	pubObj := map[string]any{}
	if len(existingPublic) > 0 {
		if err := json.Unmarshal(existingPublic, &pubObj); err != nil {
			return nil, nil, err
		}
	}
	secrets := make(map[string]string, len(existingSecrets))
	for k, v := range existingSecrets {
		secrets[k] = v
	}
	if len(patch) == 0 {
		pubBytes, err := json.Marshal(pubObj)
		if err != nil {
			return nil, nil, err
		}
		return pubBytes, secrets, nil
	}
	var patchObj map[string]any
	if err := json.Unmarshal(patch, &patchObj); err != nil {
		return nil, nil, err
	}
	extra := extraSecretKeys(extraFields)
	for k, v := range patchObj {
		if isSecretKeyWithExtra(k, extra) {
			s, ok := v.(string)
			if !ok {
				continue
			}
			if s == "" {
				continue
			}
			secrets[k] = s
			delete(pubObj, k)
			continue
		}
		pubObj[k] = v
	}
	pubBytes, err := json.Marshal(pubObj)
	if err != nil {
		return nil, nil, err
	}
	return pubBytes, secrets, nil
}

func secretSetKey(field string) string {
	if field == "" {
		return "secretSet"
	}
	return field + "Set"
}
