package consumer_test

import (
	"context"
	"testing"
	"time"

	"backend/internal/infra/rabbitmq/consumer"
)

func TestConsumeSSEBroadcast_NilChannel(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := consumer.ConsumeSSEBroadcast(ctx, nil, func(projectID string, event []byte) {})
	if err == nil {
		t.Fatal("expected error when amqp channel is nil, got nil")
	}
}
