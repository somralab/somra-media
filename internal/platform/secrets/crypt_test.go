package secrets_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/platform/secrets"
)

func TestDeriveKeyDeterministic(t *testing.T) {
	t.Parallel()
	a := secrets.DeriveKey("jwt-secret")
	b := secrets.DeriveKey("jwt-secret")
	assert.Equal(t, a, b)
	assert.Len(t, a, 32)
	assert.Nil(t, secrets.DeriveKey(""))
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	t.Parallel()
	key := secrets.DeriveKey("test-key-material")
	in := map[string]string{"apiKey": "super-secret", "password": "p@ss"}
	enc, err := secrets.EncryptMap(key, in)
	require.NoError(t, err)
	require.NotEmpty(t, enc)

	out, err := secrets.DecryptMap(key, enc)
	require.NoError(t, err)
	assert.Equal(t, in, out)
}

func TestEncryptEmptyMap(t *testing.T) {
	t.Parallel()
	key := secrets.DeriveKey("test")
	enc, err := secrets.EncryptMap(key, map[string]string{})
	require.NoError(t, err)
	assert.Empty(t, enc)

	out, err := secrets.DecryptMap(key, "")
	require.NoError(t, err)
	assert.Empty(t, out)
}

func TestDecryptWrongKey(t *testing.T) {
	t.Parallel()
	key := secrets.DeriveKey("key-a")
	enc, err := secrets.EncryptMap(key, map[string]string{"token": "x"})
	require.NoError(t, err)

	_, err = secrets.DecryptMap(secrets.DeriveKey("key-b"), enc)
	require.Error(t, err)
}

func TestEncryptRequiresKey(t *testing.T) {
	t.Parallel()
	_, err := secrets.EncryptMap(nil, map[string]string{"a": "b"})
	require.Error(t, err)
}

func TestDecryptInvalidBlob(t *testing.T) {
	t.Parallel()
	key := secrets.DeriveKey("k")
	_, err := secrets.DecryptMap(key, "not-valid-base64!!!")
	require.Error(t, err)
}
