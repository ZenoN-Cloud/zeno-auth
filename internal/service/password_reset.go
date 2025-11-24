package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ZenoN-Cloud/zeno-auth/internal/errors"
	"github.com/ZenoN-Cloud/zeno-auth/internal/model"
	"github.com/ZenoN-Cloud/zeno-auth/internal/token"
	"github.com/ZenoN-Cloud/zeno-auth/internal/validator"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type PasswordResetRepository interface {
	Create(ctx context.Context, token *model.PasswordResetToken) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*model.PasswordResetToken, error)
	MarkAsUsed(ctx context.Context, id uuid.UUID) error
	DeleteExpired(ctx context.Context) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}

type PasswordResetService struct {
	resetRepo       PasswordResetRepository
	userRepo        UserRepository
	refreshRepo     RefreshTokenRepository
	passwordManager *token.PasswordManager
	auditService    *AuditService
	emailSender     EmailSender
}

func NewPasswordResetService(
	resetRepo PasswordResetRepository,
	userRepo UserRepository,
	refreshRepo RefreshTokenRepository,
	passwordManager *token.PasswordManager,
	auditService *AuditService,
	frontendBaseURL string,
) *PasswordResetService {
	return &PasswordResetService{
		resetRepo:       resetRepo,
		userRepo:        userRepo,
		refreshRepo:     refreshRepo,
		passwordManager: passwordManager,
		auditService:    auditService,
		emailSender:     NewSendGridEmailSender(frontendBaseURL),
	}
}

func (s *PasswordResetService) RequestPasswordReset(ctx context.Context, email, ipAddress, userAgent string) (string, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Don't reveal if email exists
		log.Info().Str("email", email).Msg("Password reset requested for non-existent email")
		return "", nil
	}

	// Delete old tokens
	if err := s.resetRepo.DeleteByUserID(ctx, user.ID); err != nil {
		return "", fmt.Errorf("failed to delete old tokens: %w", err)
	}

	resetToken, err := generateResetToken()
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	tokenHash := hashResetToken(resetToken)
	passwordResetToken := &model.PasswordResetToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}

	if err := s.resetRepo.Create(ctx, passwordResetToken); err != nil {
		return "", fmt.Errorf("failed to create reset token: %w", err)
	}

	// Audit log
	if s.auditService != nil {
		_ = s.auditService.Log(ctx, &user.ID, "password_reset_requested", nil, ipAddress, userAgent)
	}

	// Send email
	if s.emailSender != nil {
		if err := s.emailSender.SendPasswordResetEmail(ctx, user.Email, resetToken); err != nil {
			log.Error().Err(err).Str("email", user.Email).Msg("Failed to send password reset email")
			return "", fmt.Errorf("failed to send reset email: %w", err)
		}
	}

	return resetToken, nil
}

func (s *PasswordResetService) ResetPassword(ctx context.Context, resetToken, newPassword, ipAddress, userAgent string) error {
	passwordValidator := validator.NewPasswordValidator()
	if err := passwordValidator.Validate(newPassword); err != nil {
		return err
	}

	hash := hashResetToken(resetToken)
	resetRecord, err := s.resetRepo.GetByTokenHash(ctx, hash)
	if err != nil {
		return errors.ErrInvalidResetToken
	}

	if resetRecord.UsedAt != nil {
		return errors.ErrInvalidResetToken
	}

	if time.Now().After(resetRecord.ExpiresAt) {
		return errors.ErrResetTokenExpired
	}

	user, err := s.userRepo.GetByID(ctx, resetRecord.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Hash new password
	newHash, err := s.passwordManager.Hash(context.Background(), newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	user.PasswordHash = newHash
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Mark token as used
	if err := s.resetRepo.MarkAsUsed(ctx, resetRecord.ID); err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	// Revoke all refresh tokens
	if err := s.refreshRepo.RevokeByUserID(ctx, user.ID); err != nil {
		return fmt.Errorf("failed to revoke tokens: %w", err)
	}

	// Audit log
	if s.auditService != nil {
		_ = s.auditService.Log(ctx, &user.ID, "password_reset_completed", nil, ipAddress, userAgent)
	}

	return nil
}

func generateResetToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashResetToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
