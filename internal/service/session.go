package service

import (
	"context"

	"github.com/ZenoN-Cloud/zeno-auth/internal/model"
	"github.com/google/uuid"
)

type SessionService struct {
	refreshRepo RefreshTokenRepository
}

func NewSessionService(refreshRepo RefreshTokenRepository) *SessionService {
	return &SessionService{
		refreshRepo: refreshRepo,
	}
}

func (s *SessionService) GetActiveSessions(ctx context.Context, userID uuid.UUID) ([]*model.RefreshToken, error) {
	return s.refreshRepo.GetActiveByUserID(ctx, userID)
}

func (s *SessionService) RevokeSession(ctx context.Context, sessionID uuid.UUID) error {
	return s.refreshRepo.RevokeByID(ctx, sessionID)
}

func (s *SessionService) RevokeAllSessions(ctx context.Context, userID uuid.UUID) error {
	return s.refreshRepo.RevokeByUserID(ctx, userID)
}
