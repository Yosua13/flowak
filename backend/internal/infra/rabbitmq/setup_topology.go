package rabbitmq

import (
	"errors"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	ExchangeEvents    = "flowak.events"
	QueueAIJobs       = "flowak.ai.jobs"
	QueueSSEBroadcast = "flowak.sse.broadcast"
	QueueAuditLogs    = "flowak.audit.logs"
)

// SetupTopology declares the core topic exchange, queues, and routing bindings in the RabbitMQ broker.
func SetupTopology(ch *amqp.Channel) error {
	if ch == nil {
		return errors.New("amqp channel cannot be nil")
	}

	// 1. Declare Topic Exchange: flowak.events
	if err := ch.ExchangeDeclare(
		ExchangeEvents, // name
		"topic",         // type
		true,            // durable
		false,           // auto-deleted
		false,           // internal
		false,           // no-wait
		nil,             // arguments
	); err != nil {
		return fmt.Errorf("failed to declare exchange %s: %w", ExchangeEvents, err)
	}

	// 2. Declare Queues and bind them to the topic exchange
	queueDeclarations := []struct {
		queueName   string
		routingKeys []string
	}{
		{
			queueName:   QueueAIJobs,
			routingKeys: []string{QueueAIJobs, "ai.jobs.#"},
		},
		{
			queueName:   QueueSSEBroadcast,
			routingKeys: []string{QueueSSEBroadcast, "events.broadcast.#"},
		},
		{
			queueName:   QueueAuditLogs,
			routingKeys: []string{QueueAuditLogs, "audit.logs.#"},
		},
	}

	for _, item := range queueDeclarations {
		_, err := ch.QueueDeclare(
			item.queueName, // name
			true,           // durable
			false,          // auto-delete
			false,          // exclusive
			false,          // no-wait
			nil,            // arguments
		)
		if err != nil {
			return fmt.Errorf("failed to declare queue %s: %w", item.queueName, err)
		}

		for _, key := range item.routingKeys {
			if err := ch.QueueBind(
				item.queueName, // queue name
				key,            // routing key
				ExchangeEvents, // exchange
				false,          // no-wait
				nil,            // arguments
			); err != nil {
				return fmt.Errorf("failed to bind queue %s with routing key %s to exchange %s: %w", item.queueName, key, ExchangeEvents, err)
			}
		}
	}

	return nil
}
