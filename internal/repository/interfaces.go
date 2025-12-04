package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/ZenoN-Cloud/zeno-auth/internal/model"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	CreateTx(ctx context.Context, tx pgx.Tx, user *model.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	UpdateTx(ctx context.Context, tx pgx.Tx, user *model.User) error
}

type OrganizationRepository interface {
	Create(ctx context.Context, org *model.Organization) error
	CreateTx(ctx context.Context, tx pgx.Tx, org *model.Organization) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Organization, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Organization, error)
	Update(ctx context.Context, org *model.Organization) error
}

type MembershipRepository interface {
	Create(ctx context.Context, membership *model.OrgMembership) error
	CreateTx(ctx context.Context, tx pgx.Tx, membership *model.OrgMembership) error
	GetByUserAndOrg(ctx context.Context, userID, orgID uuid.UUID) (*model.OrgMembership, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*model.OrgMembership, error)
	Update(ctx context.Context, membership *model.OrgMembership) error
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *model.RefreshToken) error
	CreateTx(ctx context.Context, tx pgx.Tx, token *model.RefreshToken) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*model.RefreshToken, error)
	RevokeByUserID(ctx context.Context, userID uuid.UUID) error
	RevokeByUserIDTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error
	RevokeByID(ctx context.Context, id uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}

type ConsentRepository interface {
	Create(ctx context.Context, consent *model.UserConsent) error
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*model.UserConsent, error)
	GetByUserAndType(ctx context.Context, userID uuid.UUID, consentType model.ConsentType) (*model.UserConsent, error)
	Revoke(ctx context.Context, userID uuid.UUID, consentType model.ConsentType) error
}

type AuditLogRepository interface {
	Create(ctx context.Context, log *model.AuditLog) error
	GetByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]*model.AuditLog, error)
	DeleteOlderThan(ctx context.Context, date time.Time) error
	AnonymizeByUserID(ctx context.Context, userID uuid.UUID) error
}
