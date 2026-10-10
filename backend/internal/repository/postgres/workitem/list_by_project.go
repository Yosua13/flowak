package workitem

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"backend/internal/domain/workitem"
)

type workItemCursor struct {
	Project  string `json:"project"`
	Sort     string `json:"sort"`
	Filter   string `json:"filter"`
	Sequence int64  `json:"sequence"`
	Ceiling  int64  `json:"ceiling"`
}

func workItemFilterKey(status, assignee, nodeID string) string {
	data, _ := json.Marshal([]string{status, assignee, nodeID})
	return string(data)
}

func encodeWorkItemCursor(cursor workItemCursor) string {
	data, _ := json.Marshal(cursor)
	return base64.RawURLEncoding.EncodeToString(data)
}

func decodeWorkItemCursor(raw string, cursor *workItemCursor) error {
	data, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil || len(data) > 1024 {
		return fmt.Errorf("invalid cursor")
	}
	if err := json.Unmarshal(data, cursor); err != nil {
		return err
	}
	if cursor.Sequence < 1 || cursor.Ceiling < cursor.Sequence {
		return fmt.Errorf("invalid cursor sequence")
	}
	return nil
}

// ListByProject queries work items for a project with optional filters and cursor pagination.
func (r *PostgresWorkItemRepository) ListByProject(ctx context.Context, filter workitem.ListFilter) (*workitem.ListResult, error) {
	limit := filter.Limit
	if limit < 1 {
		limit = 50
	} else if limit > 100 {
		limit = 100
	}

	sortOrder := filter.Sort
	if sortOrder != "oldest" {
		sortOrder = "newest"
	}

	args := []any{filter.ProjectID}
	where := "project_id=$1 AND deleted_at IS NULL"

	for _, item := range []struct{ value, column string }{
		{filter.Status, "status"},
		{filter.AssigneeID, "assignee_id"},
		{filter.NodeID, "node_id"},
	} {
		if item.value != "" {
			args = append(args, item.value)
			where += fmt.Sprintf(" AND %s=$%d", item.column, len(args))
		}
	}

	var cursor workItemCursor
	if filter.PageMode {
		if filter.Cursor != "" {
			if err := decodeWorkItemCursor(filter.Cursor, &cursor); err != nil || cursor.Project != filter.ProjectID || cursor.Sort != sortOrder || cursor.Filter != workItemFilterKey(filter.Status, filter.AssigneeID, filter.NodeID) {
				return nil, errors.New("invalid cursor for sort or filters")
			}
		} else {
			cursor.Project, cursor.Sort, cursor.Filter = filter.ProjectID, sortOrder, workItemFilterKey(filter.Status, filter.AssigneeID, filter.NodeID)
			if err := r.db.QueryRowContext(ctx, `SELECT COALESCE(MAX(sequence),0) FROM work_items WHERE project_id=$1`, filter.ProjectID).Scan(&cursor.Ceiling); err != nil {
				return nil, fmt.Errorf("failed to initialize pagination: %w", err)
			}
		}
		args = append(args, cursor.Ceiling)
		where += fmt.Sprintf(" AND sequence <= $%d", len(args))
		if cursor.Sequence > 0 {
			args = append(args, cursor.Sequence)
			operator := "<"
			if sortOrder == "oldest" {
				operator = ">"
			}
			where += fmt.Sprintf(" AND sequence %s $%d", operator, len(args))
		}
	} else if filter.Cursor != "" {
		n, err := strconv.ParseInt(filter.Cursor, 10, 64)
		if err != nil || n < 1 {
			return nil, errors.New("invalid cursor")
		}
		args = append(args, n)
		operator := "<"
		if sortOrder == "oldest" {
			operator = ">"
		}
		where += fmt.Sprintf(" AND sequence %s $%d", operator, len(args))
	}

	queryLimit := limit
	if filter.PageMode {
		queryLimit++
	}
	args = append(args, queryLimit)

	direction := "DESC"
	if sortOrder == "oldest" {
		direction = "ASC"
	}

	query := `SELECT sequence,` + workItemFields + ` FROM work_items WHERE ` + where + ` ORDER BY sequence ` + direction + ` LIMIT $` + strconv.Itoa(len(args))
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list work items: %w", err)
	}
	defer rows.Close()

	items := []workitem.WorkItem{}
	var lastSequence int64
	hasMore := false

	for rows.Next() {
		var i workitem.WorkItem
		var sequence int64
		var rawType, rawPriority, rawStatus string
		if err := rows.Scan(
			&sequence,
			&i.ID,
			&i.Key,
			&i.ProjectID,
			&i.ModuleID,
			&i.NodeID,
			&i.FacetKey,
			&i.ParentID,
			&rawType,
			&i.Title,
			&i.Description,
			&rawPriority,
			&i.Points,
			&rawStatus,
			&i.AssigneeID,
			&i.ReporterID,
			&i.StartDate,
			&i.DueDate,
			&i.BlockedReason,
			&i.Resolution,
			&i.RowVersion,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan work item row: %w", err)
		}

		i.Type = workitem.WorkItemType(rawType)
		i.Priority = workitem.Priority(rawPriority)
		i.Status = workitem.Status(rawStatus)

		if len(items) < limit {
			items = append(items, i)
			lastSequence = sequence
		} else {
			hasMore = true
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	nextCursor := ""
	if filter.PageMode && hasMore {
		cursor.Sequence = lastSequence
		nextCursor = encodeWorkItemCursor(cursor)
	}

	return &workitem.ListResult{
		Items:      items,
		NextCursor: nextCursor,
		Sort:       sortOrder,
		HasMore:    hasMore,
	}, nil
}
