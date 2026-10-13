package workitem

import "context"

// ListFilter specifies filtering and pagination criteria for work items.
type ListFilter struct {
	ProjectID  string
	Status     string
	AssigneeID string
	NodeID     string
	Limit      int
	Sort       string
	PageMode   bool
	Cursor     string
}

// ListResult represents the paginated result of a work items query.
type ListResult struct {
	Items      []WorkItem
	NextCursor string
	Sort       string
	HasMore    bool
}

// WorkItemRepository defines the persistence contract for work items.
type WorkItemRepository interface {
	FindByKey(ctx context.Context, key string) (*WorkItem, error)
	FindByID(ctx context.Context, id string) (*WorkItem, error)
	Insert(ctx context.Context, item *WorkItem) error
	UpdateStatus(ctx context.Context, id string, targetStatus Status, rowVersion int, note, resolution, changedBy string) (*WorkItem, error)
	Update(ctx context.Context, item *WorkItem) (*WorkItem, error)
	ListByProject(ctx context.Context, filter ListFilter) (*ListResult, error)
}

// EventPublisher defines the contract for broadcasting work item lifecycle events.
type EventPublisher interface {
	PublishWorkItemEvent(ctx context.Context, routingKey string, payload any) error
}
