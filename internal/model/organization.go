package model

import (
	"time"

	"github.com/google/uuid"
)

type OrganizationStatus string

const (
	OrgStatusActive    OrganizationStatus = "active"
	OrgStatusTrial     OrganizationStatus = "trial"
	OrgStatusSuspended OrganizationStatus = "suspended"
)

type Organization struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	Name           string     `json:"name" db:"name"`
	OwnerUserID    uuid.UUID  `json:"owner_user_id" db:"owner_user_id"`
	Status         string     `json:"status" db:"status"`
	Country        string     `json:"country" db:"country"`
	Currency       string     `json:"currency" db:"currency"`
	TrialEndsAt    *time.Time `json:"trial_ends_at,omitempty" db:"trial_ends_at"`
	SubscriptionID *uuid.UUID `json:"subscription_id,omitempty" db:"subscription_id"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}
