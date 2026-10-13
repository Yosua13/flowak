package workitem

import (
	"context"
	"database/sql"
	"errors"

	"backend/internal/domain/workitem"
)

// FindByKey queries a work item by its work_key identifier.
func (r *PostgresWorkItemRepository) FindByKey(ctx context.Context, key string) (*workitem.WorkItem, error) {
	query := `SELECT ` + workItemFields + ` FROM work_items WHERE work_key=$1 AND deleted_at IS NULL`
	item, err := scanWorkItemRow(r.db.QueryRowContext(ctx, query, key))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, workitem.ErrWorkItemNotFound
		}
		return nil, err
	}
	return item, nil
}
