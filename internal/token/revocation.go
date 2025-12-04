package token

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RevocationService struct {
	redis *redis.Client
}

func NewRevocationService(redis *redis.Client) *RevocationService {
	return &RevocationService{redis: redis}
}

func (s *RevocationService) RevokeToken(ctx context.Context, jti string, ttl time.Duration) error {
	key := fmt.Sprintf("revoked:%s", jti)
	return s.redis.Set(ctx, key, "1", ttl).Err()
}

func (s *RevocationService) IsRevoked(ctx context.Context, jti string) (bool, error) {
	key := fmt.Sprintf("revoked:%s", jti)
	val, err := s.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return val == "1", nil
}
