package workitem

import (
	"context"
	"fmt"
	"strings"

	"backend/internal/domain/workitem"
)

// Update updates general fields of a work item with reference integrity and optimistic locking.
func (r *PostgresWorkItemRepository) Update(ctx context.Context, item *workitem.WorkItem) (*workitem.WorkItem, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Reference validations
	if item.ModuleID != nil {
		var exists bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM modules WHERE id=$1 AND project_id=$2 AND status='active')`, *item.ModuleID, item.ProjectID).Scan(&exists); err != nil || !exists {
			return nil, workitem.ErrModuleNotBelongToProject
		}
	}
	if item.NodeID != nil {
		var moduleID string
		if err := tx.QueryRowContext(ctx, `SELECT n.module_id FROM workflow_nodes n JOIN modules m ON m.id=n.module_id WHERE n.id=$1 AND n.deleted_at IS NULL AND m.project_id=$2`, *item.NodeID, item.ProjectID).Scan(&moduleID); err != nil {
			return nil, workitem.ErrNodeNotBelongToProject
		}
		if item.ModuleID != nil && *item.ModuleID != moduleID {
			return nil, workitem.ErrNodeNotBelongToModule
		}
		item.ModuleID = &moduleID
	}
	if item.ParentID != nil {
		if *item.ParentID == item.ID {
			return nil, workitem.ErrSelfParent
		}
		var parentProject string
		if err := tx.QueryRowContext(ctx, `SELECT project_id FROM work_items WHERE id=$1 AND deleted_at IS NULL`, *item.ParentID).Scan(&parentProject); err != nil || parentProject != item.ProjectID {
			return nil, workitem.ErrParentNotSameProject
		}
		var cycle bool
		err = tx.QueryRowContext(ctx, `WITH RECURSIVE ancestors AS (SELECT parent_id FROM work_items WHERE id=$1 UNION ALL SELECT w.parent_id FROM work_items w JOIN ancestors a ON w.id=a.parent_id WHERE a.parent_id IS NOT NULL) SELECT EXISTS(SELECT 1 FROM ancestors WHERE parent_id=$2)`, *item.ParentID, item.ID).Scan(&cycle)
		if err != nil || cycle {
			return nil, workitem.ErrHierarchyCycle
		}
	}
	if item.AssigneeID != nil {
		var member bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM project_members WHERE project_id=$1 AND user_id=$2)`, item.ProjectID, *item.AssigneeID).Scan(&member); err != nil || !member {
			return nil, workitem.ErrAssigneeNotMember
		}
	}

	// 2. Perform optimistic locking update
	var startDateStr, dueDateStr *string
	if item.StartDate != nil {
		s := item.StartDate.Format("2006-01-02")
		startDateStr = &s
	}
	if item.DueDate != nil {
		s := item.DueDate.Format("2006-01-02")
		dueDateStr = &s
	}

	updateSQL := `UPDATE work_items SET module_id=$1,node_id=$2,facet_key=$3,parent_id=$4,type=$5,title=$6,description=$7,priority=$8,points=$9,status=$10,assignee_id=$11,start_date=NULLIF($12,'')::date,due_date=NULLIF($13,'')::date,blocked_reason=$14,resolution=$15,row_version=row_version+1,updated_at=CURRENT_TIMESTAMP WHERE id=$16 AND row_version=$17 AND deleted_at IS NULL`
	result, err := tx.ExecContext(ctx, updateSQL,
		item.ModuleID,
		item.NodeID,
		item.FacetKey,
		item.ParentID,
		string(item.Type),
		strings.TrimSpace(item.Title),
		item.Description,
		string(item.Priority),
		item.Points,
		string(item.Status),
		item.AssigneeID,
		startDateStr,
		dueDateStr,
		item.BlockedReason,
		item.Resolution,
		item.ID,
		item.RowVersion,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update work item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to check affected rows: %w", err)
	}
	if rowsAffected == 0 {
		return nil, workitem.ErrOptimisticLockConflict
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit work item update: %w", err)
	}

	// 3. Retrieve fresh entity
	query := `SELECT ` + workItemFields + ` FROM work_items WHERE id=$1 AND deleted_at IS NULL`
	fresh, err := scanWorkItemRow(r.db.QueryRowContext(ctx, query, item.ID))
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated entity: %w", err)
	}
	return fresh, nil
}
