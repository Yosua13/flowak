package session

import (
	"context"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// BlacklistTokenRepo implements auth.TokenBlacklistRepository backed by Redis.
type BlacklistTokenRepo struct {
	client *goredis.Client
}

// NewBlacklistTokenRepo creates a new BlacklistTokenRepo.
func NewBlacklistTokenRepo(client *goredis.Client) *BlacklistTokenRepo {
	return &BlacklistTokenRepo{client: client}
}

// BlacklistToken stores the token ID into Redis with a TTL matching the remaining token lifetime.
func (r *BlacklistTokenRepo) BlacklistToken(ctx context.Context, tokenID string, remainingTTL time.Duration) error {
	if r.client == nil || tokenID == "" {
		return nil
	}
	key := "blacklist:token:" + tokenID
	return r.client.Set(ctx, key, "revoked", remainingTTL).Err()
}

// BlacklistTokenFunc is a direct function helper for blacklisting a token.
func BlacklistToken(ctx context.Context, client *goredis.Client, tokenID string, remainingTTL time.Duration) error {
	repo := NewBlacklistTokenRepo(client)
	return repo.BlacklistToken(ctx, tokenID, remainingTTL)
}
