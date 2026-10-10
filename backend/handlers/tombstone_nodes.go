package handlers

import (
	"database/sql"
	"strings"

	"backend/models"
)

func tombstoneNodes(tx *sql.Tx, moduleID string, deleted []models.GraphDelete) error {
	for _, item := range deleted {
		if strings.TrimSpace(item.ID) == "" || item.RowVersion < 1 {
			return &graphSyncError{Code: graphInvalidCode, Message: "deleted node requires id and rowVersion"}
		}
		var hasEdges bool
		if err := tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM workflow_edges WHERE module_id = $1 AND deleted_at IS NULL AND (from_node_id = $2 OR to_node_id = $2))`, moduleID, item.ID).Scan(&hasEdges); err != nil {
			return err
		}
		if hasEdges {
			return &graphSyncError{Code: graphInvalidCode, Message: "connected edges must be explicitly deleted before deleting a node"}
		}
		result, err := tx.Exec(`UPDATE workflow_nodes SET deleted_at = CURRENT_TIMESTAMP, row_version = row_version + 1, updated_at = CURRENT_TIMESTAMP
			WHERE id = $1 AND module_id = $2 AND deleted_at IS NULL AND row_version = $3`, item.ID, moduleID, item.RowVersion)
		if err != nil {
			return err
		}
		n, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if n == 0 {
			return &graphSyncError{Code: graphConflictCode, Message: "node version conflict"}
		}
	}
	return nil
}
