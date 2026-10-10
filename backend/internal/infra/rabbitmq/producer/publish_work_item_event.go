package producer

import (
	"context"
	"encoding/json"
	"time"

	"backend/internal/infra/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
)

// WorkItemTransitionedPayload defines the message contract published when a work item transitions.
type WorkItemTransitionedPayload struct {
	EventID    string    `json:"id"`
	EventName  string    `json:"name"`
	ProjectID  string    `json:"project_id"`
	WorkItemID string    `json:"work_item_id"`
	WorkKey    string    `json:"work_key"`
	Status     string    `json:"status"`
	OldStatus  string    `json:"old_status,omitempty"`
	ActorID    string    `json:"actor_id,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
	Payload    any       `json:"payload,omitempty"`
}

// WorkItemEventPublisher manages publishing work item domain events to RabbitMQ.
type WorkItemEventPublisher struct {
	channel  *amqp.Channel
	exchange string
}

// NewWorkItemEventPublisher creates a publisher for work item events targeting the core exchange.
func NewWorkItemEventPublisher(ch *amqp.Channel) *WorkItemEventPublisher {
	return &WorkItemEventPublisher{
		channel:  ch,
		exchange: rabbitmq.ExchangeEvents,
	}
}

// PublishWorkItemEvent publishes an event message to the RabbitMQ exchange with the specified routing key.
func (p *WorkItemEventPublisher) PublishWorkItemEvent(ctx context.Context, routingKey string, payload any) error {
	if p == nil || p.channel == nil {
		// Graceful degradation when RabbitMQ broker is offline or nil in test environments
		return nil
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return p.channel.PublishWithContext(
		ctx,
		p.exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now().UTC(),
			Body:         body,
		},
	)
}
