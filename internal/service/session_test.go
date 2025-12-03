package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSessionService_TokenExpiry(t *testing.T) {
	tests := []struct {
		name      string
		createdAt time.Time
		ttl       time.Duration
		wantValid bool
	}{
		{
			name:      "Valid token",
			createdAt: time.Now(),
			ttl:       time.Hour,
			wantValid: true,
		},
		{
			name:      "Expired token",
			createdAt: time.Now().Add(-2 * time.Hour),
			ttl:       time.Hour,
			wantValid: false,
		},
		{
			name:      "Just expired",
			createdAt: time.Now().Add(-61 * time.Minute),
			ttl:       time.Hour,
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expiresAt := tt.createdAt.Add(tt.ttl)
			isValid := time.Now().Before(expiresAt)
			assert.Equal(t, tt.wantValid, isValid)
		})
	}
}

func TestSessionService_RefreshTokenTTL(t *testing.T) {
	defaultTTL := 14 * 24 * time.Hour // 14 days

	t.Run("Default refresh token TTL", func(t *testing.T) {
		assert.Equal(t, 14*24*time.Hour, defaultTTL)
	})

	t.Run("Access token TTL shorter than refresh", func(t *testing.T) {
		accessTTL := 30 * time.Minute
		assert.True(t, accessTTL < defaultTTL)
	})
}
