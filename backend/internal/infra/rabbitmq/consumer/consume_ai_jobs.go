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

// AIJobHandler is the callback function signature for executing background AI tasks.
type AIJobHandler func(ctx context.Context, job producer.AIJobPayload) error

// ConsumeAIJobs starts a worker loop consuming AI task messages from QueueAIJobs.
func ConsumeAIJobs(ctx context.Context, ch *amqp.Channel, handler AIJobHandler) error {
	if ch == nil {
		return fmt.Errorf("amqp channel cannot be nil")
	}

	msgs, err := ch.ConsumeWithContext(
		ctx,
		rabbitmq.QueueAIJobs,
		"flowak-ai-job-worker",
		false, // manual acknowledgement
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer on %s: %w", rabbitmq.QueueAIJobs, err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				return nil
			}

			var payload producer.AIJobPayload
			if err := json.Unmarshal(msg.Body, &payload); err != nil {
				log.Printf("[Worker] Invalid AI job message format: %v", err)
				_ = msg.Nack(false, false) // discard malformed message
				continue
			}

			if handler != nil {
				if err := handler(ctx, payload); err != nil {
					log.Printf("[Worker] Error processing AI job %s: %v", payload.JobID, err)
					_ = msg.Nack(false, true) // requeue for retry
					continue
				}
			}

			_ = msg.Ack(false)
		}
	}
}
