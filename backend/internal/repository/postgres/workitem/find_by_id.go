package workitem

import (
	"context"
	"database/sql"
	"errors"

	"backend/internal/domain/workitem"
)

// FindByID queries a work item by its primary key ID.
func (r *PostgresWorkItemRepository) FindByID(ctx context.Context, id string) (*workitem.WorkItem, error) {
	query := `SELECT ` + workItemFields + ` FROM work_items WHERE id=$1 AND deleted_at IS NULL`
	item, err := scanWorkItemRow(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, workitem.ErrWorkItemNotFound
		}
		return nil, err
	}
	return item, nil
}
