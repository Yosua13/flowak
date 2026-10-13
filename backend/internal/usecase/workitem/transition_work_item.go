package workitem

import (
	"context"
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	"backend/internal/domain/workitem"
)

// TransitionWorkItemInput represents input required to transition a work item status.
type TransitionWorkItemInput struct {
	Status     string
	RowVersion int
	Note       string
	Resolution string
}

// TransitionWorkItemUseCase orchestrates status transition validation, DB update, and event publishing.
type TransitionWorkItemUseCase struct {
	repo      workitem.WorkItemRepository
	publisher workitem.EventPublisher
}

// NewTransitionWorkItemUseCase creates a new TransitionWorkItemUseCase.
func NewTransitionWorkItemUseCase(repo workitem.WorkItemRepository, pub workitem.EventPublisher) *TransitionWorkItemUseCase {
	return &TransitionWorkItemUseCase{
		repo:      repo,
		publisher: pub,
	}
}

func generateTransitionUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// Execute performs status transition validation, database optimistic update, and RabbitMQ event publication.
func (uc *TransitionWorkItemUseCase) Execute(ctx context.Context, actorID, key string, input TransitionWorkItemInput) (*workitem.WorkItem, error) {
	// 1. Basic input validation
	if input.RowVersion < 1 || !allowedWorkItemStatuses[input.Status] {
		return nil, fmt.Errorf("valid status and row_version are required")
	}

	targetStatus := workitem.Status(input.Status)

	// 2. Query existing work item
	item, err := uc.repo.FindByKey(ctx, key)
	if err != nil {
		return nil, err
	}

	// 3. Validate state-machine transition rules
	if !workitem.CanTransitionStatus(item.Status, targetStatus) {
		return nil, workitem.ErrInvalidStatusTransition
	}

	// 4. Validate status prerequisites
	if targetStatus == workitem.StatusDone && strings.TrimSpace(input.Resolution) == "" {
		return nil, fmt.Errorf("done status requires resolution")
	}
	if targetStatus == workitem.StatusBlocked && strings.TrimSpace(input.Note) == "" {
		return nil, fmt.Errorf("blocked status requires a reason in note")
	}

	// 5. Execute optimistic locking update in repository
	updated, err := uc.repo.UpdateStatus(ctx, item.ID, targetStatus, input.RowVersion, input.Note, input.Resolution, actorID)
	if err != nil {
		return nil, err
	}

	// 6. Broadcast event to RabbitMQ for SSE broadcast & real-time updates
	if uc.publisher != nil {
		eventPayload := map[string]any{
			"id":           "evt_" + generateTransitionUUID(),
			"name":         "workitem.transitioned",
			"project_id":   updated.ProjectID,
			"work_item_id": updated.ID,
			"work_key":     updated.Key,
			"status":       string(updated.Status),
			"old_status":   string(item.Status),
			"actor_id":     actorID,
			"timestamp":    time.Now().UTC(),
		}
		_ = uc.publisher.PublishWorkItemEvent(ctx, "workitem.transitioned", eventPayload)
	}

	return updated, nil
}
