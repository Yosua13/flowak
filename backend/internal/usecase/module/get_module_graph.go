package module

import (
	"context"
	"fmt"
	"time"

	"backend/internal/domain/module"
	"backend/internal/repository/redis/cache"
	goredis "github.com/redis/go-redis/v9"
)

// ExecuteGetModuleGraph retrieves a module's canvas graph.
// It first checks Redis hot graph cache (cache:module:<id>:graph).
// On cache miss, it falls back to PostgreSQL and populates the Redis cache.
func ExecuteGetModuleGraph(
	ctx context.Context,
	repo module.ModuleRepository,
	redisClient *goredis.Client,
	moduleID string,
) (*module.ModuleGraph, error) {
	if moduleID == "" {
		return nil, fmt.Errorf("moduleID cannot be empty")
	}

	// 1. Check Redis cache
	cachedGraph, err := cache.GetGraphCache(ctx, redisClient, moduleID)
	if err == nil && cachedGraph != nil {
		return cachedGraph, nil
	}

	// 2. Fallback to PostgreSQL repository
	graph, err := repo.FindGraphByModuleID(ctx, moduleID)
	if err != nil {
		return nil, err
	}

	// 3. Populate Redis cache (5 minutes TTL)
	_ = cache.SetGraphCache(ctx, redisClient, moduleID, graph, 5*time.Minute)

	return graph, nil
}
