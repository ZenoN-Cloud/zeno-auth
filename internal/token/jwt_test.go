package token

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateTestKeys() (string, string) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)

	privateKeyBytes, _ := x509.MarshalPKCS8PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	publicKeyBytes, _ := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return string(privateKeyPEM), string(publicKeyPEM)
}

func TestJWTManager_Generate(t *testing.T) {
	privateKeyPEM, publicKeyPEM := generateTestKeys()
	jwtManager, err := NewJWTManager(privateKeyPEM, publicKeyPEM)
	require.NoError(t, err)

	ctx := context.Background()
	userID := uuid.New()
	orgID := uuid.New()
	roles := []string{"ADMIN"}
	ttl := 1800

	token, err := jwtManager.Generate(ctx, userID, orgID, roles, ttl)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestJWTManager_Validate(t *testing.T) {
	privateKeyPEM, publicKeyPEM := generateTestKeys()
	jwtManager, err := NewJWTManager(privateKeyPEM, publicKeyPEM)
	require.NoError(t, err)

	ctx := context.Background()
	userID := uuid.New()
	orgID := uuid.New()
	roles := []string{"ADMIN"}
	ttl := 1800

	token, err := jwtManager.Generate(ctx, userID, orgID, roles, ttl)
	require.NoError(t, err)

	claims, err := jwtManager.Validate(ctx, token)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, orgID, claims.OrgID)
	assert.Equal(t, roles, claims.Roles)
}
