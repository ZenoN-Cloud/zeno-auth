package validator

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
)

var (
	ErrInvalidEmail     = errors.New("invalid email format")
	ErrNameTooLong      = errors.New("name must be less than 100 characters")
	ErrNameInvalidChars = errors.New("name contains invalid characters")
	ErrEmailTooLong     = errors.New("email must be less than 255 characters")
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// InputValidator validates and sanitizes user input
type InputValidator struct{}

// NewInputValidator creates a new input validator
func NewInputValidator() *InputValidator {
	return &InputValidator{}
}

// ValidateEmail checks if email is valid RFC 5322 format
func (v *InputValidator) ValidateEmail(email string) error {
	email = strings.TrimSpace(email)

	if len(email) == 0 {
		return ErrInvalidEmail
	}

	if len(email) > 254 {
		return ErrEmailTooLong
	}

	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}

	return nil
}

// ValidateName checks if name is valid
func (v *InputValidator) ValidateName(name string) error {
	name = strings.TrimSpace(name)

	if len(name) > 100 {
		return ErrNameTooLong
	}

	// Allow letters, spaces, hyphens, apostrophes
	for _, char := range name {
		if !unicode.IsLetter(char) && char != ' ' && char != '-' && char != '\'' && char != '.' {
			return ErrNameInvalidChars
		}
	}

	return nil
}

// SanitizeName removes potentially dangerous characters
func (v *InputValidator) SanitizeName(name string) string {
	name = strings.TrimSpace(name)

	// Remove HTML tags
	name = removeHTMLTags(name)

	// Remove control characters
	var result strings.Builder
	for _, char := range name {
		if !unicode.IsControl(char) {
			result.WriteRune(char)
		}
	}

	return result.String()
}

// SanitizeEmail normalizes email
func (v *InputValidator) SanitizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func removeHTMLTags(s string) string {
	var result strings.Builder
	inTag := false
	depth := 0

	for _, char := range s {
		if char == '<' {
			inTag = true
			depth++
			continue
		}
		if char == '>' && inTag {
			depth--
			if depth == 0 {
				inTag = false
			}
			continue
		}
		if !inTag {
			result.WriteRune(char)
		}
	}

	return result.String()
}
