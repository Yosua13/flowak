package lock

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"backend/internal/domain/module"
	goredis "github.com/redis/go-redis/v9"
)

// AcquireModuleLock acquires a distributed lock for syncing a module graph.
// Key format: lock:module:<id>:sync with specified TTL (e.g. 5 seconds).
// Returns a unique lock ID string that must be passed to ReleaseModuleLock.
func AcquireModuleLock(ctx context.Context, client *goredis.Client, moduleID string, ttl time.Duration) (string, error) {
	if moduleID == "" {
		return "", fmt.Errorf("moduleID cannot be empty")
	}
	if client == nil {
		return "mock_lock", nil
	}

	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate lock token: %w", err)
	}
	lockID := hex.EncodeToString(b)

	key := fmt.Sprintf("lock:module:%s:sync", moduleID)
	ok, err := client.SetNX(ctx, key, lockID, ttl).Result()
	if err != nil {
		return "", fmt.Errorf("redis lock set failed: %w", err)
	}
	if !ok {
		return "", module.ErrLockAcquisitionFailed
	}

	return lockID, nil
}
