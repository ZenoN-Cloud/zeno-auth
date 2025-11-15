package model

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleOwner  Role = "OWNER"
	RoleAdmin  Role = "ADMIN"
	RoleMember Role = "MEMBER"
	RoleViewer Role = "VIEWER"
)

type OrgMembership struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	OrgID     uuid.UUID `json:"org_id" db:"org_id"`
	Role      Role      `json:"role" db:"role"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}