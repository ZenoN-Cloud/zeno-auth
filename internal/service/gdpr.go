package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/ZenoN-Cloud/zeno-auth/internal/model"
	"github.com/ZenoN-Cloud/zeno-auth/internal/repository/postgres"
)

type GDPRService struct {
	userRepo       UserRepository
	orgRepo        OrganizationRepository
	membershipRepo MembershipRepository
	refreshRepo    RefreshTokenRepository
	consentRepo    ConsentRepository
	auditRepo      AuditLogRepository
	db             *postgres.DB
}

func NewGDPRService(
	userRepo UserRepository,
	orgRepo OrganizationRepository,
	membershipRepo MembershipRepository,
	refreshRepo RefreshTokenRepository,
	consentRepo ConsentRepository,
	auditRepo AuditLogRepository,
	db *postgres.DB,
) *GDPRService {
	return &GDPRService{
		userRepo:       userRepo,
		orgRepo:        orgRepo,
		membershipRepo: membershipRepo,
		refreshRepo:    refreshRepo,
		consentRepo:    consentRepo,
		auditRepo:      auditRepo,
		db:             db,
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
	if userID == uuid.Nil {
		return nil, fmt.Errorf("invalid user ID")
	}

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

	if user == nil {
		return nil, fmt.Errorf("user not found")
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
	if s.db == nil {
		return fmt.Errorf("database connection not available")
	}

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	// Start transaction for atomic GDPR deletion
	tx, err := s.db.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx) // Ignore rollback error in defer
	}()

	// Get user first
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	// Revoke all refresh tokens
	if err := s.refreshRepo.RevokeByUserIDTx(ctx, tx, userID); err != nil {
		return fmt.Errorf("failed to revoke tokens: %w", err)
	}

	// Anonymize user data
	user.Email = fmt.Sprintf("deleted_%s@deleted.local", userID.String())
	user.FullName = "Deleted User"
	user.PasswordHash = uuid.New().String()
	user.IsActive = false

	if err := s.userRepo.UpdateTx(ctx, tx, user); err != nil {
		return fmt.Errorf("failed to anonymize user: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Anonymize audit logs (outside transaction - can be async)
	go func() {
		ctx := context.Background()
		_ = s.auditRepo.AnonymizeByUserID(ctx, userID)
		// Errors are logged internally, can be retried later
	}()

	// Note: Consents and memberships are kept for audit purposes
	// and can be cleaned up by the cleanup job after a retention period.

	return nil
}

func (s *GDPRService) DeleteOrganizationData(ctx context.Context, orgID uuid.UUID) error {
	// Organization data deletion is handled by cascade in database
	// Audit logs are kept for compliance
	return nil
}
