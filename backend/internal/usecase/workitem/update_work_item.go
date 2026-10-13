package workitem

import (
	"context"
	"fmt"
	"strings"
	"time"

	"backend/internal/domain/workitem"
)

// UpdateWorkItemInput contains payload fields for updating a work item.
type UpdateWorkItemInput struct {
	RowVersion    int
	Type          string
	Title         string
	Description   *string
	Priority      string
	Points        *int
	Status        string
	ModuleID      *string
	NodeID        *string
	FacetKey      *string
	ParentID      *string
	AssigneeID    *string
	StartDate     *string
	DueDate       *string
	BlockedReason *string
	Resolution    *string
	WasCleared    func(field string) bool
}

// UpdateWorkItemUseCase coordinates general work item updates.
type UpdateWorkItemUseCase struct {
	repo workitem.WorkItemRepository
}

// NewUpdateWorkItemUseCase creates a new UpdateWorkItemUseCase.
func NewUpdateWorkItemUseCase(repo workitem.WorkItemRepository) *UpdateWorkItemUseCase {
	return &UpdateWorkItemUseCase{repo: repo}
}

// Execute validates and applies a partial update to an existing work item.
func (uc *UpdateWorkItemUseCase) Execute(ctx context.Context, actorID, key string, input UpdateWorkItemInput) (*workitem.WorkItem, error) {
	item, err := uc.repo.FindByKey(ctx, key)
	if err != nil {
		return nil, err
	}

	if input.RowVersion < 1 {
		return nil, fmt.Errorf("row_version is required")
	}

	if input.Status != "" && input.Status != string(item.Status) {
		return nil, fmt.Errorf("use the transition endpoint to change status")
	}

	wasCleared := input.WasCleared
	if wasCleared == nil {
		wasCleared = func(string) bool { return false }
	}

	// Patch fields inherit existing values unless cleared or explicitly replaced
	itemType := input.Type
	if itemType == "" {
		itemType = string(item.Type)
	}

	title := input.Title
	if title == "" {
		title = item.Title
	}

	priority := input.Priority
	if priority == "" {
		priority = string(item.Priority)
	}

	status := input.Status
	if status == "" {
		status = string(item.Status)
	}

	moduleID := input.ModuleID
	if moduleID == nil {
		moduleID = item.ModuleID
	}

	nodeID := input.NodeID
	if nodeID == nil {
		nodeID = item.NodeID
	}

	facetKey := input.FacetKey
	if facetKey == nil {
		facetKey = item.FacetKey
	}

	parentID := input.ParentID
	if parentID == nil {
		parentID = item.ParentID
	}
	if wasCleared("parent_id") {
		parentID = nil
	}

	description := input.Description
	if description == nil {
		description = item.Description
	}

	points := input.Points
	if points == nil {
		points = item.Points
	}
	if wasCleared("points") {
		points = nil
	}

	assigneeID := input.AssigneeID
	if assigneeID == nil {
		assigneeID = item.AssigneeID
	}
	if wasCleared("assignee_id") {
		assigneeID = nil
	}

	blockedReason := input.BlockedReason
	if blockedReason == nil {
		blockedReason = item.BlockedReason
	}

	resolution := input.Resolution
	if resolution == nil {
		resolution = item.Resolution
	}

	startDateStr := input.StartDate
	if startDateStr == nil && item.StartDate != nil {
		s := item.StartDate.Format("2006-01-02")
		startDateStr = &s
	}

	dueDateStr := input.DueDate
	if dueDateStr == nil && item.DueDate != nil {
		s := item.DueDate.Format("2006-01-02")
		dueDateStr = &s
	}
	if wasCleared("due_date") {
		dueDateStr = nil
	}

	if err := ValidateWorkItemFields(
		itemType,
		title,
		priority,
		status,
		points,
		nodeID,
		description,
		blockedReason,
		resolution,
		startDateStr,
		dueDateStr,
		false,
	); err != nil {
		return nil, err
	}

	var startDate, dueDate *time.Time
	if startDateStr != nil && *startDateStr != "" {
		parsed, _ := time.Parse("2006-01-02", *startDateStr)
		startDate = &parsed
	}
	if dueDateStr != nil && *dueDateStr != "" {
		parsed, _ := time.Parse("2006-01-02", *dueDateStr)
		dueDate = &parsed
	}

	toUpdate := &workitem.WorkItem{
		ID:            item.ID,
		Key:           item.Key,
		ProjectID:     item.ProjectID,
		ModuleID:      moduleID,
		NodeID:        nodeID,
		FacetKey:      facetKey,
		ParentID:      parentID,
		Type:          workitem.WorkItemType(itemType),
		Title:         strings.TrimSpace(title),
		Description:   description,
		Priority:      workitem.Priority(priority),
		Points:        points,
		Status:        workitem.Status(status),
		AssigneeID:    assigneeID,
		StartDate:     startDate,
		DueDate:       dueDate,
		BlockedReason: blockedReason,
		Resolution:    resolution,
		RowVersion:    input.RowVersion,
	}

	return uc.repo.Update(ctx, toUpdate)
}
