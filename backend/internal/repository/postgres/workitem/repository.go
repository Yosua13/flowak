package workitem

import (
	"database/sql"

	"backend/internal/domain/workitem"
)

const workItemFields = `id, work_key, project_id, module_id, node_id, facet_key, parent_id, type, title, description, priority, points, status, assignee_id, reporter_id, start_date, due_date, blocked_reason, resolution, row_version, created_at, updated_at`

// PostgresWorkItemRepository implements workitem.WorkItemRepository for PostgreSQL.
type PostgresWorkItemRepository struct {
	db *sql.DB
}

// NewPostgresWorkItemRepository creates a new PostgreSQL work item repository.
func NewPostgresWorkItemRepository(db *sql.DB) *PostgresWorkItemRepository {
	return &PostgresWorkItemRepository{db: db}
}

// scanWorkItemRow helper scans a single row into a domain WorkItem entity.
func scanWorkItemRow(scanner interface{ Scan(...any) error }) (*workitem.WorkItem, error) {
	var item workitem.WorkItem
	var rawType, rawPriority, rawStatus string
	err := scanner.Scan(
		&item.ID,
		&item.Key,
		&item.ProjectID,
		&item.ModuleID,
		&item.NodeID,
		&item.FacetKey,
		&item.ParentID,
		&rawType,
		&item.Title,
		&item.Description,
		&rawPriority,
		&item.Points,
		&rawStatus,
		&item.AssigneeID,
		&item.ReporterID,
		&item.StartDate,
		&item.DueDate,
		&item.BlockedReason,
		&item.Resolution,
		&item.RowVersion,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	item.Type = workitem.WorkItemType(rawType)
	item.Priority = workitem.Priority(rawPriority)
	item.Status = workitem.Status(rawStatus)
	return &item, nil
}
