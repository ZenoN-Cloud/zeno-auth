package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserService_ValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"Valid email", "test@example.com", false},
		{"Valid email with subdomain", "user@mail.example.com", false},
		{"Invalid email - no @", "testexample.com", true},
		{"Invalid email - no domain", "test@", true},
		{"Invalid email - empty", "", true},
		{"Invalid email - spaces", "test @example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEmailFormat(tt.email)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func validateEmailFormat(email string) error {
	if email == "" {
		return assert.AnError
	}
	if len(email) < 3 || !contains(email, "@") || !contains(email, ".") {
		return assert.AnError
	}
	if contains(email, " ") {
		return assert.AnError
	}
	return nil
}

func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestUserService_Context(t *testing.T) {
	t.Run("Context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		select {
		case <-ctx.Done():
			assert.Error(t, ctx.Err())
		default:
			t.Fatal("Context should be cancelled")
		}
	})
}
