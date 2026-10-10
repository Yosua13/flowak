package module

import (
	"context"
	"database/sql"
	"strings"

	"backend/internal/domain/module"
)

// DeleteNodes tombstones workflow nodes after ensuring connected edges have been removed.
func DeleteNodes(ctx context.Context, tx *sql.Tx, moduleID string, deletedNodes []module.GraphDelete) error {
	for _, item := range deletedNodes {
		if strings.TrimSpace(item.ID) == "" || item.RowVersion < 1 {
			return module.ErrInvalidGraphReference
		}

		var hasEdges bool
		err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM workflow_edges WHERE module_id = $1 AND deleted_at IS NULL AND (from_node_id = $2 OR to_node_id = $2))`, moduleID, item.ID).Scan(&hasEdges)
		if err != nil {
			return err
		}
		if hasEdges {
			return module.ErrInvalidGraphReference
		}

		result, err := tx.ExecContext(ctx, `UPDATE workflow_nodes SET deleted_at = CURRENT_TIMESTAMP, row_version = row_version + 1, updated_at = CURRENT_TIMESTAMP
			WHERE id = $1 AND module_id = $2 AND deleted_at IS NULL AND row_version = $3`, item.ID, moduleID, item.RowVersion)
		if err != nil {
			return err
		}
		n, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if n == 0 {
			return module.ErrGraphConflict
		}
	}
	return nil
}

// DeleteEdges tombstones workflow transitions.
func DeleteEdges(ctx context.Context, tx *sql.Tx, moduleID string, deletedEdges []module.GraphDelete) error {
	for _, item := range deletedEdges {
		if strings.TrimSpace(item.ID) == "" || item.RowVersion < 1 {
			return module.ErrInvalidGraphReference
		}

		result, err := tx.ExecContext(ctx, `UPDATE workflow_edges SET deleted_at = CURRENT_TIMESTAMP, row_version = row_version + 1, updated_at = CURRENT_TIMESTAMP
			WHERE id = $1 AND module_id = $2 AND deleted_at IS NULL AND row_version = $3`, item.ID, moduleID, item.RowVersion)
		if err != nil {
			return err
		}
		n, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if n == 0 {
			return module.ErrGraphConflict
		}
	}
	return nil
}
