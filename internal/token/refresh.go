package token

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/ZenoN-Cloud/zeno-auth/internal/model"
	"github.com/google/uuid"
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

func (r *RefreshManager) Hash(ctx context.Context, token string) string {
	select {
	case <-ctx.Done():
		return ""
	default:
	}

	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func (r *RefreshManager) CreateToken(ctx context.Context, userID, orgID uuid.UUID, token, userAgent, ipAddress string) *model.RefreshToken {
	select {
	case <-ctx.Done():
		return nil
	default:
	}

	return &model.RefreshToken{
		UserID:    userID,
		OrgID:     orgID,
		TokenHash: r.Hash(ctx, token),
		UserAgent: userAgent,
		IPAddress: ipAddress,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour), // 30 days
	}
}