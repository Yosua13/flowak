package rabbitmq_test

import (
	"errors"
	"os"
	"testing"

	"backend/internal/infra/rabbitmq"
)

func TestNewRabbitMQConnection_EmptyURL(t *testing.T) {
	conn, err := rabbitmq.NewRabbitMQConnection("")
	if err == nil {
		t.Fatal("expected error for empty URL, got nil")
	}
	if conn != nil {
		t.Fatal("expected nil connection for empty URL")
	}
	if !errors.Is(err, rabbitmq.ErrEmptyURL) {
		t.Fatalf("expected ErrEmptyURL, got: %v", err)
	}
}

func TestNewRabbitMQConnection_OfflineBroker(t *testing.T) {
	// Point to an unused local port to test offline error handling
	conn, err := rabbitmq.NewRabbitMQConnection("amqp://guest:guest@127.0.0.1:56729/")
	if err == nil {
		t.Fatal("expected connection error for offline broker, got nil")
	}
	if conn != nil {
		t.Fatal("expected nil connection for offline broker")
	}
}

func TestSetupTopology_NilChannel(t *testing.T) {
	err := rabbitmq.SetupTopology(nil)
	if err == nil {
		t.Fatal("expected error when channel is nil, got nil")
	}
}

func TestSetupTopology_Constants(t *testing.T) {
	if rabbitmq.ExchangeEvents != "flowak.events" {
		t.Errorf("expected ExchangeEvents to be flowak.events, got: %s", rabbitmq.ExchangeEvents)
	}
	if rabbitmq.QueueAIJobs != "flowak.ai.jobs" {
		t.Errorf("expected QueueAIJobs to be flowak.ai.jobs, got: %s", rabbitmq.QueueAIJobs)
	}
	if rabbitmq.QueueSSEBroadcast != "flowak.sse.broadcast" {
		t.Errorf("expected QueueSSEBroadcast to be flowak.sse.broadcast, got: %s", rabbitmq.QueueSSEBroadcast)
	}
	if rabbitmq.QueueAuditLogs != "flowak.audit.logs" {
		t.Errorf("expected QueueAuditLogs to be flowak.audit.logs, got: %s", rabbitmq.QueueAuditLogs)
	}
}

func TestRabbitMQ_LiveIntegrationIfAvailable(t *testing.T) {
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		url = "amqp://guest:guest@localhost:5672/"
	}

	conn, err := rabbitmq.NewRabbitMQConnection(url)
	if err != nil {
		t.Logf("RabbitMQ broker is offline or unreachable (%v); skipping live topology integration test", err)
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		t.Fatalf("failed to open channel: %v", err)
	}
	defer ch.Close()

	if err := rabbitmq.SetupTopology(ch); err != nil {
		t.Fatalf("SetupTopology failed: %v", err)
	}
}
