package config_test

import (
	"os"
	"testing"

	"backend/config"
)

func TestInitConfig_RedisAndRabbitMQDefaults(t *testing.T) {
	// Clear relevant environment variables
	os.Unsetenv("REDIS_ADDR")
	os.Unsetenv("REDIS_PASSWORD")
	os.Unsetenv("REDIS_DB")
	os.Unsetenv("RABBITMQ_URL")

	config.InitConfig()

	if config.ActiveConfig.RedisAddr != "localhost:6379" {
		t.Errorf("expected RedisAddr default to be localhost:6379, got: %s", config.ActiveConfig.RedisAddr)
	}

	if config.ActiveConfig.RedisPassword != "" {
		t.Errorf("expected RedisPassword default to be empty, got: %s", config.ActiveConfig.RedisPassword)
	}

	if config.ActiveConfig.RedisDB != 0 {
		t.Errorf("expected RedisDB default to be 0, got: %d", config.ActiveConfig.RedisDB)
	}

	if config.ActiveConfig.RabbitMQURL != "amqp://guest:guest@localhost:5672/" {
		t.Errorf("expected RabbitMQURL default to be amqp://guest:guest@localhost:5672/, got: %s", config.ActiveConfig.RabbitMQURL)
	}
}

func TestInitConfig_CustomRedisAndRabbitMQ(t *testing.T) {
	os.Setenv("REDIS_ADDR", "redis.internal:6380")
	os.Setenv("REDIS_PASSWORD", "secret123")
	os.Setenv("REDIS_DB", "3")
	os.Setenv("RABBITMQ_URL", "amqp://admin:pass@rmq.internal:5672/vhost")

	defer func() {
		os.Unsetenv("REDIS_ADDR")
		os.Unsetenv("REDIS_PASSWORD")
		os.Unsetenv("REDIS_DB")
		os.Unsetenv("RABBITMQ_URL")
		config.InitConfig()
	}()

	config.InitConfig()

	if config.ActiveConfig.RedisAddr != "redis.internal:6380" {
		t.Errorf("expected RedisAddr to be redis.internal:6380, got: %s", config.ActiveConfig.RedisAddr)
	}

	if config.ActiveConfig.RedisPassword != "secret123" {
		t.Errorf("expected RedisPassword to be secret123, got: %s", config.ActiveConfig.RedisPassword)
	}

	if config.ActiveConfig.RedisDB != 3 {
		t.Errorf("expected RedisDB to be 3, got: %d", config.ActiveConfig.RedisDB)
	}

	if config.ActiveConfig.RabbitMQURL != "amqp://admin:pass@rmq.internal:5672/vhost" {
		t.Errorf("expected RabbitMQURL to be amqp://admin:pass@rmq.internal:5672/vhost, got: %s", config.ActiveConfig.RabbitMQURL)
	}
}
