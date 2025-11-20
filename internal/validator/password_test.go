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
			password: "SecurePass123",
			wantErr:  nil,
		},
		{
			name:     "too short",
			password: "Short1",
			wantErr:  ErrPasswordTooShort,
		},
		{
			name:     "no uppercase",
			password: "securepass123",
			wantErr:  ErrPasswordNoUppercase,
		},
		{
			name:     "no lowercase",
			password: "SECUREPASS123",
			wantErr:  ErrPasswordNoLowercase,
		},
		{
			name:     "no digit",
			password: "SecurePassword",
			wantErr:  ErrPasswordNoDigit,
		},
		{
			name:     "common password",
			password: "password",
			wantErr:  ErrPasswordCommon,
		},
		{
			name:     "common password with number",
			password: "password123",
			wantErr:  ErrPasswordCommon,
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
