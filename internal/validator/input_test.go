package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInputValidator_ValidateEmail(t *testing.T) {
	validator := NewInputValidator()

	tests := []struct {
		name    string
		email   string
		wantErr error
	}{
		{
			name:    "valid email",
			email:   "user@example.com",
			wantErr: nil,
		},
		{
			name:    "valid email with subdomain",
			email:   "user@mail.example.com",
			wantErr: nil,
		},
		{
			name:    "invalid email - no @",
			email:   "userexample.com",
			wantErr: ErrInvalidEmail,
		},
		{
			name:    "invalid email - no domain",
			email:   "user@",
			wantErr: ErrInvalidEmail,
		},
		{
			name:    "invalid email - no TLD",
			email:   "user@example",
			wantErr: ErrInvalidEmail,
		},
		{
			name:    "empty email",
			email:   "",
			wantErr: ErrInvalidEmail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateEmail(tt.email)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInputValidator_ValidateName(t *testing.T) {
	validator := NewInputValidator()

	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{
			name:    "valid name",
			input:   "John Doe",
			wantErr: nil,
		},
		{
			name:    "valid name with hyphen",
			input:   "Mary-Jane",
			wantErr: nil,
		},
		{
			name:    "valid name with apostrophe",
			input:   "O'Brien",
			wantErr: nil,
		},
		{
			name:    "too long",
			input:   string(make([]byte, 101)),
			wantErr: ErrNameTooLong,
		},
		{
			name:    "invalid characters",
			input:   "John<script>",
			wantErr: ErrNameInvalidChars,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateName(tt.input)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInputValidator_SanitizeName(t *testing.T) {
	validator := NewInputValidator()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "remove HTML tags",
			input:    "John<b>Bold</b>Doe",
			expected: "JohnBoldDoe",
		},
		{
			name:     "trim whitespace",
			input:    "  John Doe  ",
			expected: "John Doe",
		},
		{
			name:     "remove control characters",
			input:    "John\x00Doe",
			expected: "JohnDoe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.SanitizeName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInputValidator_SanitizeEmail(t *testing.T) {
	validator := NewInputValidator()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "lowercase and trim",
			input:    "  User@Example.COM  ",
			expected: "user@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.SanitizeEmail(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
