package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                  uuid.UUID  `json:"id" db:"id"`
	Email               string     `json:"email" db:"email"`
	PasswordHash        string     `json:"-" db:"password_hash" log:"-"`
	FullName            string     `json:"full_name" db:"full_name"`
	IsActive            bool       `json:"is_active" db:"is_active"`
	FailedLoginAttempts int        `json:"-" db:"failed_login_attempts"`
	LockedUntil         *time.Time `json:"-" db:"locked_until"`
	DeletedAt           *time.Time `json:"-" db:"deleted_at"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
}
