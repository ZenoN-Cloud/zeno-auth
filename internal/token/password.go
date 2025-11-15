package token

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/argon2"
)

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
	salt := make([]byte, p.saltLength)
	_, err := rand.Read(salt)
	return salt, err
}

func (p *PasswordManager) decodeHash(encodedHash string) (salt, hash []byte, err error) {
	var version int
	var memory, iterations uint32
	var parallelism uint8
	var saltStr, hashStr string

	_, err = fmt.Sscanf(encodedHash, "$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", &version, &memory, &iterations, &parallelism, &saltStr, &hashStr)
	if err != nil {
		return nil, nil, err
	}

	salt, err = base64.RawStdEncoding.DecodeString(saltStr)
	if err != nil {
		return nil, nil, err
	}

	hash, err = base64.RawStdEncoding.DecodeString(hashStr)
	if err != nil {
		return nil, nil, err
	}

	return salt, hash, nil
}