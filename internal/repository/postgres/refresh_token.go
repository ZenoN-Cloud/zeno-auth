package postgres

import (
	"context"
	"time"

	"github.com/ZenoN-Cloud/zeno-auth/internal/model"
	"github.com/google/uuid"
)

type RefreshTokenRepo struct {
	db *DB
}

func NewRefreshTokenRepo(db *DB) *RefreshTokenRepo {
	return &RefreshTokenRepo{db: db}
}

func (r *RefreshTokenRepo) Create(ctx context.Context, token *model.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (user_id, org_id, token_hash, user_agent, ip_address, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`

	var orgID interface{}
	if token.OrgID == uuid.Nil {
		orgID = nil
	} else {
		orgID = token.OrgID
	}

	return r.db.pool.QueryRow(ctx, query, token.UserID, orgID, token.TokenHash, token.UserAgent, token.IPAddress, token.CreatedAt, token.ExpiresAt).Scan(&token.ID)
}

func (r *RefreshTokenRepo) GetByTokenHash(ctx context.Context, tokenHash string) (*model.RefreshToken, error) {
	query := `SELECT id, user_id, org_id, token_hash, user_agent, ip_address, created_at, expires_at, revoked_at FROM refresh_tokens WHERE token_hash = $1`

	token := &model.RefreshToken{}
	err := r.db.pool.QueryRow(ctx, query, tokenHash).Scan(&token.ID, &token.UserID, &token.OrgID, &token.TokenHash, &token.UserAgent, &token.IPAddress, &token.CreatedAt, &token.ExpiresAt, &token.RevokedAt)
	return token, err
}

func (r *RefreshTokenRepo) RevokeByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE refresh_tokens SET revoked_at = $2 WHERE user_id = $1 AND revoked_at IS NULL`

	_, err := r.db.pool.Exec(ctx, query, userID, time.Now())
	return err
}

func (r *RefreshTokenRepo) RevokeByID(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE refresh_tokens SET revoked_at = $2 WHERE id = $1`

	_, err := r.db.pool.Exec(ctx, query, id, time.Now())
	return err
}

func (r *RefreshTokenRepo) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM refresh_tokens WHERE expires_at < $1 OR revoked_at < $2`

	cutoff := time.Now().Add(-24 * time.Hour)
	_, err := r.db.pool.Exec(ctx, query, time.Now(), cutoff)
	return err
}
