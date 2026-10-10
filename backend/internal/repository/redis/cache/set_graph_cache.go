package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"backend/internal/domain/module"
	goredis "github.com/redis/go-redis/v9"
)

// SetGraphCache saves the module graph structure to Redis hot graph cache with a specified TTL.
func SetGraphCache(ctx context.Context, client *goredis.Client, moduleID string, graph *module.ModuleGraph, ttl time.Duration) error {
	if client == nil || moduleID == "" || graph == nil {
		return nil
	}

	data, err := json.Marshal(graph)
	if err != nil {
		return fmt.Errorf("failed to marshal graph for cache: %w", err)
	}

	key := fmt.Sprintf("cache:module:%s:graph", moduleID)
	if err := client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set graph cache: %w", err)
	}

	return nil
}
