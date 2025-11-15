package service

import (
	"context"

	"github.com/ZenoN-Cloud/zeno-auth/internal/model"
	"github.com/ZenoN-Cloud/zeno-auth/internal/repository"
	"github.com/google/uuid"
)

type UserService struct {
	userRepo       repository.UserRepository
	membershipRepo repository.MembershipRepository
}

func NewUserService(
	userRepo repository.UserRepository,
	membershipRepo repository.MembershipRepository,
) *UserService {
	return &UserService{
		userRepo:       userRepo,
		membershipRepo: membershipRepo,
	}
}

func (s *UserService) GetProfile(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

func (s *UserService) GetMemberships(ctx context.Context, userID uuid.UUID) ([]*model.OrgMembership, error) {
	return s.membershipRepo.GetByUserID(ctx, userID)
}