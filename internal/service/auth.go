package service

import (
	"context"
	stdErrors "errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"

	appErrors "github.com/ZenoN-Cloud/zeno-auth/internal/errors"
	"github.com/ZenoN-Cloud/zeno-auth/internal/model"
	"github.com/ZenoN-Cloud/zeno-auth/internal/repository"
	"github.com/ZenoN-Cloud/zeno-auth/internal/repository/postgres"
	"github.com/ZenoN-Cloud/zeno-auth/internal/token"
	"github.com/ZenoN-Cloud/zeno-auth/internal/validator"
)

type BillingClient interface {
	CreateTrialSubscription(ctx context.Context, orgID uuid.UUID) error
}

type AuthService struct {
	userRepo        repository.UserRepository
	orgRepo         repository.OrganizationRepository
	membershipRepo  repository.MembershipRepository
	refreshRepo     repository.RefreshTokenRepository
	jwtManager      *token.JWTManager
	refreshManager  *token.RefreshManager
	passwordManager token.PasswordHasher
	emailService    *EmailService
	billingClient   BillingClient
	config          *Config
	db              *postgres.DB
}

func NewAuthService(
	userRepo repository.UserRepository,
	orgRepo repository.OrganizationRepository,
	membershipRepo repository.MembershipRepository,
	refreshRepo repository.RefreshTokenRepository,
	jwtManager *token.JWTManager,
	refreshManager *token.RefreshManager,
	passwordManager token.PasswordHasher,
	emailService *EmailService,
	billingClient BillingClient,
	config *Config,
	db *postgres.DB,
) *AuthService {
	return &AuthService{
		userRepo:        userRepo,
		orgRepo:         orgRepo,
		membershipRepo:  membershipRepo,
		refreshRepo:     refreshRepo,
		jwtManager:      jwtManager,
		refreshManager:  refreshManager,
		passwordManager: passwordManager,
		emailService:    emailService,
		billingClient:   billingClient,
		config:          config,
		db:              db,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password, fullName, organizationName string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	email = strings.ToLower(strings.TrimSpace(email))

	// Validate password strength
	passwordValidator := validator.NewPasswordValidator()
	if err := passwordValidator.Validate(password); err != nil {
		// propagate validator error upward (it maps to 400)
		return nil, err
	}

	_, err := s.userRepo.GetByEmail(ctx, email)
	if err == nil {
		return nil, appErrors.ErrEmailAlreadyUsed
	}
	if !stdErrors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	passwordHash, err := s.passwordManager.Hash(ctx, password)
	if err != nil {
		return nil, err
	}

	// Start transaction for atomic registration
	tx, err := s.db.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback(ctx)
			panic(r)
		}
	}()

	user := &model.User{
		Email:        email,
		PasswordHash: passwordHash,
		FullName:     fullName,
		IsActive:     false, // User must verify email first
	}

	// Create user
	if err := s.userRepo.CreateTx(ctx, tx, user); err != nil {
		_ = tx.Rollback(ctx)
		return nil, err
	}

	// Create organization for user
	org := &model.Organization{
		Name:        organizationName,
		OwnerUserID: user.ID,
		Status:      "created",
	}

	if err := s.orgRepo.CreateTx(ctx, tx, org); err != nil {
		_ = tx.Rollback(ctx)
		return nil, err
	}

	// Create membership with OWNER role
	membership := &model.OrgMembership{
		UserID:   user.ID,
		OrgID:    org.ID,
		Role:     model.RoleOwner,
		IsActive: true,
	}

	if err := s.membershipRepo.CreateTx(ctx, tx, membership); err != nil {
		_ = tx.Rollback(ctx)
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	// Create trial subscription in billing service (async, don't fail registration)
	if s.billingClient != nil {
		go func() {
			bgCtx := context.Background()
			if err := s.billingClient.CreateTrialSubscription(bgCtx, org.ID); err != nil {
				log.Error().Err(err).Str("org_id", org.ID.String()).Msg("Failed to create trial subscription")
			}
		}()
	}

	// Send email verification (outside transaction)
	if s.emailService != nil {
		_, _ = s.emailService.SendVerificationEmail(ctx, user.ID)
		// Errors are logged internally, don't fail registration
	}

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password, userAgent, ipAddress string) (string, string, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if stdErrors.Is(err, pgx.ErrNoRows) {
			return "", "", appErrors.ErrInvalidCredentials
		}
		return "", "", err
	}

	if !user.IsActive {
		return "", "", appErrors.ErrInvalidCredentials
	}

	// LockedUntil
	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		return "", "", appErrors.ErrInvalidCredentials
	}

	valid, err := s.passwordManager.Verify(ctx, password, user.PasswordHash)
	if err != nil {
		return "", "", err
	}
	// Update user state based on login result
	needsUpdate := false
	if !valid {
		user.FailedLoginAttempts++
		if user.FailedLoginAttempts >= 5 {
			lockUntil := time.Now().Add(30 * time.Minute)
			user.LockedUntil = &lockUntil
			if s.emailService != nil {
				go func() { _ = s.emailService.SendAccountLockoutNotification(ctx, user.ID, lockUntil) }()
			}
		}
		needsUpdate = true
	} else if user.FailedLoginAttempts > 0 || user.LockedUntil != nil {
		// Reset failed attempts on successful login
		user.FailedLoginAttempts = 0
		user.LockedUntil = nil
		needsUpdate = true
	}

	if needsUpdate {
		if err := s.userRepo.Update(ctx, user); err != nil {
			log.Error().Err(err).Str("user_id", user.ID.String()).Msg("Failed to update user login state")
		}
	}

	if !valid {
		return "", "", appErrors.ErrInvalidCredentials
	}

	// Get user's first active membership (includes org and role)
	memberships, err := s.membershipRepo.GetByUserID(ctx, user.ID)
	if err != nil || len(memberships) == 0 {
		log.Error().Err(err).Str("user_id", user.ID.String()).Msg("User has no active memberships")
		return "", "", appErrors.ErrInvalidCredentials
	}

	// Use first active membership
	membership := memberships[0]
	orgID := membership.OrgID
	roles := []string{string(membership.Role)}

	accessToken, err := s.jwtManager.Generate(ctx, user.ID, orgID, roles, s.config.AccessTokenTTL)
	if err != nil {
		return "", "", err
	}

	refreshTokenStr, err := s.refreshManager.Generate(ctx)
	if err != nil {
		return "", "", err
	}

	// Generate fingerprint for session security
	fingerprint, err := token.GenerateFingerprint(userAgent, ipAddress, "")
	if err != nil {
		return "", "", err
	}
	refreshToken, err := s.refreshManager.CreateToken(ctx, user.ID, orgID, refreshTokenStr, userAgent, ipAddress, s.config.RefreshTokenTTL)
	if err != nil {
		return "", "", err
	}
	refreshToken.FingerprintHash = &fingerprint
	if err := s.refreshRepo.Create(ctx, refreshToken); err != nil {
		return "", "", err
	}

	return accessToken, refreshTokenStr, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshTokenStr, userAgent, ipAddress string) (string, error) {
	tokenHash, err := s.refreshManager.Hash(ctx, refreshTokenStr)
	if err != nil {
		return "", err
	}

	refreshToken, err := s.refreshRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		if stdErrors.Is(err, pgx.ErrNoRows) {
			return "", appErrors.ErrInvalidCredentials
		}
		return "", err
	}

	if refreshToken.RevokedAt != nil || refreshToken.ExpiresAt.Before(time.Now()) {
		return "", appErrors.ErrInvalidCredentials
	}

	if refreshToken.FingerprintHash != nil && *refreshToken.FingerprintHash != "" {
		currentFingerprint, err := token.GenerateFingerprint(userAgent, ipAddress, "")
		if err != nil {
			return "", err
		}
		if currentFingerprint != *refreshToken.FingerprintHash {
			return "", appErrors.ErrInvalidFingerprint
		}
	}

	// Get roles if user has organization membership
	var roles []string
	if refreshToken.OrgID != uuid.Nil {
		if membership, err := s.membershipRepo.GetByUserAndOrg(ctx, refreshToken.UserID, refreshToken.OrgID); err == nil && membership != nil {
			roles = []string{string(membership.Role)}
		}
	}
	return s.jwtManager.Generate(ctx, refreshToken.UserID, refreshToken.OrgID, roles, s.config.AccessTokenTTL)
}

func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	return s.refreshRepo.RevokeByUserID(ctx, userID)
}

func (s *AuthService) LogoutToken(ctx context.Context, refreshTokenStr string) error {
	tokenHash, err := s.refreshManager.Hash(ctx, refreshTokenStr)
	if err != nil {
		return err
	}
	refreshToken, err := s.refreshRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		if stdErrors.Is(err, pgx.ErrNoRows) {
			return appErrors.ErrInvalidCredentials
		}
		return err
	}
	return s.refreshRepo.RevokeByID(ctx, refreshToken.ID)
}
