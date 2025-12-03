package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPasswordValidator_Validate(t *testing.T) {
	validator := NewPasswordValidator()

	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{
			name:     "valid password",
			password: "SecurePass123456", // 12+ chars for EU compliance
			wantErr:  nil,
		},
		{
			name:     "too short",
			password: "Short1",
			wantErr:  ErrPasswordTooShort,
		},
		{
			name:     "no uppercase",
			password: "securepass123456",
			wantErr:  ErrPasswordNoUppercase,
		},
		{
			name:     "no lowercase",
			password: "SECUREPASS123456",
			wantErr:  ErrPasswordNoLowercase,
		},
		{
			name:     "no digit",
			password: "SecurePasswordLong",
			wantErr:  ErrPasswordNoDigit,
		},
		{
			name:     "common password",
			password: "password",
			wantErr:  ErrPasswordTooShort, // Too short first
		},
		{
			name:     "common password with number",
			password: "Password123456", // Long enough but common pattern
			wantErr:  nil,              // Not in common list
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.password)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
