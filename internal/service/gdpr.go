package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/ZenoN-Cloud/zeno-auth/internal/model"
)

type GDPRService struct {
	userRepo       UserRepository
	orgRepo        OrganizationRepository
	membershipRepo MembershipRepository
	refreshRepo    RefreshTokenRepository
	consentRepo    ConsentRepository
	auditRepo      AuditLogRepository
}

func NewGDPRService(
	userRepo UserRepository,
	orgRepo OrganizationRepository,
	membershipRepo MembershipRepository,
	refreshRepo RefreshTokenRepository,
	consentRepo ConsentRepository,
	auditRepo AuditLogRepository,
) *GDPRService {
	return &GDPRService{
		userRepo:       userRepo,
		orgRepo:        orgRepo,
		membershipRepo: membershipRepo,
		refreshRepo:    refreshRepo,
		consentRepo:    consentRepo,
		auditRepo:      auditRepo,
	}
}

type UserDataExport struct {
	User          *model.User            `json:"user"`
	Organizations []*model.Organization  `json:"organizations"`
	Memberships   []*model.OrgMembership `json:"memberships"`
	Consents      []*model.UserConsent   `json:"consents"`
	AuditLogs     []*model.AuditLog      `json:"audit_logs"`
}

func (s *GDPRService) ExportUserData(ctx context.Context, userID uuid.UUID) (interface{}, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	orgs, err := s.orgRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organizations: %w", err)
	}

	memberships, err := s.membershipRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get memberships: %w", err)
	}

	consents, err := s.consentRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get consents: %w", err)
	}

	auditLogs, err := s.auditRepo.GetByUserID(ctx, userID, 1000)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs: %w", err)
	}

	export := &UserDataExport{
		User:          user,
		Organizations: orgs,
		Memberships:   memberships,
		Consents:      consents,
		AuditLogs:     auditLogs,
	}
	return export, nil
}

func (s *GDPRService) DeleteUserAccount(ctx context.Context, userID uuid.UUID) error {
	// Revoke all refresh tokens
	if err := s.refreshRepo.RevokeByUserID(ctx, userID); err != nil {
		return fmt.Errorf("failed to revoke tokens: %w", err)
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Anonymize user data
	user.Email = fmt.Sprintf("deleted_%s@deleted.local", userID.String())
	user.FullName = "Deleted User"
	user.PasswordHash = uuid.New().String()
	user.IsActive = false

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to anonymize user: %w", err)
	}

	// Anonymize audit logs (GDPR compliance - 2 year retention)
	if err := s.auditRepo.AnonymizeByUserID(ctx, userID); err != nil {
		return fmt.Errorf("failed to anonymize user audit logs: %w", err)
	}

	// Note: Consents and memberships are kept for audit purposes
	// They can be cleaned up by the cleanup job after retention period

	return nil
}

func (s *GDPRService) DeleteOrganizationData(ctx context.Context, orgID uuid.UUID) error {
	// Organization data deletion is handled by cascade in database
	// Audit logs are kept for compliance
	return nil
}
