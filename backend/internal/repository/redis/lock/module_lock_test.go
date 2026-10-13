package lock_test

import (
	"context"
	"testing"
	"time"

	"backend/internal/repository/redis/lock"
)

func TestModuleLock_NilClientGraceful(t *testing.T) {
	ctx := context.Background()

	// Acquire with nil client should return mock token and nil error
	token, err := lock.AcquireModuleLock(ctx, nil, "mod_123", 5*time.Second)
	if err != nil || token != "mock_lock" {
		t.Fatalf("expected mock_lock and nil error, got token: %q, err: %v", token, err)
	}

	// Release with nil client should return nil error
	err = lock.ReleaseModuleLock(ctx, nil, "mod_123", token)
	if err != nil {
		t.Fatalf("expected nil error for ReleaseModuleLock, got: %v", err)
	}
}

func TestModuleLock_EmptyModuleID(t *testing.T) {
	ctx := context.Background()

	_, err := lock.AcquireModuleLock(ctx, nil, "", 5*time.Second)
	if err == nil {
		t.Fatalf("expected error for empty moduleID")
	}

	err = lock.ReleaseModuleLock(ctx, nil, "", "any_token")
	if err != nil {
		t.Fatalf("expected nil error for empty moduleID on release, got: %v", err)
	}
}
