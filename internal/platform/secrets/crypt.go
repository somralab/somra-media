package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
)

// DeriveKey hashes keyMaterial into a 32-byte AES-256 key.
func DeriveKey(keyMaterial string) []byte {
	if keyMaterial == "" {
		return nil
	}
	sum := sha256.Sum256([]byte(keyMaterial))
	return sum[:]
}

// EncryptMap encrypts a string map with AES-256-GCM.
// The returned string is base64(nonce || ciphertext). Empty maps yield "".
func EncryptMap(key []byte, secrets map[string]string) (string, error) {
	if len(secrets) == 0 {
		return "", nil
	}
	if len(key) == 0 {
		return "", fmt.Errorf("encrypt secrets: key is required")
	}
	plain, err := json.Marshal(secrets)
	if err != nil {
		return "", fmt.Errorf("encrypt secrets marshal: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("encrypt secrets cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("encrypt secrets gcm: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("encrypt secrets nonce: %w", err)
	}
	sealed := gcm.Seal(nonce, nonce, plain, nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// DecryptMap decrypts a blob produced by [EncryptMap].
func DecryptMap(key []byte, encoded string) (map[string]string, error) {
	if encoded == "" {
		return map[string]string{}, nil
	}
	if len(key) == 0 {
		return nil, fmt.Errorf("decrypt secrets: key is required")
	}
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("decrypt secrets decode: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("decrypt secrets cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("decrypt secrets gcm: %w", err)
	}
	nonceSize := gcm.NonceSize()
	if len(raw) < nonceSize {
		return nil, fmt.Errorf("decrypt secrets: ciphertext too short")
	}
	nonce, ciphertext := raw[:nonceSize], raw[nonceSize:]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt secrets open: %w", err)
	}
	out := map[string]string{}
	if err := json.Unmarshal(plain, &out); err != nil {
		return nil, fmt.Errorf("decrypt secrets unmarshal: %w", err)
	}
	return out, nil
}
