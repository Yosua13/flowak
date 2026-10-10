package workitem_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"backend/internal/domain/workitem"
	usecase "backend/internal/usecase/workitem"
)

type mockWorkItemRepo struct {
	findByKeyFn    func(ctx context.Context, key string) (*workitem.WorkItem, error)
	updateStatusFn  func(ctx context.Context, id string, targetStatus workitem.Status, rowVersion int, note, resolution, changedBy string) (*workitem.WorkItem, error)
	insertFn        func(ctx context.Context, item *workitem.WorkItem) error
	updateFn        func(ctx context.Context, item *workitem.WorkItem) (*workitem.WorkItem, error)
	listByProjectFn func(ctx context.Context, filter workitem.ListFilter) (*workitem.ListResult, error)
	findByIDFn      func(ctx context.Context, id string) (*workitem.WorkItem, error)
}

func (m *mockWorkItemRepo) FindByKey(ctx context.Context, key string) (*workitem.WorkItem, error) {
	if m.findByKeyFn != nil {
		return m.findByKeyFn(ctx, key)
	}
	return nil, nil
}
func (m *mockWorkItemRepo) FindByID(ctx context.Context, id string) (*workitem.WorkItem, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, nil
}
func (m *mockWorkItemRepo) Insert(ctx context.Context, item *workitem.WorkItem) error {
	if m.insertFn != nil {
		return m.insertFn(ctx, item)
	}
	return nil
}
func (m *mockWorkItemRepo) UpdateStatus(ctx context.Context, id string, targetStatus workitem.Status, rowVersion int, note, resolution, changedBy string) (*workitem.WorkItem, error) {
	if m.updateStatusFn != nil {
		return m.updateStatusFn(ctx, id, targetStatus, rowVersion, note, resolution, changedBy)
	}
	return nil, nil
}
func (m *mockWorkItemRepo) Update(ctx context.Context, item *workitem.WorkItem) (*workitem.WorkItem, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, item)
	}
	return nil, nil
}
func (m *mockWorkItemRepo) ListByProject(ctx context.Context, filter workitem.ListFilter) (*workitem.ListResult, error) {
	if m.listByProjectFn != nil {
		return m.listByProjectFn(ctx, filter)
	}
	return nil, nil
}

type mockEventPublisher struct {
	published []struct {
		RoutingKey string
		Payload    any
	}
}

func (m *mockEventPublisher) PublishWorkItemEvent(ctx context.Context, routingKey string, payload any) error {
	m.published = append(m.published, struct {
		RoutingKey string
		Payload    any
	}{RoutingKey: routingKey, Payload: payload})
	return nil
}

func TestTransitionWorkItemUseCase(t *testing.T) {
	sampleItem := &workitem.WorkItem{
		ID:         "wi_1",
		Key:        "FLOW-1",
		ProjectID:  "proj_1",
		Status:     workitem.StatusReady,
		RowVersion: 1,
	}

	t.Run("Rejects invalid row version", func(t *testing.T) {
		repo := &mockWorkItemRepo{}
		pub := &mockEventPublisher{}
		uc := usecase.NewTransitionWorkItemUseCase(repo, pub)

		_, err := uc.Execute(context.Background(), "user-1", "FLOW-1", usecase.TransitionWorkItemInput{
			Status:     "In Progress",
			RowVersion: 0,
		})
		if err == nil {
			t.Fatal("expected error on row_version < 1, got nil")
		}
	})

	t.Run("Rejects invalid state transition", func(t *testing.T) {
		repo := &mockWorkItemRepo{
			findByKeyFn: func(ctx context.Context, key string) (*workitem.WorkItem, error) {
				return sampleItem, nil
			},
		}
		pub := &mockEventPublisher{}
		uc := usecase.NewTransitionWorkItemUseCase(repo, pub)

		// Ready to Done is invalid in state machine
		_, err := uc.Execute(context.Background(), "user-1", "FLOW-1", usecase.TransitionWorkItemInput{
			Status:     "Done",
			RowVersion: 1,
			Resolution: "Done directly",
		})
		if !errors.Is(err, workitem.ErrInvalidStatusTransition) {
			t.Fatalf("expected ErrInvalidStatusTransition, got: %v", err)
		}
	})

	t.Run("Rejects Blocked without note", func(t *testing.T) {
		repo := &mockWorkItemRepo{
			findByKeyFn: func(ctx context.Context, key string) (*workitem.WorkItem, error) {
				return sampleItem, nil
			},
		}
		pub := &mockEventPublisher{}
		uc := usecase.NewTransitionWorkItemUseCase(repo, pub)

		_, err := uc.Execute(context.Background(), "user-1", "FLOW-1", usecase.TransitionWorkItemInput{
			Status:     "Blocked",
			RowVersion: 1,
			Note:       "",
		})
		if err == nil || err.Error() != "blocked status requires a reason in note" {
			t.Fatalf("expected blocked reason error, got: %v", err)
		}
	})

	t.Run("Rejects Done without resolution", func(t *testing.T) {
		inReviewItem := &workitem.WorkItem{
			ID:         "wi_2",
			Key:        "FLOW-2",
			ProjectID:  "proj_1",
			Status:     workitem.StatusInReview,
			RowVersion: 1,
		}
		repo := &mockWorkItemRepo{
			findByKeyFn: func(ctx context.Context, key string) (*workitem.WorkItem, error) {
				return inReviewItem, nil
			},
		}
		pub := &mockEventPublisher{}
		uc := usecase.NewTransitionWorkItemUseCase(repo, pub)

		_, err := uc.Execute(context.Background(), "user-1", "FLOW-2", usecase.TransitionWorkItemInput{
			Status:     "Done",
			RowVersion: 1,
			Resolution: "",
		})
		if err == nil || err.Error() != "done status requires resolution" {
			t.Fatalf("expected resolution required error, got: %v", err)
		}
	})

	t.Run("Successful transition updates DB and publishes RabbitMQ event", func(t *testing.T) {
		repo := &mockWorkItemRepo{
			findByKeyFn: func(ctx context.Context, key string) (*workitem.WorkItem, error) {
				return sampleItem, nil
			},
			updateStatusFn: func(ctx context.Context, id string, targetStatus workitem.Status, rowVersion int, note, resolution, changedBy string) (*workitem.WorkItem, error) {
				return &workitem.WorkItem{
					ID:         id,
					Key:        sampleItem.Key,
					ProjectID:  sampleItem.ProjectID,
					Status:     targetStatus,
					RowVersion: rowVersion + 1,
					UpdatedAt:  time.Now(),
				}, nil
			},
		}
		pub := &mockEventPublisher{}
		uc := usecase.NewTransitionWorkItemUseCase(repo, pub)

		result, err := uc.Execute(context.Background(), "user-1", "FLOW-1", usecase.TransitionWorkItemInput{
			Status:     "In Progress",
			RowVersion: 1,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Status != workitem.StatusInProgress || result.RowVersion != 2 {
			t.Fatalf("unexpected result: %+v", result)
		}

		// Verify event publication
		if len(pub.published) != 1 {
			t.Fatalf("expected 1 published event, got %d", len(pub.published))
		}
		if pub.published[0].RoutingKey != "workitem.transitioned" {
			t.Fatalf("expected routing key workitem.transitioned, got %s", pub.published[0].RoutingKey)
		}
	})

	t.Run("Propagates optimistic lock conflict", func(t *testing.T) {
		repo := &mockWorkItemRepo{
			findByKeyFn: func(ctx context.Context, key string) (*workitem.WorkItem, error) {
				return sampleItem, nil
			},
			updateStatusFn: func(ctx context.Context, id string, targetStatus workitem.Status, rowVersion int, note, resolution, changedBy string) (*workitem.WorkItem, error) {
				return nil, workitem.ErrOptimisticLockConflict
			},
		}
		pub := &mockEventPublisher{}
		uc := usecase.NewTransitionWorkItemUseCase(repo, pub)

		_, err := uc.Execute(context.Background(), "user-1", "FLOW-1", usecase.TransitionWorkItemInput{
			Status:     "In Progress",
			RowVersion: 1,
		})
		if !errors.Is(err, workitem.ErrOptimisticLockConflict) {
			t.Fatalf("expected ErrOptimisticLockConflict, got: %v", err)
		}
	})
}
