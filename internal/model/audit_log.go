package model

import (
	"time"

	"github.com/google/uuid"
)

type AuditEventType string

const (
	EventUserRegistered  AuditEventType = "user_registered"
	EventUserLoggedIn    AuditEventType = "user_logged_in"
	EventUserLoggedOut   AuditEventType = "user_logged_out"
	EventLoginFailed     AuditEventType = "login_failed"
	EventPasswordChanged AuditEventType = "password_changed"
	EventEmailChanged    AuditEventType = "email_changed"
	EventAccountDeleted  AuditEventType = "account_deleted"
	EventDataExported    AuditEventType = "data_exported"
	EventConsentGranted  AuditEventType = "consent_granted"
	EventConsentRevoked  AuditEventType = "consent_revoked"
)

type AuditLog struct {
	ID        uuid.UUID              `json:"id" db:"id"`
	UserID    *uuid.UUID             `json:"user_id,omitempty" db:"user_id"`
	EventType AuditEventType         `json:"event_type" db:"event_type"`
	EventData map[string]interface{} `json:"event_data,omitempty" db:"event_data"`
	IPAddress string                 `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent string                 `json:"user_agent,omitempty" db:"user_agent"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
}
