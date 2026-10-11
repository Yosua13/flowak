package producer_test

import (
	"context"
	"testing"
	"time"

	"backend/internal/infra/rabbitmq/producer"
)

func TestAIJobPublisher_NilChannelGraceful(t *testing.T) {
	pub := producer.NewAIJobPublisher(nil)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	job := producer.AIJobPayload{
		JobID:     "job_test_123",
		Type:      "generate_flow",
		Prompt:    "Buatkan alur onboarding KYC",
		UserID:    "usr_test",
		Timestamp: time.Now().UTC(),
	}

	err := pub.PublishAIJob(ctx, job)
	if err != nil {
		t.Fatalf("expected nil error on nil channel graceful fallback, got: %v", err)
	}

	var nilPub *producer.AIJobPublisher
	if err := nilPub.PublishAIJob(ctx, job); err != nil {
		t.Fatalf("expected nil error on nil receiver graceful fallback, got: %v", err)
	}
}
