package model

import (
	"time"

	"github.com/google/uuid"
)

type ConsentType string

const (
	ConsentTypeTerms     ConsentType = "terms"
	ConsentTypePrivacy   ConsentType = "privacy"
	ConsentTypeMarketing ConsentType = "marketing"
	ConsentTypeAnalytics ConsentType = "analytics"
)

type UserConsent struct {
	ID          uuid.UUID   `json:"id" db:"id"`
	UserID      uuid.UUID   `json:"user_id" db:"user_id"`
	ConsentType ConsentType `json:"consent_type" db:"consent_type"`
	Version     string      `json:"version" db:"version"`
	Granted     bool        `json:"granted" db:"granted"`
	GrantedAt   time.Time   `json:"granted_at" db:"granted_at"`
	RevokedAt   *time.Time  `json:"revoked_at,omitempty" db:"revoked_at"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
}
