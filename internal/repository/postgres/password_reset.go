package postgres

import (
	"context"
	"fmt"

	"github.com/ZenoN-Cloud/zeno-auth/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PasswordResetRepository struct {
	db        *pgxpool.Pool
	dbWrapper *DB
}

func NewPasswordResetRepository(db *pgxpool.Pool) *PasswordResetRepository {
	return &PasswordResetRepository{db: db}
}

func NewPasswordResetRepositoryWithDB(dbWrapper *DB) *PasswordResetRepository {
	return &PasswordResetRepository{db: dbWrapper.Pool(), dbWrapper: dbWrapper}
}

func (r *PasswordResetRepository) Create(ctx context.Context, token *model.PasswordResetToken) error {
	query := `
		INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`
	return r.db.QueryRow(ctx, query, token.UserID, token.TokenHash, token.ExpiresAt).
		Scan(&token.ID, &token.CreatedAt)
}

func (r *PasswordResetRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*model.PasswordResetToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, used_at, created_at
		FROM password_reset_tokens
		WHERE token_hash = $1`

	var token model.PasswordResetToken
	err := r.db.QueryRow(ctx, query, tokenHash).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.UsedAt,
		&token.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *PasswordResetRepository) MarkAsUsed(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE password_reset_tokens SET used_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *PasswordResetRepository) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM password_reset_tokens WHERE expires_at < NOW() - INTERVAL '7 days'`
	_, err := r.db.Exec(ctx, query)
	return err
}

func (r *PasswordResetRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM password_reset_tokens WHERE user_id = $1`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}

func (r *PasswordResetRepository) ResetPasswordTx(ctx context.Context, user *model.User, tokenID uuid.UUID) error {
	if r.dbWrapper == nil {
		return fmt.Errorf("transaction support not available")
	}

	tx, err := r.dbWrapper.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// Update password
	updateQuery := `UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2`
	if _, err = tx.Exec(ctx, updateQuery, user.PasswordHash, user.ID); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Mark token as used
	tokenQuery := `UPDATE password_reset_tokens SET used_at = NOW() WHERE id = $1`
	if _, err = tx.Exec(ctx, tokenQuery, tokenID); err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	// Revoke all refresh tokens
	revokeQuery := `UPDATE refresh_tokens SET revoked_at = NOW() WHERE user_id = $1 AND revoked_at IS NULL`
	if _, err = tx.Exec(ctx, revokeQuery, user.ID); err != nil {
		return fmt.Errorf("failed to revoke tokens: %w", err)
	}

	return tx.Commit(ctx)
}
