package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"github.com/ZenoN-Cloud/zeno-auth/internal/model"
)

type RefreshTokenRepo struct {
	db *DB
}

func NewRefreshTokenRepo(db *DB) *RefreshTokenRepo {
	return &RefreshTokenRepo{db: db}
}

func (r *RefreshTokenRepo) Create(ctx context.Context, token *model.RefreshToken) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `
		INSERT INTO refresh_tokens (user_id, org_id, token_hash, user_agent, ip_address, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`

	return r.db.pool.QueryRow(ctx, query, token.UserID, token.OrgID, token.TokenHash, token.UserAgent, token.IPAddress, token.CreatedAt, token.ExpiresAt).Scan(&token.ID)
}

func (r *RefreshTokenRepo) CreateTx(ctx context.Context, tx *sql.Tx, token *model.RefreshToken) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `
		INSERT INTO refresh_tokens (user_id, org_id, token_hash, user_agent, ip_address, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`

	return tx.QueryRowContext(ctx, query, token.UserID, token.OrgID, token.TokenHash, token.UserAgent, token.IPAddress, token.CreatedAt, token.ExpiresAt).Scan(&token.ID)
}

func (r *RefreshTokenRepo) GetByTokenHash(ctx context.Context, tokenHash string) (*model.RefreshToken, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `SELECT id, user_id, org_id, token_hash, user_agent, ip_address, created_at, expires_at, revoked_at FROM refresh_tokens WHERE token_hash = $1`

	token := &model.RefreshToken{}
	err := r.db.pool.QueryRow(ctx, query, tokenHash).Scan(&token.ID, &token.UserID, &token.OrgID, &token.TokenHash, &token.UserAgent, &token.IPAddress, &token.CreatedAt, &token.ExpiresAt, &token.RevokedAt)
	return token, err
}

func (r *RefreshTokenRepo) RevokeByUserID(ctx context.Context, userID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `UPDATE refresh_tokens SET revoked_at = $2 WHERE user_id = $1 AND revoked_at IS NULL`

	_, err := r.db.pool.Exec(ctx, query, userID, time.Now())
	return err
}

func (r *RefreshTokenRepo) RevokeByUserIDTx(ctx context.Context, tx *sql.Tx, userID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `UPDATE refresh_tokens SET revoked_at = $2 WHERE user_id = $1 AND revoked_at IS NULL`

	_, err := tx.ExecContext(ctx, query, userID, time.Now())
	return err
}

func (r *RefreshTokenRepo) RevokeByID(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `UPDATE refresh_tokens SET revoked_at = $2 WHERE id = $1`

	_, err := r.db.pool.Exec(ctx, query, id, time.Now())
	return err
}

func (r *RefreshTokenRepo) DeleteExpired(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `DELETE FROM refresh_tokens WHERE expires_at < $1 OR revoked_at < $2`

	cutoff := time.Now().Add(-24 * time.Hour)
	_, err := r.db.pool.Exec(ctx, query, time.Now(), cutoff)
	return err
}

func (r *RefreshTokenRepo) GetActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*model.RefreshToken, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `
		SELECT id, user_id, org_id, token_hash, user_agent, ip_address, fingerprint_hash, created_at, expires_at, revoked_at
		FROM refresh_tokens
		WHERE user_id = $1 AND revoked_at IS NULL AND expires_at > NOW()
		ORDER BY created_at DESC`

	rows, err := r.db.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []*model.RefreshToken
	for rows.Next() {
		token := &model.RefreshToken{}
		err := rows.Scan(&token.ID, &token.UserID, &token.OrgID, &token.TokenHash, &token.UserAgent, &token.IPAddress, &token.FingerprintHash, &token.CreatedAt, &token.ExpiresAt, &token.RevokedAt)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}

	return tokens, rows.Err()
}
