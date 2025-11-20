package service

import (
	"context"
	"time"

	"github.com/ZenoN-Cloud/zeno-auth/internal/model"
	"github.com/google/uuid"
)

type AuditLogRepository interface {
	Create(ctx context.Context, log *model.AuditLog) error
	GetByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]*model.AuditLog, error)
	DeleteOlderThan(ctx context.Context, date time.Time) error
	AnonymizeByUserID(ctx context.Context, userID uuid.UUID) error
}

type AuditService struct {
	auditRepo AuditLogRepository
}

func NewAuditService(auditRepo AuditLogRepository) *AuditService {
	return &AuditService{
		auditRepo: auditRepo,
	}
}

func (s *AuditService) Log(ctx context.Context, userID *uuid.UUID, eventType interface{}, eventData map[string]interface{}, ipAddress, userAgent string) error {
	var auditEventType model.AuditEventType
	switch v := eventType.(type) {
	case model.AuditEventType:
		auditEventType = v
	case string:
		auditEventType = model.AuditEventType(v)
	default:
		auditEventType = model.AuditEventType("unknown")
	}

	log := &model.AuditLog{
		UserID:    userID,
		EventType: auditEventType,
		EventData: eventData,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	return s.auditRepo.Create(ctx, log)
}

func (s *AuditService) GetUserLogs(ctx context.Context, userID uuid.UUID, limit int) ([]*model.AuditLog, error) {
	if limit <= 0 {
		limit = 100
	}
	return s.auditRepo.GetByUserID(ctx, userID, limit)
}

func (s *AuditService) CleanupOldLogs(ctx context.Context, retentionDays int) error {
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)
	return s.auditRepo.DeleteOlderThan(ctx, cutoffDate)
}
