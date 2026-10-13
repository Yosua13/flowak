package ai

import (
	"context"
	"fmt"
	"time"

	"backend/handlers"
	domainAI "backend/internal/domain/ai"
	"backend/internal/infra/rabbitmq/producer"
)

// ExecuteEnqueueAIFlow enqueues an AI flow generation task to RabbitMQ if async mode is requested
// and publisher is available; otherwise it falls back smoothly to synchronous processing.
func ExecuteEnqueueAIFlow(ctx context.Context, repo domainAI.AIJobRepository, publisher *producer.AIJobPublisher, prompt string, userID string, isAsync bool) (*domainAI.AIJob, error) {
	if prompt == "" {
		return nil, fmt.Errorf("prompt cannot be empty")
	}

	jobID := "job_" + handlers.GenerateUUID()
	now := time.Now().UTC()

	job := &domainAI.AIJob{
		ID:        jobID,
		Type:      "generate_flow",
		Status:    domainAI.StatusPending,
		Prompt:    prompt,
		Payload:   map[string]any{"user_id": userID},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if repo != nil {
		if err := repo.CreateJob(ctx, job); err != nil {
			return nil, fmt.Errorf("failed to initialize AI job record: %w", err)
		}
	}

	payload := producer.AIJobPayload{
		JobID:     jobID,
		Type:      "generate_flow",
		Prompt:    prompt,
		UserID:    userID,
		Timestamp: now,
	}

	// Asynchronous mode: Enqueue to RabbitMQ if publisher is active
	if isAsync && publisher != nil {
		if err := publisher.PublishAIJob(ctx, payload); err == nil {
			return job, nil
		}
		// If publishing fails unexpectedly, gracefully fall back to synchronous execution
	}

	// Synchronous fallback mode: Process immediately and return completed job
	return ExecuteProcessAIFlowJob(ctx, repo, payload)
}
