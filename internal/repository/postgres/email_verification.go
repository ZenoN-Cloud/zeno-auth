package postgres

import (
	"context"

	"github.com/ZenoN-Cloud/zeno-auth/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EmailVerificationRepository struct {
	db *pgxpool.Pool
}

func NewEmailVerificationRepository(db *pgxpool.Pool) *EmailVerificationRepository {
	return &EmailVerificationRepository{db: db}
}

func (r *EmailVerificationRepository) Create(ctx context.Context, verification *model.EmailVerification) error {
	query := `
		INSERT INTO email_verifications (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`
	return r.db.QueryRow(ctx, query, verification.UserID, verification.TokenHash, verification.ExpiresAt).
		Scan(&verification.ID, &verification.CreatedAt)
}

func (r *EmailVerificationRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*model.EmailVerification, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, verified_at, created_at
		FROM email_verifications
		WHERE token_hash = $1`

	var verification model.EmailVerification
	err := r.db.QueryRow(ctx, query, tokenHash).Scan(
		&verification.ID,
		&verification.UserID,
		&verification.TokenHash,
		&verification.ExpiresAt,
		&verification.VerifiedAt,
		&verification.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &verification, nil
}

func (r *EmailVerificationRepository) MarkAsVerified(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE email_verifications SET verified_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *EmailVerificationRepository) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM email_verifications WHERE expires_at < NOW() - INTERVAL '7 days'`
	_, err := r.db.Exec(ctx, query)
	return err
}

func (r *EmailVerificationRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM email_verifications WHERE user_id = $1`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}
