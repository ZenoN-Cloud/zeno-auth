package service

import (
	"context"
	"database/sql"

	"github.com/ZenoN-Cloud/zeno-auth/internal/model"
	"github.com/google/uuid"
)

type AuthServiceInterface interface {
	Register(ctx context.Context, email, password, fullName, organizationName string) (*model.User, error)
	Login(ctx context.Context, email, password, userAgent, ipAddress string) (string, string, error)
	RefreshToken(ctx context.Context, refreshToken, userAgent, ipAddress string) (string, error)
	Logout(ctx context.Context, userID uuid.UUID) error
}

type UserServiceInterface interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*model.User, error)
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *model.RefreshToken) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*model.RefreshToken, error)
	GetActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*model.RefreshToken, error)
	RevokeByUserID(ctx context.Context, userID uuid.UUID) error
	RevokeByUserIDTx(ctx context.Context, tx *sql.Tx, userID uuid.UUID) error
	RevokeByID(ctx context.Context, id uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	UpdateTx(ctx context.Context, tx *sql.Tx, user *model.User) error
}

type OrganizationRepository interface {
	Create(ctx context.Context, org *model.Organization) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Organization, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Organization, error)
	Update(ctx context.Context, org *model.Organization) error
}

type MembershipRepository interface {
	Create(ctx context.Context, membership *model.OrgMembership) error
	GetByUserAndOrg(ctx context.Context, userID, orgID uuid.UUID) (*model.OrgMembership, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*model.OrgMembership, error)
	Update(ctx context.Context, membership *model.OrgMembership) error
}
