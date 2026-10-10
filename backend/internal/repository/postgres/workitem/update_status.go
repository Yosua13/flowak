package workitem

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"backend/internal/domain/workitem"
)

const transitionWorkItemSQL = `UPDATE work_items SET status=$1::varchar,blocked_reason=CASE WHEN $1::varchar='Blocked' THEN $2 ELSE blocked_reason END,resolution=CASE WHEN $1::varchar='Done' THEN $3 ELSE resolution END,row_version=row_version+1,updated_at=CURRENT_TIMESTAMP WHERE id=$4 AND row_version=$5 AND deleted_at IS NULL`

// UpdateStatus executes a state transition update using optimistic locking on row_version.
func (r *PostgresWorkItemRepository) UpdateStatus(ctx context.Context, id string, targetStatus workitem.Status, rowVersion int, note, resolution, changedBy string) (*workitem.WorkItem, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Fetch current status and project_id for audit & notifications
	var currentStatus string
	var projectID string
	err = tx.QueryRowContext(ctx, `SELECT status, project_id FROM work_items WHERE id=$1 AND deleted_at IS NULL`, id).Scan(&currentStatus, &projectID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, workitem.ErrWorkItemNotFound
		}
		return nil, fmt.Errorf("failed to read work item before transition: %w", err)
	}

	// 2. Perform optimistic locking update
	var noteParam *string
	if note != "" {
		noteParam = &note
	}
	var resParam *string
	if resolution != "" {
		resParam = &resolution
	}

	result, err := tx.ExecContext(ctx, transitionWorkItemSQL, string(targetStatus), noteParam, resParam, id, rowVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to execute status transition: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to check affected rows: %w", err)
	}
	if rowsAffected == 0 {
		return nil, workitem.ErrOptimisticLockConflict
	}

	// 3. Record audit history
	historyID := "wih_" + generateUUID()
	_, err = tx.ExecContext(ctx, `INSERT INTO work_item_status_history(id,work_item_id,from_status,to_status,changed_by,note) VALUES($1,$2,$3,$4,$5,$6)`,
		historyID,
		id,
		currentStatus,
		string(targetStatus),
		changedBy,
		noteParam,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to record status history: %w", err)
	}

	// 4. Create watcher notifications
	watcherRows, err := tx.QueryContext(ctx, `SELECT user_id FROM work_item_watchers WHERE work_item_id=$1`, id)
	if err == nil {
		defer watcherRows.Close()
		transitionKey := fmt.Sprintf("transition:%s:%d", id, rowVersion)
		for watcherRows.Next() {
			var watcherID string
			if err := watcherRows.Scan(&watcherID); err == nil && watcherID != changedBy {
				dedup := "watcher:" + transitionKey + ":" + watcherID
				notifID := "ntf_" + generateUUID()
				payload := `{"work_item_id":"` + id + `"}`
				_ = tx.QueryRowContext(ctx, `INSERT INTO notifications(id,user_id,project_id,title,body,type,event_name,payload,dedup_key) VALUES($1,$2,$3,$4,$5,'info','work_item.watcher_updated',$6::jsonb,$7) ON CONFLICT (dedup_key) WHERE dedup_key IS NOT NULL DO NOTHING RETURNING id`,
					notifID, watcherID, projectID, "Work item yang diikuti diperbarui", "Buka work item untuk meninjau perubahan terbaru.", payload, dedup).Scan(&notifID)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transition: %w", err)
	}

	// 5. Query updated work item
	query := `SELECT ` + workItemFields + ` FROM work_items WHERE id=$1 AND deleted_at IS NULL`
	updated, err := scanWorkItemRow(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated work item: %w", err)
	}

	// Suppress unused warning for time package if needed
	_ = time.Now()

	return updated, nil
}
