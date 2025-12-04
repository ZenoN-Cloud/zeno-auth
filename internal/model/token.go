package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidUserID = errors.New("invalid user ID")
	ErrInvalidOrgID  = errors.New("invalid org ID")
	ErrInvalidToken  = errors.New("invalid token")
	ErrTokenExpired  = errors.New("token expired")
)

type RefreshToken struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	UserID          uuid.UUID  `json:"user_id" db:"user_id"`
	OrgID           uuid.UUID  `json:"org_id" db:"org_id"`
	TokenHash       string     `json:"-" db:"token_hash"`
	UserAgent       string     `json:"user_agent" db:"user_agent"`
	IPAddress       string     `json:"ip_address" db:"ip_address"`
	FingerprintHash *string    `json:"-" db:"fingerprint_hash"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	ExpiresAt       time.Time  `json:"expires_at" db:"expires_at"`
	RevokedAt       *time.Time `json:"revoked_at" db:"revoked_at"`
}

func (rt *RefreshToken) Validate() error {
	if rt.UserID == uuid.Nil {
		return ErrInvalidUserID
	}
	if rt.OrgID == uuid.Nil {
		return ErrInvalidOrgID
	}
	if rt.TokenHash == "" {
		return ErrInvalidToken
	}
	if rt.ExpiresAt.Before(time.Now()) {
		return ErrTokenExpired
	}
	return nil
}
