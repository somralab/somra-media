package plugin

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsSecretKey(t *testing.T) {
	t.Parallel()
	assert.True(t, IsSecretKey("apiKey"))
	assert.True(t, IsSecretKey("password"))
	assert.True(t, IsSecretKey("indexerApiKey"))
	assert.False(t, IsSecretKey("prefix"))
	assert.False(t, IsSecretKey("url"))
}

func TestSplitConfig(t *testing.T) {
	t.Parallel()
	pub, sec, err := SplitConfig(json.RawMessage(`{"prefix":"demo","apiKey":"secret-key","url":"http://x"}`), nil)
	require.NoError(t, err)
	assert.Equal(t, "secret-key", sec["apiKey"])
	assert.NotContains(t, string(pub), "secret-key")
	assert.Contains(t, string(pub), "prefix")
}

func TestRedactConfig(t *testing.T) {
	t.Parallel()
	pub, err := RedactConfig(json.RawMessage(`{"prefix":"demo"}`), map[string]string{"apiKey": "x"})
	require.NoError(t, err)
	assert.Equal(t, true, pub["apiKeySet"])
	assert.Equal(t, "demo", pub["prefix"])
	assert.NotContains(t, pub, "apiKey")
}

func TestMergeConfigPatchPreservesSecrets(t *testing.T) {
	t.Parallel()
	pub, sec, err := MergeConfigPatch(
		json.RawMessage(`{"prefix":"v1"}`),
		map[string]string{"apiKey": "keep-me"},
		json.RawMessage(`{"prefix":"v2","apiKey":""}`),
		nil,
	)
	require.NoError(t, err)
	assert.Equal(t, "keep-me", sec["apiKey"])
	assert.Contains(t, string(pub), "v2")
}

func TestMergeConfigPatchUpdatesSecret(t *testing.T) {
	t.Parallel()
	_, sec, err := MergeConfigPatch(
		json.RawMessage(`{}`),
		map[string]string{"apiKey": "old"},
		json.RawMessage(`{"apiKey":"new-key"}`),
		nil,
	)
	require.NoError(t, err)
	assert.Equal(t, "new-key", sec["apiKey"])
}

func TestMergeConfig(t *testing.T) {
	t.Parallel()
	merged, err := MergeConfig(json.RawMessage(`{"prefix":"a"}`), map[string]string{"apiKey": "k"})
	require.NoError(t, err)
	var obj map[string]any
	require.NoError(t, json.Unmarshal(merged, &obj))
	assert.Equal(t, "a", obj["prefix"])
	assert.Equal(t, "k", obj["apiKey"])
}

func TestSplitConfigWithExtraFields(t *testing.T) {
	t.Parallel()
	pub, sec, err := SplitConfig(json.RawMessage(`{"customToken":"abc","url":"http://x"}`), []string{"customToken"})
	require.NoError(t, err)
	assert.Equal(t, "abc", sec["customToken"])
	assert.Contains(t, string(pub), "http://x")
}

func TestMergeConfigPatchNonStringSecretIgnored(t *testing.T) {
	t.Parallel()
	_, sec, err := MergeConfigPatch(json.RawMessage(`{}`), map[string]string{"apiKey": "old"}, json.RawMessage(`{"apiKey":123}`), nil)
	require.NoError(t, err)
	assert.Equal(t, "old", sec["apiKey"])
}
