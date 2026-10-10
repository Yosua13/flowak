package session

import (
	"context"

	goredis "github.com/redis/go-redis/v9"
)

// IsTokenBlacklisted checks whether the token ID exists in the Redis blacklist.
func (r *BlacklistTokenRepo) IsTokenBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	if r.client == nil || tokenID == "" {
		return false, nil
	}
	key := "blacklist:token:" + tokenID
	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

// IsTokenBlacklistedFunc is a direct function helper for checking if a token is blacklisted.
func IsTokenBlacklisted(ctx context.Context, client *goredis.Client, tokenID string) (bool, error) {
	repo := NewBlacklistTokenRepo(client)
	return repo.IsTokenBlacklisted(ctx, tokenID)
}
