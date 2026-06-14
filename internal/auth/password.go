package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/crypto/argon2"
)

var (
	ErrWeakPassword       = errors.New("auth: password does not meet policy")
	ErrInvalidCredentials = errors.New("auth: invalid credentials")
)

// PasswordPolicy describes minimum password requirements.
type PasswordPolicy struct {
	MinLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireDigit   bool
	RequireSpecial bool
}

// DefaultPasswordPolicy returns the Somra default policy.
func DefaultPasswordPolicy() PasswordPolicy {
	return PasswordPolicy{
		MinLength:      8,
		RequireUpper:   true,
		RequireLower:   true,
		RequireDigit:   true,
		RequireSpecial: false,
	}
}

// PasswordHasher hashes and verifies passwords with argon2id.
type PasswordHasher struct {
	policy PasswordPolicy
}

// NewPasswordHasher returns a hasher with the given policy.
func NewPasswordHasher(policy PasswordPolicy) *PasswordHasher {
	return &PasswordHasher{policy: policy}
}

const (
	argonTime    = 3
	argonMemory  = 64 * 1024
	argonThreads = 4
	argonKeyLen  = 32
	argonSaltLen = 16
)

// Hash returns an encoded argon2id hash of password.
func (h *PasswordHasher) Hash(password string) (string, error) {
	if err := h.Validate(password); err != nil {
		return "", err
	}
	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("auth hash salt: %w", err)
	}
	key := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Key := base64.RawStdEncoding.EncodeToString(key)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argonMemory, argonTime, argonThreads, b64Salt, b64Key), nil
}

// Verify compares password against an encoded hash.
func (h *PasswordHasher) Verify(password, encoded string) (bool, error) {
	salt, key, params, err := decodeArgonHash(encoded)
	if err != nil {
		return false, err
	}
	other := argon2.IDKey([]byte(password), salt, params.time, params.memory, params.threads, uint32(len(key)))
	if subtle.ConstantTimeCompare(key, other) == 1 {
		return true, nil
	}
	return false, nil
}

// Validate checks password against policy without hashing.
func (h *PasswordHasher) Validate(password string) error {
	if len(password) < h.policy.MinLength {
		return fmt.Errorf("%w: minimum length %d", ErrWeakPassword, h.policy.MinLength)
	}
	var upper, lower, digit, special bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			upper = true
		case unicode.IsLower(r):
			lower = true
		case unicode.IsDigit(r):
			digit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			special = true
		}
	}
	if h.policy.RequireUpper && !upper {
		return fmt.Errorf("%w: requires uppercase letter", ErrWeakPassword)
	}
	if h.policy.RequireLower && !lower {
		return fmt.Errorf("%w: requires lowercase letter", ErrWeakPassword)
	}
	if h.policy.RequireDigit && !digit {
		return fmt.Errorf("%w: requires digit", ErrWeakPassword)
	}
	if h.policy.RequireSpecial && !special {
		return fmt.Errorf("%w: requires special character", ErrWeakPassword)
	}
	return nil
}

type argonParams struct {
	memory  uint32
	time    uint32
	threads uint8
}

func decodeArgonHash(encoded string) (salt, key []byte, params argonParams, err error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return nil, nil, params, fmt.Errorf("auth verify: malformed hash")
	}
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return nil, nil, params, fmt.Errorf("auth verify: parse version: %w", err)
	}
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &params.memory, &params.time, &params.threads); err != nil {
		return nil, nil, params, fmt.Errorf("auth verify: parse params: %w", err)
	}
	salt, err = base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, params, fmt.Errorf("auth verify: decode salt: %w", err)
	}
	key, err = base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, params, fmt.Errorf("auth verify: decode key: %w", err)
	}
	return salt, key, params, nil
}
