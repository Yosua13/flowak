package consumer_test

import (
	"context"
	"testing"
	"time"

	"backend/internal/infra/rabbitmq/consumer"
	"backend/internal/infra/rabbitmq/producer"
)

func TestConsumeAPIRunnerJobs_NilChannel(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := consumer.ConsumeAPIRunnerJobs(ctx, nil, func(ctx context.Context, job producer.APIRunnerJobPayload) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected error on nil channel, got nil")
	}
}
