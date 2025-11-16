package service

import (
	"context"

	"github.com/ZenoN-Cloud/zeno-auth/internal/model"
	"github.com/google/uuid"
)

type AuthServiceInterface interface {
	Register(ctx context.Context, email, password, fullName string) (*model.User, error)
	Login(ctx context.Context, email, password, userAgent, ipAddress string) (string, string, error)
	RefreshToken(ctx context.Context, refreshToken string) (string, error)
	Logout(ctx context.Context, userID uuid.UUID) error
}

type UserServiceInterface interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*model.User, error)
}
