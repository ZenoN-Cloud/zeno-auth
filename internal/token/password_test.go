package token

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasswordManager_Hash(t *testing.T) {
	pm := NewPasswordManager()
	ctx := context.Background()

	password := "testpassword123"
	hash, err := pm.Hash(ctx, password)

	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.Contains(t, hash, "$argon2id$")
}

func TestPasswordManager_Verify(t *testing.T) {
	pm := NewPasswordManager()
	ctx := context.Background()

	password := "testpassword123"
	hash, err := pm.Hash(ctx, password)
	require.NoError(t, err)

	// Valid password
	valid, err := pm.Verify(ctx, password, hash)
	require.NoError(t, err)
	assert.True(t, valid)

	// Invalid password
	valid, err = pm.Verify(ctx, "wrongpassword", hash)
	require.NoError(t, err)
	assert.False(t, valid)
}
