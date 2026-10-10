package lock

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
)

const releaseLockScript = `
if redis.call("get", KEYS[1]) == ARGV[1] then
    return redis.call("del", KEYS[1])
else
    return 0
end
`

// ReleaseModuleLock releases the distributed lock if the lock ID matches the current owner.
func ReleaseModuleLock(ctx context.Context, client *goredis.Client, moduleID string, lockID string) error {
	if client == nil || moduleID == "" || lockID == "" || lockID == "mock_lock" {
		return nil
	}

	key := fmt.Sprintf("lock:module:%s:sync", moduleID)
	return client.Eval(ctx, releaseLockScript, []string{key}, lockID).Err()
}
