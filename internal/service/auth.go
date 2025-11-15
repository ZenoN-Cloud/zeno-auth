package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/ZenoN-Cloud/zeno-auth/internal/model"
	"github.com/ZenoN-Cloud/zeno-auth/internal/repository"
	"github.com/ZenoN-Cloud/zeno-auth/internal/token"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
	passwordManager *token.PasswordManager
}

func NewAuthService(
	userRepo repository.UserRepository,
	orgRepo repository.OrganizationRepository,
	membershipRepo repository.MembershipRepository,
	refreshRepo repository.RefreshTokenRepository,
	jwtManager *token.JWTManager,
	refreshManager *token.RefreshManager,
	passwordManager *token.PasswordManager,
) *AuthService {
	return &AuthService{
		userRepo:        userRepo,
		orgRepo:         orgRepo,
		membershipRepo:  membershipRepo,
		refreshRepo:     refreshRepo,
		jwtManager:      jwtManager,
		refreshManager:  refreshManager,
		passwordManager: passwordManager,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password, fullName string) (*model.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	
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

	user := &model.User{
		Email:        email,
		PasswordHash: passwordHash,
		FullName:     fullName,
		IsActive:     true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
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

	valid, err := s.passwordManager.Verify(ctx, password, user.PasswordHash)
	if err != nil {
		return "", "", err
	}
	if !valid {
		return "", "", ErrInvalidCredentials
	}

	orgs, err := s.orgRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		return "", "", err
	}

	var orgID uuid.UUID
	var roles []string
	if len(orgs) > 0 {
		orgID = orgs[0].ID
		membership, err := s.membershipRepo.GetByUserAndOrg(ctx, user.ID, orgID)
		if err == nil {
			roles = []string{string(membership.Role)}
		}
	}

	accessToken, err := s.jwtManager.Generate(ctx, user.ID, orgID, roles)
	if err != nil {
		return "", "", err
	}

	refreshTokenStr, err := s.refreshManager.Generate(ctx)
	if err != nil {
		return "", "", err
	}

	refreshToken := s.refreshManager.CreateToken(ctx, user.ID, orgID, refreshTokenStr, userAgent, ipAddress)
	if err := s.refreshRepo.Create(ctx, refreshToken); err != nil {
		return "", "", err
	}

	return accessToken, refreshTokenStr, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshTokenStr string) (string, error) {
	tokenHash := s.refreshManager.Hash(ctx, refreshTokenStr)
	
	refreshToken, err := s.refreshRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return "", ErrInvalidCredentials
	}

	if refreshToken.RevokedAt != nil || refreshToken.ExpiresAt.Before(time.Now()) {
		return "", ErrInvalidCredentials
	}

	membership, err := s.membershipRepo.GetByUserAndOrg(ctx, refreshToken.UserID, refreshToken.OrgID)
	if err != nil {
		return "", err
	}

	roles := []string{string(membership.Role)}
	return s.jwtManager.Generate(ctx, refreshToken.UserID, refreshToken.OrgID, roles)
}

func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	return s.refreshRepo.RevokeByUserID(ctx, userID)
}