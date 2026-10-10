package auth

import (
	"context"
	"time"
)

// TokenBlacklistRepository defines the contract for checking and revoking JWT tokens.
type TokenBlacklistRepository interface {
	BlacklistToken(ctx context.Context, tokenID string, ttl time.Duration) error
	IsTokenBlacklisted(ctx context.Context, tokenID string) (bool, error)
}
