package workitem

import (
	"fmt"
	"strings"
	"time"

	"backend/internal/domain/workitem"
)

var allowedWorkItemTypes = map[string]bool{
	string(workitem.TypeStory):    true,
	string(workitem.TypeTask):     true,
	string(workitem.TypeBug):      true,
	string(workitem.TypeReview):   true,
	string(workitem.TypeResearch): true,
	string(workitem.TypeSubtask):  true,
}

var allowedWorkItemStatuses = map[string]bool{
	string(workitem.StatusBacklog):    true,
	string(workitem.StatusReady):      true,
	string(workitem.StatusInProgress): true,
	string(workitem.StatusInReview):   true,
	string(workitem.StatusBlocked):    true,
	string(workitem.StatusDone):       true,
	string(workitem.StatusCanceled):   true,
}

var allowedWorkItemPriorities = map[string]bool{
	string(workitem.PriorityLow):      true,
	string(workitem.PriorityMedium):   true,
	string(workitem.PriorityHigh):     true,
	string(workitem.PriorityCritical): true,
}

var allowedPoints = map[int]bool{1: true, 2: true, 3: true, 5: true, 8: true, 13: true}

// ValidateWorkItemFields validates inputs for creation or update.
func ValidateWorkItemFields(itemType, title, priority, status string, points *int, nodeID, description, blockedReason, resolution, startDate, dueDate *string, creating bool) error {
	if !allowedWorkItemTypes[itemType] {
		return fmt.Errorf("invalid work item type")
	}
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("title is required")
	}
	if priority != "" && !allowedWorkItemPriorities[priority] {
		return fmt.Errorf("invalid priority")
	}
	if points != nil && !allowedPoints[*points] {
		return fmt.Errorf("points must be one of 1, 2, 3, 5, 8, or 13")
	}
	if status != "" && !allowedWorkItemStatuses[status] {
		return fmt.Errorf("invalid status")
	}
	st := status
	if st == "" && creating {
		st = string(workitem.StatusBacklog)
	}
	if st == string(workitem.StatusBlocked) && (blockedReason == nil || strings.TrimSpace(*blockedReason) == "") {
		return fmt.Errorf("blocked status requires blocked_reason")
	}
	if st == string(workitem.StatusDone) && (resolution == nil || strings.TrimSpace(*resolution) == "") {
		return fmt.Errorf("done status requires resolution")
	}
	if nodeID == nil && (description == nil || strings.TrimSpace(*description) == "") {
		return fmt.Errorf("project-level work item requires description as project scope")
	}
	for _, date := range []*string{startDate, dueDate} {
		if date != nil && *date != "" {
			if _, err := time.Parse("2006-01-02", *date); err != nil {
				return fmt.Errorf("dates must use YYYY-MM-DD")
			}
		}
	}
	return nil
}
