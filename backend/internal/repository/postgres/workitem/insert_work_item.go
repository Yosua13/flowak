package workitem

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"strings"

	"backend/internal/domain/workitem"
)

func generateUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// Insert creates a new work item with transactional sequence allocation and audit recording.
func (r *PostgresWorkItemRepository) Insert(ctx context.Context, item *workitem.WorkItem) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Reference validations
	if item.ModuleID != nil {
		var exists bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM modules WHERE id=$1 AND project_id=$2 AND status='active')`, *item.ModuleID, item.ProjectID).Scan(&exists); err != nil || !exists {
			return workitem.ErrModuleNotBelongToProject
		}
	}
	if item.NodeID != nil {
		var moduleID string
		if err := tx.QueryRowContext(ctx, `SELECT n.module_id FROM workflow_nodes n JOIN modules m ON m.id=n.module_id WHERE n.id=$1 AND n.deleted_at IS NULL AND m.project_id=$2`, *item.NodeID, item.ProjectID).Scan(&moduleID); err != nil {
			return workitem.ErrNodeNotBelongToProject
		}
		if item.ModuleID != nil && *item.ModuleID != moduleID {
			return workitem.ErrNodeNotBelongToModule
		}
		item.ModuleID = &moduleID
	}
	if item.ParentID != nil {
		var parentProject string
		if err := tx.QueryRowContext(ctx, `SELECT project_id FROM work_items WHERE id=$1 AND deleted_at IS NULL`, *item.ParentID).Scan(&parentProject); err != nil || parentProject != item.ProjectID {
			return workitem.ErrParentNotSameProject
		}
	}
	if item.AssigneeID != nil {
		var member bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM project_members WHERE project_id=$1 AND user_id=$2)`, item.ProjectID, *item.AssigneeID).Scan(&member); err != nil || !member {
			return workitem.ErrAssigneeNotMember
		}
	}

	// 2. Sequence and key allocation
	var sequence int64
	if err := tx.QueryRowContext(ctx, `INSERT INTO project_work_item_sequences(project_id,next_value) VALUES($1,2) ON CONFLICT(project_id) DO UPDATE SET next_value=project_work_item_sequences.next_value+1 RETURNING next_value-1`, item.ProjectID).Scan(&sequence); err != nil {
		return fmt.Errorf("failed to allocate work item sequence: %w", err)
	}

	var prefix string
	if err := tx.QueryRowContext(ctx, `SELECT work_item_prefix FROM projects WHERE id=$1`, item.ProjectID).Scan(&prefix); err != nil {
		if err == sql.ErrNoRows {
			return workitem.ErrProjectNotFound
		}
		return fmt.Errorf("failed to get project prefix: %w", err)
	}

	if item.ID == "" {
		item.ID = "wi_" + generateUUID()
	}
	item.Key = fmt.Sprintf("%s-%d", prefix, sequence)

	// 3. Insert record
	var startDateStr, dueDateStr *string
	if item.StartDate != nil {
		s := item.StartDate.Format("2006-01-02")
		startDateStr = &s
	}
	if item.DueDate != nil {
		s := item.DueDate.Format("2006-01-02")
		dueDateStr = &s
	}

	insertSQL := `INSERT INTO work_items(id,work_key,project_id,module_id,node_id,facet_key,parent_id,sequence,type,title,description,priority,points,status,assignee_id,reporter_id,start_date,due_date,blocked_reason,resolution) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,NULLIF($17,'')::date,NULLIF($18,'')::date,$19,$20)`
	_, err = tx.ExecContext(ctx, insertSQL,
		item.ID,
		item.Key,
		item.ProjectID,
		item.ModuleID,
		item.NodeID,
		item.FacetKey,
		item.ParentID,
		sequence,
		string(item.Type),
		strings.TrimSpace(item.Title),
		item.Description,
		string(item.Priority),
		item.Points,
		string(item.Status),
		item.AssigneeID,
		item.ReporterID,
		startDateStr,
		dueDateStr,
		item.BlockedReason,
		item.Resolution,
	)
	if err != nil {
		return fmt.Errorf("failed to insert work item: %w", err)
	}

	// 4. Record status history
	_, err = tx.ExecContext(ctx, `INSERT INTO work_item_status_history(id,work_item_id,to_status,changed_by,note) VALUES($1,$2,$3,$4,'created')`,
		"wih_"+generateUUID(),
		item.ID,
		string(item.Status),
		item.ReporterID,
	)
	if err != nil {
		return fmt.Errorf("failed to record initial status history: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit work item insert: %w", err)
	}

	// 5. Query back populated entity
	query := `SELECT ` + workItemFields + ` FROM work_items WHERE id=$1 AND deleted_at IS NULL`
	fresh, err := scanWorkItemRow(r.db.QueryRowContext(ctx, query, item.ID))
	if err != nil {
		return fmt.Errorf("failed to retrieve inserted work item: %w", err)
	}
	*item = *fresh
	return nil
}
