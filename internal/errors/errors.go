package errors

import (
	"errors"
	"net/http"
)

// Domain errors
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailAlreadyUsed   = errors.New("email already used")
	ErrUserNotFound       = errors.New("user not found")
	ErrAccountLocked      = errors.New("account locked")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
	ErrTokenRevoked       = errors.New("token revoked")
	ErrInvalidFingerprint = errors.New("invalid fingerprint")
	ErrEmailNotVerified   = errors.New("email not verified")
	ErrInvalidResetToken  = errors.New("invalid reset token")
	ErrResetTokenExpired  = errors.New("reset token expired")
	ErrWeakPassword       = errors.New("password too weak")
	ErrInvalidInput       = errors.New("invalid input")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrNotFound           = errors.New("not found")
	ErrConflict           = errors.New("conflict")
	ErrRateLimitExceeded  = errors.New("rate limit exceeded")
	ErrInternalServer     = errors.New("internal server error")
)

// MapError maps domain errors to HTTP status codes and messages
func MapError(err error) (int, string) {
	switch {
	case errors.Is(err, ErrInvalidCredentials):
		return http.StatusUnauthorized, "Invalid email or password"
	case errors.Is(err, ErrEmailAlreadyUsed):
		return http.StatusConflict, "Email already registered"
	case errors.Is(err, ErrUserNotFound):
		return http.StatusNotFound, "User not found"
	case errors.Is(err, ErrAccountLocked):
		return http.StatusForbidden, "Account locked due to too many failed login attempts"
	case errors.Is(err, ErrInvalidToken):
		return http.StatusUnauthorized, "Invalid token"
	case errors.Is(err, ErrTokenExpired):
		return http.StatusUnauthorized, "Token expired"
	case errors.Is(err, ErrTokenRevoked):
		return http.StatusUnauthorized, "Token revoked"
	case errors.Is(err, ErrInvalidFingerprint):
		return http.StatusUnauthorized, "Invalid session fingerprint"
	case errors.Is(err, ErrEmailNotVerified):
		return http.StatusForbidden, "Email not verified"
	case errors.Is(err, ErrInvalidResetToken):
		return http.StatusBadRequest, "Invalid or expired reset token"
	case errors.Is(err, ErrResetTokenExpired):
		return http.StatusBadRequest, "Reset token expired"
	case errors.Is(err, ErrWeakPassword):
		return http.StatusBadRequest, "Password does not meet security requirements"
	case errors.Is(err, ErrInvalidInput):
		return http.StatusBadRequest, "Invalid input"
	case errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized, "Unauthorized"
	case errors.Is(err, ErrForbidden):
		return http.StatusForbidden, "Forbidden"
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound, "Resource not found"
	case errors.Is(err, ErrConflict):
		return http.StatusConflict, "Resource conflict"
	case errors.Is(err, ErrRateLimitExceeded):
		return http.StatusTooManyRequests, "Rate limit exceeded"
	default:
		return http.StatusInternalServerError, "Internal server error"
	}
}
