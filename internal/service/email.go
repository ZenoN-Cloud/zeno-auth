package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/ZenoN-Cloud/zeno-auth/internal/model"
)

var (
	ErrTokenExpired = errors.New("verification token expired")
	ErrTokenUsed    = errors.New("verification token already used")
)

type EmailVerificationRepository interface {
	Create(ctx context.Context, verification *model.EmailVerification) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*model.EmailVerification, error)
	MarkAsVerified(ctx context.Context, id uuid.UUID) error
	DeleteExpired(ctx context.Context) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}

type EmailService struct {
	verificationRepo EmailVerificationRepository
	userRepo         UserRepository
	auditService     *AuditService
	emailSender      EmailSender
}

func NewEmailService(
	verificationRepo EmailVerificationRepository,
	userRepo UserRepository,
	auditService *AuditService,
	frontendBaseURL string,
) *EmailService {
	sender := NewSendGridEmailSender(frontendBaseURL)
	log.Info().Bool("email_sender_initialized", sender != nil).Msg("EmailService created")
	return &EmailService{
		verificationRepo: verificationRepo,
		userRepo:         userRepo,
		auditService:     auditService,
		emailSender:      sender,
	}
}

func (s *EmailService) SendVerificationEmail(ctx context.Context, userID uuid.UUID) (string, error) {
	token, err := generateToken()
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	tokenHash := hashToken(token)
	verification := &model.EmailVerification{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := s.verificationRepo.Create(ctx, verification); err != nil {
		return "", fmt.Errorf("failed to create verification: %w", err)
	}

	// Get user email
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	// Send email
	if s.emailSender != nil {
		if err := s.emailSender.SendVerificationEmail(ctx, user.Email, token); err != nil {
			log.Error().Err(err).Str("email", user.Email).Msg("Failed to send verification email")
			return "", fmt.Errorf("failed to send verification email: %w", err)
		}
		log.Info().Str("email", user.Email).Msg("Verification email sent successfully")
	} else {
		log.Warn().Msg("EmailSender is nil, cannot send email")
		return "", fmt.Errorf("email service not configured")
	}

	return token, nil
}

func (s *EmailService) VerifyEmail(ctx context.Context, token, ipAddress, userAgent string) error {
	tokenHash := hashToken(token)
	verification, err := s.verificationRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return errors.New("invalid verification token")
	}

	if verification.VerifiedAt != nil {
		return ErrTokenUsed
	}

	if time.Now().After(verification.ExpiresAt) {
		return ErrTokenExpired
	}

	if err := s.verificationRepo.MarkAsVerified(ctx, verification.ID); err != nil {
		return fmt.Errorf("failed to mark as verified: %w", err)
	}

	// Mark user as email verified
	user, err := s.userRepo.GetByID(ctx, verification.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	user.IsActive = true
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Audit log
	if s.auditService != nil {
		_ = s.auditService.Log(ctx, &verification.UserID, "email_verified", nil, ipAddress, userAgent)
	}

	return nil
}

func (s *EmailService) ResendVerification(ctx context.Context, userID uuid.UUID) (string, error) {
	// Delete old tokens
	if err := s.verificationRepo.DeleteByUserID(ctx, userID); err != nil {
		return "", fmt.Errorf("failed to delete old tokens: %w", err)
	}

	return s.SendVerificationEmail(ctx, userID)
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// SendAccountDeletionNotification sends email notification when account is deleted (GDPR Art. 34)
func (s *EmailService) SendAccountDeletionNotification(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("Failed to get user for deletion notification")
		return err
	}

	// TODO: Send actual email via SendGrid/AWS SES
	log.Info().
		Str("user_id", userID.String()).
		Str("email", user.Email).
		Msg("Account deletion notification (email sending not implemented)")

	return nil
}

// SendDataExportNotification sends email notification when data is exported (GDPR Art. 34)
func (s *EmailService) SendDataExportNotification(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("Failed to get user for export notification")
		return err
	}

	// TODO: Send actual email via SendGrid/AWS SES
	log.Info().
		Str("user_id", userID.String()).
		Str("email", user.Email).
		Msg("Data export notification (email sending not implemented)")

	return nil
}

// SendAccountLockoutNotification sends email notification when account is locked due to failed login attempts
func (s *EmailService) SendAccountLockoutNotification(
	ctx context.Context,
	userID uuid.UUID,
	lockedUntil time.Time,
) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("Failed to get user for lockout notification")
		return err
	}

	if s.emailSender != nil {
		if err := s.emailSender.SendAccountLockoutEmail(ctx, user.Email, lockedUntil.Format("2006-01-02 15:04:05 MST")); err != nil {
			log.Error().Err(err).Str("email", user.Email).Msg("Failed to send lockout email")
		}
	}

	return nil
}

// SendPasswordChangedNotification sends email notification when password is changed
func (s *EmailService) SendPasswordChangedNotification(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("Failed to get user for password change notification")
		return err
	}

	if s.emailSender != nil {
		if err := s.emailSender.SendPasswordChangedEmail(ctx, user.Email); err != nil {
			log.Error().Err(err).Str("email", user.Email).Msg("Failed to send password changed email")
			return fmt.Errorf("failed to send notification: %w", err)
		}
	}

	return nil
}
