package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"backend/internal/infra/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
)

// EventBroadcaster is the callback function signature for forwarding broadcast events to active SSE client streams.
type EventBroadcaster func(projectID string, event []byte)

// ConsumeSSEBroadcast starts a worker loop consuming messages from QueueSSEBroadcast and broadcasting them to SSE subscribers.
func ConsumeSSEBroadcast(ctx context.Context, ch *amqp.Channel, broadcaster EventBroadcaster) error {
	if ch == nil {
		return fmt.Errorf("amqp channel cannot be nil")
	}

	msgs, err := ch.ConsumeWithContext(
		ctx,
		rabbitmq.QueueSSEBroadcast,
		"flowak-sse-broadcast-worker",
		false, // manual acknowledgement
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer on %s: %w", rabbitmq.QueueSSEBroadcast, err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				return nil
			}

			var envelope struct {
				ProjectID string `json:"project_id"`
			}
			if err := json.Unmarshal(msg.Body, &envelope); err == nil && envelope.ProjectID != "" {
				if broadcaster != nil {
					broadcaster(envelope.ProjectID, msg.Body)
				}
			} else {
				log.Printf("SSE broadcast event missing project_id: %s", string(msg.Body))
			}

			_ = msg.Ack(false)
		}
	}
}
