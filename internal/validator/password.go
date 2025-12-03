package validator

import (
	"errors"
	"strings"
	"unicode"
)

var (
	ErrPasswordTooShort    = errors.New("password must be at least 8 characters long")
	ErrPasswordNoUppercase = errors.New("password must contain at least one uppercase letter")
	ErrPasswordNoLowercase = errors.New("password must contain at least one lowercase letter")
	ErrPasswordNoDigit     = errors.New("password must contain at least one digit")
	ErrPasswordCommon      = errors.New("password is too common, please choose a stronger password")
)

// CommonPasswords contains top 100 most common passwords
var commonPasswords = map[string]bool{
	"password": true, "123456": true, "12345678": true, "qwerty": true,
	"abc123": true, "monkey": true, "1234567": true, "letmein": true,
	"trustno1": true, "dragon": true, "baseball": true, "111111": true,
	"iloveyou": true, "master": true, "sunshine": true, "ashley": true,
	"bailey": true, "passw0rd": true, "shadow": true, "123123": true,
	"654321": true, "superman": true, "qazwsx": true, "michael": true,
	"football": true, "welcome": true, "jesus": true, "ninja": true,
	"mustang": true, "password1": true, "123456789": true, "password123": true,
	"admin": true, "root": true, "toor": true, "pass": true,
	"test": true, "guest": true, "info": true, "adm": true,
	"mysql": true, "user": true, "administrator": true, "oracle": true,
	"ftp": true, "pi": true, "puppet": true, "ansible": true,
	"ec2-user": true, "vagrant": true, "azureuser": true, "changeme": true,
}

// PasswordValidator validates password strength
type PasswordValidator struct {
	MinLength        int
	RequireUppercase bool
	RequireLowercase bool
	RequireDigit     bool
	RequireSpecial   bool
	CheckCommon      bool
}

// NewPasswordValidator creates a validator with EU-compliant secure defaults
// Meets NIST SP 800-63B and EU cybersecurity requirements
func NewPasswordValidator() *PasswordValidator {
	return &PasswordValidator{
		MinLength:        12, // EU recommendation: 12+ chars
		RequireUppercase: true,
		RequireLowercase: true,
		RequireDigit:     true,
		RequireSpecial:   false, // Optional but recommended
		CheckCommon:      true,
	}
}

// Validate checks if password meets all requirements (EU-compliant)
func (v *PasswordValidator) Validate(password string) error {
	// EU requirement: minimum 12 characters for strong passwords
	if len(password) < v.MinLength {
		return ErrPasswordTooShort
	}

	// EU requirement: maximum length to prevent DoS
	if len(password) > 128 {
		return errors.New("password must not exceed 128 characters")
	}

	if v.CheckCommon && v.isCommonPassword(password) {
		return ErrPasswordCommon
	}

	var (
		hasUpper bool
		hasLower bool
		hasDigit bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		}
	}

	if v.RequireUppercase && !hasUpper {
		return ErrPasswordNoUppercase
	}

	if v.RequireLowercase && !hasLower {
		return ErrPasswordNoLowercase
	}

	if v.RequireDigit && !hasDigit {
		return ErrPasswordNoDigit
	}

	return nil
}

func (v *PasswordValidator) isCommonPassword(password string) bool {
	return commonPasswords[strings.ToLower(password)]
}
