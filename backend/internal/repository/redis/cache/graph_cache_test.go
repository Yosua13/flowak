package cache_test

import (
	"context"
	"testing"
	"time"

	"backend/internal/domain/module"
	"backend/internal/repository/redis/cache"
)

func TestGraphCache_NilClientGraceful(t *testing.T) {
	ctx := context.Background()

	// Get with nil client should return nil, nil
	g, err := cache.GetGraphCache(ctx, nil, "mod_123")
	if err != nil || g != nil {
		t.Fatalf("expected nil, nil for nil client, got: %v, %v", g, err)
	}

	// Set with nil client should return nil
	err = cache.SetGraphCache(ctx, nil, "mod_123", &module.ModuleGraph{ID: "mod_123"}, time.Minute)
	if err != nil {
		t.Fatalf("expected nil for SetGraphCache with nil client, got: %v", err)
	}

	// Invalidate with nil client should return nil
	err = cache.InvalidateGraphCache(ctx, nil, "mod_123")
	if err != nil {
		t.Fatalf("expected nil for InvalidateGraphCache with nil client, got: %v", err)
	}
}

func TestGraphCache_EmptyModuleID(t *testing.T) {
	ctx := context.Background()

	g, err := cache.GetGraphCache(ctx, nil, "")
	if err != nil || g != nil {
		t.Fatalf("expected nil, nil for empty moduleID, got: %v, %v", g, err)
	}

	err = cache.SetGraphCache(ctx, nil, "", &module.ModuleGraph{}, time.Minute)
	if err != nil {
		t.Fatalf("expected nil for empty moduleID, got: %v", err)
	}

	err = cache.InvalidateGraphCache(ctx, nil, "")
	if err != nil {
		t.Fatalf("expected nil for empty moduleID, got: %v", err)
	}
}
