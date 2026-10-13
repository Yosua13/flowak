package redis_test

import (
	"errors"
	"testing"

	"backend/config"
	"backend/internal/repository/redis"
)

func TestNewRedisClient_NilConfig(t *testing.T) {
	client, err := redis.NewRedisClient(nil)
	if err == nil {
		t.Fatal("expected error when cfg is nil, got nil")
	}
	if client != nil {
		t.Fatal("expected nil client when cfg is nil")
	}
}

func TestNewRedisClient_Offline_GracefulFallback(t *testing.T) {
	// Point to an unused local port to simulate offline Redis
	cfg := &config.Config{
		RedisAddr:     "127.0.0.1:63799",
		RedisPassword: "",
		RedisDB:        0,
	}

	client, err := redis.NewRedisClient(cfg)
	// Graceful fallback requires returning the initialized client instance
	// so the application can degrade gracefully without panicking.
	if client == nil {
		t.Fatal("expected non-nil client for graceful fallback")
	}
	defer client.Close()

	if err == nil {
		t.Fatal("expected ping error when redis is offline, got nil")
	}

	if !errors.Is(err, redis.ErrRedisUnavailable) {
		t.Fatalf("expected error to wrap ErrRedisUnavailable, got: %v", err)
	}
}

func TestNewRedisClient_DefaultAddress_GracefulOrSuccess(t *testing.T) {
	cfg := &config.Config{
		RedisAddr:     "localhost:6379",
		RedisPassword: "",
		RedisDB:        0,
	}

	client, err := redis.NewRedisClient(cfg)
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	defer client.Close()

	// If Redis is running on localhost, err should be nil.
	// If Redis is offline, err should wrap ErrRedisUnavailable.
	if err != nil && !errors.Is(err, redis.ErrRedisUnavailable) {
		t.Fatalf("unexpected error format: %v", err)
	}
}
