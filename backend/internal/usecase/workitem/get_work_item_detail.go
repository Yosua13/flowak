package workitem

import (
	"context"
	"strings"

	"backend/internal/domain/workitem"
)

// GetWorkItemDetailUseCase handles fetching work item details.
type GetWorkItemDetailUseCase struct {
	repo workitem.WorkItemRepository
}

// NewGetWorkItemDetailUseCase creates a new GetWorkItemDetailUseCase.
func NewGetWorkItemDetailUseCase(repo workitem.WorkItemRepository) *GetWorkItemDetailUseCase {
	return &GetWorkItemDetailUseCase{repo: repo}
}

// Execute queries the work item entity by its unique key.
func (uc *GetWorkItemDetailUseCase) Execute(ctx context.Context, key string) (*workitem.WorkItem, error) {
	cleanKey := strings.TrimSpace(key)
	if cleanKey == "" {
		return nil, workitem.ErrWorkItemNotFound
	}
	return uc.repo.FindByKey(ctx, cleanKey)
}
