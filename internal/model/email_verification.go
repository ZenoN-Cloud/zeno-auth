package model

import (
	"time"

	"github.com/google/uuid"
)

type EmailVerification struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	TokenHash  string     `json:"-"`
	ExpiresAt  time.Time  `json:"expires_at"`
	VerifiedAt *time.Time `json:"verified_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}
