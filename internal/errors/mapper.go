package errors

import (
	"errors"
)

type HTTPError struct {
	StatusCode int    `json:"-"`
	Code       string `json:"code"`
	Message    string `json:"message"`
}

func MapErrorToHTTP(err error) HTTPError {
	switch {
	// Authentication errors
	case errors.Is(err, ErrInvalidCredentials):
		return HTTPError{401, "invalid_credentials", "Invalid email or password"}
	case errors.Is(err, ErrAccountLocked):
		return HTTPError{423, "account_locked", "Account is temporarily locked due to too many failed login attempts"}
	case errors.Is(err, ErrInvalidToken):
		return HTTPError{401, "invalid_token", "Invalid or expired token"}
	case errors.Is(err, ErrTokenExpired):
		return HTTPError{401, "token_expired", "Token has expired"}
	case errors.Is(err, ErrTokenRevoked):
		return HTTPError{401, "token_revoked", "Token has been revoked"}
	case errors.Is(err, ErrInvalidFingerprint):
		return HTTPError{401, "invalid_fingerprint", "Invalid session fingerprint"}

	// Registration errors
	case errors.Is(err, ErrEmailAlreadyUsed):
		return HTTPError{409, "email_exists", "Email already registered"}
	case errors.Is(err, ErrUserNotFound):
		return HTTPError{404, "user_not_found", "User not found"}
	case errors.Is(err, ErrEmailNotVerified):
		return HTTPError{403, "email_not_verified", "Email address not verified"}

	// Validation errors
	case errors.Is(err, ErrWeakPassword):
		return HTTPError{400, "password_too_weak", "Password does not meet security requirements"}
	case errors.Is(err, ErrInvalidInput):
		return HTTPError{400, "invalid_input", "Invalid input data"}

	// Password reset errors
	case errors.Is(err, ErrInvalidResetToken):
		return HTTPError{400, "reset_token_invalid", "Invalid or expired reset token"}
	case errors.Is(err, ErrResetTokenExpired):
		return HTTPError{400, "reset_token_expired", "Reset token has expired"}

	// Rate limiting
	case errors.Is(err, ErrRateLimitExceeded):
		return HTTPError{429, "rate_limited", "Too many requests, please try again later"}

	// Generic errors
	case errors.Is(err, ErrUnauthorized):
		return HTTPError{401, "unauthorized", "Unauthorized"}
	case errors.Is(err, ErrForbidden):
		return HTTPError{403, "forbidden", "Forbidden"}
	case errors.Is(err, ErrNotFound):
		return HTTPError{404, "not_found", "Resource not found"}
	case errors.Is(err, ErrConflict):
		return HTTPError{409, "conflict", "Resource conflict"}

	// Default internal error
	default:
		return HTTPError{500, "internal_error", "Internal server error"}
	}
}
