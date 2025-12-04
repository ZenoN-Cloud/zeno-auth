package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/ZenoN-Cloud/zeno-auth/internal/repository/postgres"
	"github.com/ZenoN-Cloud/zeno-auth/internal/token"
	"github.com/ZenoN-Cloud/zeno-auth/internal/validator"
)

type PasswordService struct {
	userRepo        UserRepository
	refreshRepo     RefreshTokenRepository
	passwordManager *token.PasswordManager
	auditService    *AuditService
	emailService    *EmailService
	db              *postgres.DB
}

func NewPasswordService(
	userRepo UserRepository,
	refreshRepo RefreshTokenRepository,
	passwordManager *token.PasswordManager,
	auditService *AuditService,
	emailService *EmailService,
	db *postgres.DB,
) *PasswordService {
	return &PasswordService{
		userRepo:        userRepo,
		refreshRepo:     refreshRepo,
		passwordManager: passwordManager,
		auditService:    auditService,
		emailService:    emailService,
		db:              db,
	}
}

func (s *PasswordService) ChangePassword(
	ctx context.Context,
	userID uuid.UUID,
	currentPassword, newPassword, ipAddress, userAgent string,
) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	valid, err := s.passwordManager.Verify(ctx, currentPassword, user.PasswordHash)
	if err != nil || !valid {
		return fmt.Errorf("current password is incorrect")
	}

	// Validate new password strength
	passwordValidator := validator.NewPasswordValidator()
	if err := passwordValidator.Validate(newPassword); err != nil {
		return err
	}

	newHash, err := s.passwordManager.Hash(ctx, newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Start transaction for atomic password change
	tx, err := s.db.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Update password
	user.PasswordHash = newHash
	if err := s.userRepo.UpdateTx(ctx, tx, user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Revoke all refresh tokens to force re-login
	if err := s.refreshRepo.RevokeByUserIDTx(ctx, tx, userID); err != nil {
		return fmt.Errorf("failed to revoke tokens: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Audit log (outside transaction)
	if s.auditService != nil {
		_ = s.auditService.Log(ctx, &userID, "password_changed", nil, ipAddress, userAgent)
	}

	// Send email notification (outside transaction)
	if s.emailService != nil {
		go func() { _ = s.emailService.SendPasswordChangedNotification(ctx, userID) }()
	}

	return nil
}
