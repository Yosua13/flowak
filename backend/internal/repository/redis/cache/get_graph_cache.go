package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"backend/internal/domain/module"
	goredis "github.com/redis/go-redis/v9"
)

// GetGraphCache retrieves cached module graph JSON from Redis.
// Returns nil, nil on cache miss or when Redis client is unconfigured.
func GetGraphCache(ctx context.Context, client *goredis.Client, moduleID string) (*module.ModuleGraph, error) {
	if client == nil || moduleID == "" {
		return nil, nil
	}

	key := fmt.Sprintf("cache:module:%s:graph", moduleID)
	data, err := client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get graph from cache: %w", err)
	}

	var graph module.ModuleGraph
	if err := json.Unmarshal(data, &graph); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached graph: %w", err)
	}

	return &graph, nil
}
