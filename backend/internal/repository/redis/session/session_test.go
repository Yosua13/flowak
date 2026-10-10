package session_test

import (
	"context"
	"testing"
	"time"

	"backend/internal/repository/redis/session"
)

func TestSession_NilClientGraceful(t *testing.T) {
	ctx := context.Background()
	repo := session.NewBlacklistTokenRepo(nil)

	// Blacklist with nil client should not panic
	if err := repo.BlacklistToken(ctx, "test_token_123", time.Minute); err != nil {
		t.Errorf("expected nil error with nil client, got: %v", err)
	}

	// Check blacklisted with nil client should return false, nil
	blacklisted, err := repo.IsTokenBlacklisted(ctx, "test_token_123")
	if err != nil {
		t.Errorf("expected nil error with nil client, got: %v", err)
	}
	if blacklisted {
		t.Errorf("expected false for blacklisted with nil client")
	}

	// Function helpers with nil client
	if err := session.BlacklistToken(ctx, nil, "token_direct", time.Minute); err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}
	blacklisted, err = session.IsTokenBlacklisted(ctx, nil, "token_direct")
	if err != nil || blacklisted {
		t.Errorf("expected false and nil error, got %v, %v", blacklisted, err)
	}
}

func TestSession_EmptyTokenID(t *testing.T) {
	ctx := context.Background()
	repo := session.NewBlacklistTokenRepo(nil)

	if err := repo.BlacklistToken(ctx, "", time.Minute); err != nil {
		t.Errorf("expected nil error for empty token ID, got: %v", err)
	}

	blacklisted, err := repo.IsTokenBlacklisted(ctx, "")
	if err != nil || blacklisted {
		t.Errorf("expected false and nil error for empty token ID, got %v, %v", blacklisted, err)
	}
}
