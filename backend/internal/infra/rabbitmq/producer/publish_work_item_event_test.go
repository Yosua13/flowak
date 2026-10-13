package producer_test

import (
	"context"
	"testing"
	"time"

	"backend/internal/infra/rabbitmq/producer"
)

func TestWorkItemEventPublisher_NilChannelGraceful(t *testing.T) {
	pub := producer.NewWorkItemEventPublisher(nil)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	payload := producer.WorkItemTransitionedPayload{
		EventID:    "evt_test",
		EventName:  "workitem.transitioned",
		ProjectID:  "proj_test",
		WorkItemID: "wi_test",
		WorkKey:    "PROJ-1",
		Status:     "In Progress",
		OldStatus:  "Ready",
		ActorID:    "usr_test",
		Timestamp:  time.Now().UTC(),
	}

	err := pub.PublishWorkItemEvent(ctx, "workitem.transitioned", payload)
	if err != nil {
		t.Fatalf("expected nil error on nil channel, got: %v", err)
	}

	// Also test nil publisher receiver
	var nilPub *producer.WorkItemEventPublisher
	if err := nilPub.PublishWorkItemEvent(ctx, "workitem.transitioned", payload); err != nil {
		t.Fatalf("expected nil error on nil receiver, got: %v", err)
	}
}
