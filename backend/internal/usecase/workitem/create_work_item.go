package workitem

import (
	"context"
	"strings"
	"time"

	"backend/internal/domain/workitem"
)

// CreateWorkItemInput holds parameters for creating a new work item.
type CreateWorkItemInput struct {
	ProjectID     string
	ReporterID    string
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
}

// CreateWorkItemUseCase handles work item creation workflow.
type CreateWorkItemUseCase struct {
	repo workitem.WorkItemRepository
}

// NewCreateWorkItemUseCase creates a new CreateWorkItemUseCase.
func NewCreateWorkItemUseCase(repo workitem.WorkItemRepository) *CreateWorkItemUseCase {
	return &CreateWorkItemUseCase{repo: repo}
}

// Execute validates and creates a new work item through the repository.
func (uc *CreateWorkItemUseCase) Execute(ctx context.Context, input CreateWorkItemInput) (*workitem.WorkItem, error) {
	if err := ValidateWorkItemFields(
		input.Type,
		input.Title,
		input.Priority,
		input.Status,
		input.Points,
		input.NodeID,
		input.Description,
		input.BlockedReason,
		input.Resolution,
		input.StartDate,
		input.DueDate,
		true,
	); err != nil {
		return nil, err
	}

	priority := input.Priority
	if priority == "" {
		priority = string(workitem.PriorityMedium)
	}

	status := input.Status
	if status == "" {
		status = string(workitem.StatusBacklog)
	}

	var startDate, dueDate *time.Time
	if input.StartDate != nil && *input.StartDate != "" {
		parsed, _ := time.Parse("2006-01-02", *input.StartDate)
		startDate = &parsed
	}
	if input.DueDate != nil && *input.DueDate != "" {
		parsed, _ := time.Parse("2006-01-02", *input.DueDate)
		dueDate = &parsed
	}

	item := &workitem.WorkItem{
		ProjectID:     input.ProjectID,
		ReporterID:    input.ReporterID,
		ModuleID:      input.ModuleID,
		NodeID:        input.NodeID,
		FacetKey:      input.FacetKey,
		ParentID:      input.ParentID,
		Type:          workitem.WorkItemType(input.Type),
		Title:         strings.TrimSpace(input.Title),
		Description:   input.Description,
		Priority:      workitem.Priority(priority),
		Points:        input.Points,
		Status:        workitem.Status(status),
		AssigneeID:    input.AssigneeID,
		StartDate:     startDate,
		DueDate:       dueDate,
		BlockedReason: input.BlockedReason,
		Resolution:    input.Resolution,
	}

	if err := uc.repo.Insert(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}
