package service

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
)

type CleanupService struct {
	refreshTokenRepo RefreshTokenRepository
	auditRepo        AuditLogRepository
}

func NewCleanupService(refreshTokenRepo RefreshTokenRepository, auditRepo AuditLogRepository) *CleanupService {
	return &CleanupService{
		refreshTokenRepo: refreshTokenRepo,
		auditRepo:        auditRepo,
	}
}

func (s *CleanupService) CleanupExpiredTokens(ctx context.Context) (int, error) {
	log.Info().Msg("Starting cleanup of expired refresh tokens")

	if err := s.refreshTokenRepo.DeleteExpired(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to cleanup expired tokens")
		return 0, err
	}

	log.Info().Msg("Expired tokens cleanup completed")
	return 0, nil
}

func (s *CleanupService) CleanupOldAuditLogs(ctx context.Context, retentionDays int) (int, error) {
	log.Info().Int("retention_days", retentionDays).Msg("Starting cleanup of old audit logs")

	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)
	if err := s.auditRepo.DeleteOlderThan(ctx, cutoffDate); err != nil {
		log.Error().Err(err).Msg("Failed to cleanup old audit logs")
		return 0, err
	}

	log.Info().Msg("Old audit logs cleanup completed")
	return 0, nil
}
