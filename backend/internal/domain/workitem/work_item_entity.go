package workitem

import "time"

type WorkItemType string
type Priority string
type Status string

const (
	TypeStory    WorkItemType = "Story"
	TypeTask     WorkItemType = "Task"
	TypeBug      WorkItemType = "Bug"
	TypeReview   WorkItemType = "Review"
	TypeResearch WorkItemType = "Research"
	TypeSubtask  WorkItemType = "Subtask"
)

const (
	StatusBacklog    Status = "Backlog"
	StatusReady      Status = "Ready"
	StatusInProgress Status = "In Progress"
	StatusInReview   Status = "In Review"
	StatusBlocked    Status = "Blocked"
	StatusDone       Status = "Done"
	StatusCanceled   Status = "Canceled"
)

const (
	PriorityLow      Priority = "low"
	PriorityMedium   Priority = "medium"
	PriorityHigh     Priority = "high"
	PriorityCritical Priority = "critical"
)

// WorkItem represents the core work item domain entity.
type WorkItem struct {
	ID            string       `json:"id"`
	Key           string       `json:"key"`
	ProjectID     string       `json:"project_id"`
	ModuleID      *string      `json:"module_id,omitempty"`
	NodeID        *string      `json:"node_id,omitempty"`
	FacetKey      *string      `json:"facet_key,omitempty"`
	ParentID      *string      `json:"parent_id,omitempty"`
	Type          WorkItemType `json:"type"`
	Title         string       `json:"title"`
	Description   *string      `json:"description,omitempty"`
	Priority      Priority     `json:"priority"`
	Points        *int         `json:"points,omitempty"`
	Status        Status       `json:"status"`
	AssigneeID    *string      `json:"assignee_id,omitempty"`
	ReporterID    string       `json:"reporter_id"`
	StartDate     *time.Time   `json:"start_date,omitempty"`
	DueDate       *time.Time   `json:"due_date,omitempty"`
	BlockedReason *string      `json:"blocked_reason,omitempty"`
	Resolution    *string      `json:"resolution,omitempty"`
	RowVersion    int          `json:"row_version"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
}
