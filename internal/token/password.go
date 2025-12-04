package token

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

func sanitizeLog(s string) string {
	replacer := strings.NewReplacer(
		"\n", " ",
		"\r", " ",
		"\t", " ",
		"\f", " ",
		"\v", " ",
		"\b", " ",
		"\a", " ",
		"\x00", " ",
	)
	return replacer.Replace(s)
}

type PasswordManager struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

func NewPasswordManager() *PasswordManager {
	return &PasswordManager{
		memory:      64 * 1024,
		iterations:  3,
		parallelism: 2,
		saltLength:  16,
		keyLength:   32,
	}
}

func (p *PasswordManager) Hash(ctx context.Context, password string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	salt, err := p.generateSalt()
	if err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, p.iterations, p.memory, p.parallelism, p.keyLength)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, p.memory, p.iterations, p.parallelism, b64Salt, b64Hash), nil
}

func (p *PasswordManager) Verify(ctx context.Context, password, encodedHash string) (bool, error) {
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}

	salt, hash, err := p.decodeHash(encodedHash)
	if err != nil {
		return false, err
	}

	otherHash := argon2.IDKey([]byte(password), salt, p.iterations, p.memory, p.parallelism, p.keyLength)

	return subtle.ConstantTimeCompare(hash, otherHash) == 1, nil
}

func (p *PasswordManager) generateSalt() ([]byte, error) {
	if p.saltLength == 0 {
		return nil, fmt.Errorf("salt length cannot be zero")
	}

	salt := make([]byte, p.saltLength)
	n, err := rand.Read(salt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random salt: %w", err)
	}
	if n != int(p.saltLength) {
		return nil, fmt.Errorf("insufficient random bytes generated: got %d, expected %d", n, p.saltLength)
	}

	return salt, nil
}

func (p *PasswordManager) decodeHash(encodedHash string) (salt, hash []byte, err error) {
	// Input validation
	if encodedHash == "" {
		return nil, nil, fmt.Errorf("empty hash string")
	}

	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" {
		return nil, nil, fmt.Errorf("invalid hash format: expected argon2id format")
	}

	// Validate parts are not empty
	for i := 2; i < 6; i++ {
		if parts[i] == "" {
			return nil, nil, fmt.Errorf("invalid hash format: empty part at index %d", i)
		}
	}

	// parts[2] is version (v=19)
	// parts[3] is parameters (m=65536,t=3,p=2)
	// parts[4] is salt
	// parts[5] is hash

	salt, err = base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode salt: %s", sanitizeLog(err.Error()))
	}

	// Validate salt length
	if len(salt) == 0 {
		return nil, nil, fmt.Errorf("decoded salt is empty")
	}

	hash, err = base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode hash: %s", sanitizeLog(err.Error()))
	}

	// Validate hash length
	if len(hash) == 0 {
		return nil, nil, fmt.Errorf("decoded hash is empty")
	}

	return salt, hash, nil
}
