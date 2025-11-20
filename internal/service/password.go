package service

import (
	"context"
	"fmt"

	"github.com/ZenoN-Cloud/zeno-auth/internal/token"
	"github.com/ZenoN-Cloud/zeno-auth/internal/validator"
	"github.com/google/uuid"
)

type PasswordService struct {
	userRepo        UserRepository
	refreshRepo     RefreshTokenRepository
	passwordManager *token.PasswordManager
	auditService    *AuditService
}

func NewPasswordService(
	userRepo UserRepository,
	refreshRepo RefreshTokenRepository,
	passwordManager *token.PasswordManager,
	auditService *AuditService,
) *PasswordService {
	return &PasswordService{
		userRepo:        userRepo,
		refreshRepo:     refreshRepo,
		passwordManager: passwordManager,
		auditService:    auditService,
	}
}

func (s *PasswordService) ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword, ipAddress, userAgent string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	valid, err := s.passwordManager.Verify(context.Background(), currentPassword, user.PasswordHash)
	if err != nil || !valid {
		return fmt.Errorf("current password is incorrect")
	}

	// Validate new password strength
	passwordValidator := validator.NewPasswordValidator()
	if err := passwordValidator.Validate(newPassword); err != nil {
		return err
	}

	newHash, err := s.passwordManager.Hash(context.Background(), newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user.PasswordHash = newHash
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Revoke all refresh tokens to force re-login
	if err := s.refreshRepo.RevokeByUserID(ctx, userID); err != nil {
		return fmt.Errorf("failed to revoke tokens: %w", err)
	}

	// Audit log
	if s.auditService != nil {
		s.auditService.Log(ctx, &userID, "password_changed", nil, ipAddress, userAgent)
	}

	return nil
}
