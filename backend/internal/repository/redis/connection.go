package redis

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"backend/config"
	goredis "github.com/redis/go-redis/v9"
)

// ErrRedisUnavailable is returned when initial ping check fails, indicating degraded mode.
var ErrRedisUnavailable = errors.New("redis service is unavailable")

// NewRedisClient initializes a go-redis client and performs an initial ping check.
// If the ping fails, it returns the initialized client alongside ErrRedisUnavailable,
// allowing graceful degradation (caller can proceed with caching disabled or wait for reconnect).
func NewRedisClient(cfg *config.Config) (*goredis.Client, error) {
	if cfg == nil {
		return nil, errors.New("configuration cannot be nil")
	}

	addr := cfg.RedisAddr
	if addr == "" {
		addr = "localhost:6379"
	}

	client := goredis.NewClient(&goredis.Options{
		Addr:     addr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("[Redis] Warning: connection ping to %s failed: %v. Operating in graceful fallback mode.\n", addr, err)
		return client, fmt.Errorf("%w: %v", ErrRedisUnavailable, err)
	}

	log.Printf("[Redis] Successfully connected to %s (DB %d)\n", addr, cfg.RedisDB)
	return client, nil
}
