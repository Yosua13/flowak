package auth

import (
	"context"
	"time"

	domainAuth "backend/internal/domain/auth"
)

// LogoutInput contains the session tokens to revoke.
type LogoutInput struct {
	TokenID      string
	RefreshToken string
	RemainingTTL time.Duration
}

// LogoutUserUseCase handles user logout, session revocation, and Redis token blacklisting.
type LogoutUserUseCase struct {
	repo          domainAuth.AuthRepository
	blacklistRepo domainAuth.TokenBlacklistRepository
}

// NewLogoutUserUseCase creates a new LogoutUserUseCase.
func NewLogoutUserUseCase(repo domainAuth.AuthRepository, blacklistRepo domainAuth.TokenBlacklistRepository) *LogoutUserUseCase {
	return &LogoutUserUseCase{
		repo:          repo,
		blacklistRepo: blacklistRepo,
	}
}

// Execute revokes the PostgreSQL session and blacklists the JWT in Redis.
func (uc *LogoutUserUseCase) Execute(ctx context.Context, input LogoutInput) error {
	if input.RefreshToken != "" && uc.repo != nil {
		_ = uc.repo.RevokeSessionByRefreshToken(ctx, TokenHash(input.RefreshToken))
	}

	ttl := input.RemainingTTL
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}

	if input.TokenID != "" && uc.blacklistRepo != nil {
		_ = uc.blacklistRepo.BlacklistToken(ctx, input.TokenID, ttl)
	}

	return nil
}
