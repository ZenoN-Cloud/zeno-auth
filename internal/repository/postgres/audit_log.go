package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ZenoN-Cloud/zeno-auth/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditLogRepository struct {
	db *pgxpool.Pool
}

func NewAuditLogRepository(db *pgxpool.Pool) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

func (r *AuditLogRepository) Create(ctx context.Context, log *model.AuditLog) error {
	query := `
		INSERT INTO audit_logs (user_id, event_type, event_data, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`

	var eventDataJSON []byte
	var err error
	if log.EventData != nil {
		eventDataJSON, err = json.Marshal(log.EventData)
		if err != nil {
			return err
		}
	}

	return r.db.QueryRow(
		ctx, query,
		log.UserID, log.EventType, eventDataJSON, log.IPAddress, log.UserAgent,
	).Scan(&log.ID, &log.CreatedAt)
}

func (r *AuditLogRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]*model.AuditLog, error) {
	query := `
		SELECT id, user_id, event_type, event_data, ip_address, user_agent, created_at
		FROM audit_logs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2`

	rows, err := r.db.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*model.AuditLog
	for rows.Next() {
		var log model.AuditLog
		var eventDataJSON []byte

		if err := rows.Scan(&log.ID, &log.UserID, &log.EventType, &eventDataJSON, &log.IPAddress, &log.UserAgent, &log.CreatedAt); err != nil {
			return nil, err
		}

		if len(eventDataJSON) > 0 {
			if err := json.Unmarshal(eventDataJSON, &log.EventData); err != nil {
				return nil, err
			}
		}

		logs = append(logs, &log)
	}

	return logs, rows.Err()
}

func (r *AuditLogRepository) DeleteOlderThan(ctx context.Context, date time.Time) error {
	query := `DELETE FROM audit_logs WHERE created_at < $1`
	_, err := r.db.Exec(ctx, query, date)
	return err
}

func (r *AuditLogRepository) AnonymizeByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE audit_logs SET user_id = NULL WHERE user_id = $1`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}
