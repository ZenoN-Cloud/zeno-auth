package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/ZenoN-Cloud/zeno-auth/internal/model"
	"github.com/ZenoN-Cloud/zeno-auth/internal/repository"
	"github.com/ZenoN-Cloud/zeno-auth/internal/repository/postgres"
	"github.com/ZenoN-Cloud/zeno-auth/internal/token"
	"github.com/ZenoN-Cloud/zeno-auth/internal/validator"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotActive      = errors.New("user not active")
	ErrEmailExists        = errors.New("email already exists")
)

type AuthService struct {
	userRepo        repository.UserRepository
	orgRepo         repository.OrganizationRepository
	membershipRepo  repository.MembershipRepository
	refreshRepo     repository.RefreshTokenRepository
	jwtManager      *token.JWTManager
	refreshManager  *token.RefreshManager
	passwordManager token.PasswordHasher
	emailService    *EmailService
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
		config:          config,
		db:              db,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password, fullName string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	email = strings.ToLower(strings.TrimSpace(email))

	// Validate password strength
	passwordValidator := validator.NewPasswordValidator()
	if err := passwordValidator.Validate(password); err != nil {
		return nil, err
	}

	_, err := s.userRepo.GetByEmail(ctx, email)
	if err == nil {
		return nil, ErrEmailExists
	}
	if !errors.Is(err, pgx.ErrNoRows) {
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
	defer func() { _ = tx.Rollback() }()

	user := &model.User{
		Email:        email,
		PasswordHash: passwordHash,
		FullName:     fullName,
		IsActive:     true,
	}

	// Create user
	if err := s.userRepo.CreateTx(ctx, tx, user); err != nil {
		return nil, err
	}

	// Create default organization for user
	org := &model.Organization{
		Name:        fullName + "'s Organization",
		OwnerUserID: user.ID,
		Status:      "active",
	}

	if err := s.orgRepo.CreateTx(ctx, tx, org); err != nil {
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
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, err
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
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", ErrInvalidCredentials
		}
		return "", "", err
	}

	if !user.IsActive {
		return "", "", ErrUserNotActive
	}

	// Check if account is locked
	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		return "", "", errors.New("account is locked due to too many failed login attempts")
	}

	valid, err := s.passwordManager.Verify(ctx, password, user.PasswordHash)
	if err != nil {
		return "", "", err
	}
	if !valid {
		// Increment failed attempts
		user.FailedLoginAttempts++
		if user.FailedLoginAttempts >= 5 {
			lockUntil := time.Now().Add(30 * time.Minute)
			user.LockedUntil = &lockUntil
			// Notify user about account lockout
			if s.emailService != nil {
				go func() { _ = s.emailService.SendAccountLockoutNotification(ctx, user.ID, lockUntil) }()
			}
		}
		_ = s.userRepo.Update(ctx, user)
		return "", "", ErrInvalidCredentials
	}

	// Reset failed attempts on successful login
	if user.FailedLoginAttempts > 0 || user.LockedUntil != nil {
		user.FailedLoginAttempts = 0
		user.LockedUntil = nil
		_ = s.userRepo.Update(ctx, user)
	}

	// Get user's organizations
	orgs, err := s.orgRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		return "", "", err
	}

	// User must have at least one organization
	if len(orgs) == 0 {
		return "", "", errors.New("user has no organizations")
	}

	// Use first organization (later can add org selection)
	orgID := orgs[0].ID

	// Get user's role in organization
	var roles []string
	membership, err := s.membershipRepo.GetByUserAndOrg(ctx, user.ID, orgID)
	if err == nil {
		roles = []string{string(membership.Role)}
	}

	accessToken, err := s.jwtManager.Generate(ctx, user.ID, orgID, roles, s.config.AccessTokenTTL)
	if err != nil {
		return "", "", err
	}

	refreshTokenStr, err := s.refreshManager.Generate(ctx)
	if err != nil {
		return "", "", err
	}

	// Generate fingerprint for session security
	fingerprint := token.GenerateFingerprint(userAgent, ipAddress, "")
	refreshToken := s.refreshManager.CreateToken(ctx, user.ID, orgID, refreshTokenStr, userAgent, ipAddress, s.config.RefreshTokenTTL)
	refreshToken.FingerprintHash = &fingerprint
	if err := s.refreshRepo.Create(ctx, refreshToken); err != nil {
		return "", "", err
	}

	return accessToken, refreshTokenStr, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshTokenStr, userAgent, ipAddress string) (string, error) {
	tokenHash := s.refreshManager.Hash(ctx, refreshTokenStr)

	refreshToken, err := s.refreshRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return "", ErrInvalidCredentials
	}

	if refreshToken.RevokedAt != nil || refreshToken.ExpiresAt.Before(time.Now()) {
		return "", ErrInvalidCredentials
	}

	// Validate session fingerprint to prevent session hijacking
	if refreshToken.FingerprintHash != nil && *refreshToken.FingerprintHash != "" {
		currentFingerprint := token.GenerateFingerprint(userAgent, ipAddress, "")
		if currentFingerprint != *refreshToken.FingerprintHash {
			return "", errors.New("session fingerprint mismatch - possible session hijacking")
		}
	}

	// Get roles if user has organization membership
	var roles []string
	if refreshToken.OrgID != uuid.Nil {
		membership, err := s.membershipRepo.GetByUserAndOrg(ctx, refreshToken.UserID, refreshToken.OrgID)
		if err == nil {
			roles = []string{string(membership.Role)}
		}
	}
	return s.jwtManager.Generate(ctx, refreshToken.UserID, refreshToken.OrgID, roles, s.config.AccessTokenTTL)
}

func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	return s.refreshRepo.RevokeByUserID(ctx, userID)
}
