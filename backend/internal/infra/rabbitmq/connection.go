package rabbitmq

import (
	"errors"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// ErrEmptyURL is returned when the provided RabbitMQ URL is empty.
var ErrEmptyURL = errors.New("rabbitmq url cannot be empty")

// NewRabbitMQConnection establishes an AMQP connection to the specified broker URL.
func NewRabbitMQConnection(url string) (*amqp.Connection, error) {
	if url == "" {
		return nil, ErrEmptyURL
	}

	conn, err := amqp.Dial(url)
	if err != nil {
		log.Printf("[RabbitMQ] Warning: connection to broker at %s failed: %v\n", url, err)
		return nil, fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	log.Printf("[RabbitMQ] Successfully connected to broker\n")
	return conn, nil
}
