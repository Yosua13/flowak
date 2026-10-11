package producer_test

import (
	"context"
	"testing"
	"time"

	"backend/internal/infra/rabbitmq/producer"
)

func TestAPIRunnerJobPublisher_NilChannelGraceful(t *testing.T) {
	pub := producer.NewAPIRunnerJobPublisher(nil)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	job := producer.APIRunnerJobPayload{
		RunID:         "run_test_123",
		APIRequestID:  "req_test_456",
		EnvironmentID: "env_test_789",
		ActorID:       "usr_test",
		Method:        "GET",
		RelativePath:  "/api/v1/health",
		Timestamp:     time.Now().UTC(),
	}

	err := pub.PublishAPIRunnerJob(ctx, job)
	if err != nil {
		t.Fatalf("expected nil error on nil channel graceful fallback, got: %v", err)
	}

	var nilPub *producer.APIRunnerJobPublisher
	if err := nilPub.PublishAPIRunnerJob(ctx, job); err != nil {
		t.Fatalf("expected nil error on nil receiver graceful fallback, got: %v", err)
	}
}
