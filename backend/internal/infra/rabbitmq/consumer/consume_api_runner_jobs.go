package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"backend/internal/infra/rabbitmq"
	"backend/internal/infra/rabbitmq/producer"
	amqp "github.com/rabbitmq/amqp091-go"
)

// APIRunnerJobHandler is the callback function signature for executing server-side API contract runs.
type APIRunnerJobHandler func(ctx context.Context, job producer.APIRunnerJobPayload) error

// ConsumeAPIRunnerJobs starts a worker loop consuming API contract run messages from QueueAPIRunnerJobs.
func ConsumeAPIRunnerJobs(ctx context.Context, ch *amqp.Channel, handler APIRunnerJobHandler) error {
	if ch == nil {
		return fmt.Errorf("amqp channel cannot be nil")
	}

	msgs, err := ch.ConsumeWithContext(
		ctx,
		rabbitmq.QueueAPIRunnerJobs,
		"flowak-api-runner-job-worker",
		false, // manual acknowledgement
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer on %s: %w", rabbitmq.QueueAPIRunnerJobs, err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				return nil
			}

			var payload producer.APIRunnerJobPayload
			if err := json.Unmarshal(msg.Body, &payload); err != nil {
				log.Printf("[Worker] Invalid API runner job message format: %v", err)
				_ = msg.Nack(false, false) // discard malformed message
				continue
			}

			if handler != nil {
				if err := handler(ctx, payload); err != nil {
					log.Printf("[Worker] Error processing API runner job %s: %v", payload.RunID, err)
					_ = msg.Nack(false, true) // requeue for retry
					continue
				}
			}

			_ = msg.Ack(false)
		}
	}
}
