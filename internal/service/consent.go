package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ZenoN-Cloud/zeno-auth/internal/model"
	"github.com/google/uuid"
)

var (
	ErrInvalidConsentType = errors.New("invalid consent type")
	ErrEmptyVersion       = errors.New("version cannot be empty")
)

type ConsentRepository interface {
	Create(ctx context.Context, consent *model.UserConsent) error
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*model.UserConsent, error)
	GetByUserAndType(ctx context.Context, userID uuid.UUID, consentType model.ConsentType) (*model.UserConsent, error)
	Revoke(ctx context.Context, userID uuid.UUID, consentType model.ConsentType) error
}

type ConsentService struct {
	consentRepo ConsentRepository
}

func NewConsentService(consentRepo ConsentRepository) *ConsentService {
	return &ConsentService{
		consentRepo: consentRepo,
	}
}

func (s *ConsentService) GrantConsent(ctx context.Context, userID uuid.UUID, consentType model.ConsentType, version string) error {
	if userID == uuid.Nil {
		return ErrInvalidUserID
	}
	if consentType == "" {
		return ErrInvalidConsentType
	}
	if version == "" {
		return ErrEmptyVersion
	}

	existing, err := s.consentRepo.GetByUserAndType(ctx, userID, consentType)
	if err != nil {
		return fmt.Errorf("failed to check existing consent: %w", err)
	}

	if existing != nil && existing.Version == version {
		return nil
	}

	if existing != nil {
		if err := s.consentRepo.Revoke(ctx, userID, consentType); err != nil {
			return fmt.Errorf("failed to revoke old consent: %w", err)
		}
	}

	consent := &model.UserConsent{
		UserID:      userID,
		ConsentType: consentType,
		Version:     version,
		Granted:     true,
		GrantedAt:   time.Now(),
	}

	if err := s.consentRepo.Create(ctx, consent); err != nil {
		return fmt.Errorf("failed to create consent: %w", err)
	}

	return nil
}

func (s *ConsentService) RevokeConsent(ctx context.Context, userID uuid.UUID, consentType model.ConsentType) error {
	if userID == uuid.Nil {
		return ErrInvalidUserID
	}
	if consentType == "" {
		return ErrInvalidConsentType
	}

	if err := s.consentRepo.Revoke(ctx, userID, consentType); err != nil {
		return fmt.Errorf("failed to revoke consent: %w", err)
	}
	return nil
}

func (s *ConsentService) GetUserConsents(ctx context.Context, userID uuid.UUID) ([]*model.UserConsent, error) {
	consents, err := s.consentRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user consents: %w", err)
	}
	return consents, nil
}

func (s *ConsentService) HasConsent(ctx context.Context, userID uuid.UUID, consentType model.ConsentType) (bool, error) {
	consent, err := s.consentRepo.GetByUserAndType(ctx, userID, consentType)
	if err != nil {
		return false, fmt.Errorf("failed to check consent: %w", err)
	}
	return consent != nil && consent.Granted, nil
}
