package cache

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
)

// InvalidateGraphCache deletes the module graph entry from Redis hot cache upon mutations.
func InvalidateGraphCache(ctx context.Context, client *goredis.Client, moduleID string) error {
	if client == nil || moduleID == "" {
		return nil
	}

	key := fmt.Sprintf("cache:module:%s:graph", moduleID)
	if err := client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to invalidate graph cache: %w", err)
	}

	return nil
}
