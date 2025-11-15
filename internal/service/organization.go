package service

import (
	"context"

	"github.com/ZenoN-Cloud/zeno-auth/internal/model"
	"github.com/ZenoN-Cloud/zeno-auth/internal/repository"
	"github.com/google/uuid"
)

type OrganizationService struct {
	orgRepo        repository.OrganizationRepository
	membershipRepo repository.MembershipRepository
}

func NewOrganizationService(
	orgRepo repository.OrganizationRepository,
	membershipRepo repository.MembershipRepository,
) *OrganizationService {
	return &OrganizationService{
		orgRepo:        orgRepo,
		membershipRepo: membershipRepo,
	}
}

func (s *OrganizationService) Create(ctx context.Context, name string, ownerUserID uuid.UUID) (*model.Organization, error) {
	org := &model.Organization{
		Name:        name,
		OwnerUserID: ownerUserID,
		Status:      model.OrgStatusActive,
	}

	if err := s.orgRepo.Create(ctx, org); err != nil {
		return nil, err
	}

	membership := &model.OrgMembership{
		UserID:   ownerUserID,
		OrgID:    org.ID,
		Role:     model.RoleOwner,
		IsActive: true,
	}

	if err := s.membershipRepo.Create(ctx, membership); err != nil {
		return nil, err
	}

	return org, nil
}

func (s *OrganizationService) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Organization, error) {
	return s.orgRepo.GetByUserID(ctx, userID)
}