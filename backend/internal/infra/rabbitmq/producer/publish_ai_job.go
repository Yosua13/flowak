package producer

import (
	"context"
	"encoding/json"
	"time"

	"backend/internal/infra/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
)

// AIJobPayload defines the message contract published to flowak.ai.jobs.
type AIJobPayload struct {
	JobID     string          `json:"job_id"`
	Type      string          `json:"type"` // e.g., "generate_flow", "audit_flow"
	Prompt    string          `json:"prompt,omitempty"`
	ModuleID  string          `json:"module_id,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"`
	UserID    string          `json:"user_id,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
}

// AIJobPublisher manages publishing AI generation/audit jobs to RabbitMQ.
type AIJobPublisher struct {
	channel  *amqp.Channel
	exchange string
}

// NewAIJobPublisher creates a new publisher targeting the core event exchange.
func NewAIJobPublisher(ch *amqp.Channel) *AIJobPublisher {
	return &AIJobPublisher{
		channel:  ch,
		exchange: rabbitmq.ExchangeEvents,
	}
}

// PublishAIJob enqueues an AI task to the flowak.ai.jobs queue via the event exchange.
func (p *AIJobPublisher) PublishAIJob(ctx context.Context, job AIJobPayload) error {
	if p == nil || p.channel == nil {
		// Graceful degradation when broker is offline or nil channel in test environments
		return nil
	}

	body, err := json.Marshal(job)
	if err != nil {
		return err
	}

	return p.channel.PublishWithContext(
		ctx,
		p.exchange,
		rabbitmq.QueueAIJobs,
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
