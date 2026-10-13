package workitem

import (
	"context"
	"fmt"

	"backend/internal/domain/workitem"
)

// ListWorkItemsUseCase handles work items querying and pagination.
type ListWorkItemsUseCase struct {
	repo workitem.WorkItemRepository
}

// NewListWorkItemsUseCase creates a new ListWorkItemsUseCase.
func NewListWorkItemsUseCase(repo workitem.WorkItemRepository) *ListWorkItemsUseCase {
	return &ListWorkItemsUseCase{repo: repo}
}

// Execute validates query criteria and retrieves a paginated list of work items.
func (uc *ListWorkItemsUseCase) Execute(ctx context.Context, filter workitem.ListFilter) (*workitem.ListResult, error) {
	if filter.Status != "" && !allowedWorkItemStatuses[filter.Status] {
		return nil, fmt.Errorf("invalid status filter")
	}

	if filter.Limit < 0 || filter.Limit > 100 {
		return nil, fmt.Errorf("limit must be 1-100")
	}

	if filter.Sort != "" && filter.Sort != "newest" && filter.Sort != "oldest" {
		return nil, fmt.Errorf("sort must be newest or oldest")
	}

	return uc.repo.ListByProject(ctx, filter)
}
