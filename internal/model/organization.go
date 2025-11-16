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
	ID          uuid.UUID          `json:"id" db:"id"`
	Name        string             `json:"name" db:"name"`
	OwnerUserID uuid.UUID          `json:"owner_user_id" db:"owner_user_id"`
	Status      OrganizationStatus `json:"status" db:"status"`
	CreatedAt   time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" db:"updated_at"`
}
