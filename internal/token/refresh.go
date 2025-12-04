package token

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/google/uuid"

	"github.com/ZenoN-Cloud/zeno-auth/internal/model"
)

type RefreshManager struct{}

func NewRefreshManager() *RefreshManager {
	return &RefreshManager{}
}

func (r *RefreshManager) Generate(ctx context.Context) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (r *RefreshManager) Hash(ctx context.Context, token string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:]), nil
}

func (r *RefreshManager) CreateToken(
	ctx context.Context,
	userID, orgID uuid.UUID,
	token, userAgent, ipAddress string,
	ttlSeconds int,
) (*model.RefreshToken, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	tokenHash, err := r.Hash(ctx, token)
	if err != nil {
		return nil, err
	}

	return &model.RefreshToken{
		UserID:    userID,
		OrgID:     orgID,
		TokenHash: tokenHash,
		UserAgent: userAgent,
		IPAddress: ipAddress,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Duration(ttlSeconds) * time.Second),
	}, nil
}
