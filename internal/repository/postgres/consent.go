package postgres

import (
	"context"
	"time"

	"github.com/ZenoN-Cloud/zeno-auth/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ConsentRepository struct {
	db *pgxpool.Pool
}

func NewConsentRepository(db *pgxpool.Pool) *ConsentRepository {
	return &ConsentRepository{db: db}
}

func (r *ConsentRepository) Create(ctx context.Context, consent *model.UserConsent) error {
	query := `
		INSERT INTO user_consents (user_id, consent_type, version, granted, granted_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	return r.db.QueryRow(
		ctx, query,
		consent.UserID, consent.ConsentType, consent.Version, consent.Granted, consent.GrantedAt,
	).Scan(&consent.ID, &consent.CreatedAt, &consent.UpdatedAt)
}

func (r *ConsentRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*model.UserConsent, error) {
	query := `
		SELECT id, user_id, consent_type, version, granted, granted_at, revoked_at, created_at, updated_at
		FROM user_consents
		WHERE user_id = $1 AND revoked_at IS NULL
		ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var consents []*model.UserConsent
	for rows.Next() {
		var c model.UserConsent
		if err := rows.Scan(&c.ID, &c.UserID, &c.ConsentType, &c.Version, &c.Granted, &c.GrantedAt, &c.RevokedAt, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		consents = append(consents, &c)
	}

	return consents, rows.Err()
}

func (r *ConsentRepository) GetByUserAndType(ctx context.Context, userID uuid.UUID, consentType model.ConsentType) (*model.UserConsent, error) {
	query := `
		SELECT id, user_id, consent_type, version, granted, granted_at, revoked_at, created_at, updated_at
		FROM user_consents
		WHERE user_id = $1 AND consent_type = $2 AND revoked_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1`

	var consent model.UserConsent
	err := r.db.QueryRow(ctx, query, userID, consentType).Scan(
		&consent.ID, &consent.UserID, &consent.ConsentType, &consent.Version,
		&consent.Granted, &consent.GrantedAt, &consent.RevokedAt,
		&consent.CreatedAt, &consent.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &consent, nil
}

func (r *ConsentRepository) Revoke(ctx context.Context, userID uuid.UUID, consentType model.ConsentType) error {
	query := `
		UPDATE user_consents
		SET revoked_at = $1, updated_at = $1
		WHERE user_id = $2 AND consent_type = $3 AND revoked_at IS NULL`

	_, err := r.db.Exec(ctx, query, time.Now(), userID, consentType)
	return err
}
