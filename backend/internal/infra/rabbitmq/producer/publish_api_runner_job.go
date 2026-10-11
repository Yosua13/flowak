package producer

import (
	"context"
	"encoding/json"
	"time"

	"backend/internal/infra/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
)

// APIRunnerJobPayload defines the message contract published to flowak.api_runner.jobs.
type APIRunnerJobPayload struct {
	RunID         string            `json:"run_id"`
	APIRequestID  string            `json:"api_request_id"`
	EnvironmentID string            `json:"environment_id"`
	ActorID       string            `json:"actor_id"`
	Method        string            `json:"method,omitempty"`
	RelativePath  string            `json:"relative_path,omitempty"`
	Headers       map[string]string `json:"headers,omitempty"`
	Body          string            `json:"body,omitempty"`
	Timestamp     time.Time         `json:"timestamp"`
}

// APIRunnerJobPublisher manages publishing server-side API contract run jobs to RabbitMQ.
type APIRunnerJobPublisher struct {
	channel  *amqp.Channel
	exchange string
}

// NewAPIRunnerJobPublisher creates a new publisher targeting the core event exchange.
func NewAPIRunnerJobPublisher(ch *amqp.Channel) *APIRunnerJobPublisher {
	return &APIRunnerJobPublisher{
		channel:  ch,
		exchange: rabbitmq.ExchangeEvents,
	}
}

// PublishAPIRunnerJob enqueues an API contract execution job to the flowak.api_runner.jobs queue.
func (p *APIRunnerJobPublisher) PublishAPIRunnerJob(ctx context.Context, job APIRunnerJobPayload) error {
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
		rabbitmq.QueueAPIRunnerJobs,
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
