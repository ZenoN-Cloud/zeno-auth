package handler

import (
	"context"

	"github.com/google/uuid"
)

type RegisterRequest struct {
	Email            string `json:"email" binding:"required,email"`
	Password         string `json:"password" binding:"required,min=8"`
	FullName         string `json:"full_name" binding:"required"`
	OrganizationName string `json:"organization_name" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type UserResponse struct {
	ID       uuid.UUID `json:"id"`
	Email    string    `json:"email"`
	FullName string    `json:"full_name"`
	IsActive bool      `json:"is_active"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// MetricsCollector interface for collecting metrics
type MetricsCollector interface {
	IncrementRegistrations()
	IncrementLogins()
	IncrementLoginFailures()
	IncrementTokenRefreshes()
	SetActiveSessions(count int64)
}

// AuditService interface for audit logging
type AuditService interface {
	Log(ctx context.Context, userID *uuid.UUID, eventType interface{}, eventData map[string]interface{}, ipAddress, userAgent string) error
}
