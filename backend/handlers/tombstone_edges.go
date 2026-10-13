package handlers

import (
	"database/sql"
	"strings"

	"backend/models"
)

func tombstoneEdges(tx *sql.Tx, moduleID string, deleted []models.GraphDelete) error {
	for _, item := range deleted {
		if strings.TrimSpace(item.ID) == "" || item.RowVersion < 1 {
			return &graphSyncError{Code: graphInvalidCode, Message: "deleted edge requires id and rowVersion"}
		}
		result, err := tx.Exec(`UPDATE workflow_edges SET deleted_at = CURRENT_TIMESTAMP, row_version = row_version + 1, updated_at = CURRENT_TIMESTAMP
			WHERE id = $1 AND module_id = $2 AND deleted_at IS NULL AND row_version = $3`, item.ID, moduleID, item.RowVersion)
		if err != nil {
			return err
		}
		n, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if n == 0 {
			return &graphSyncError{Code: graphConflictCode, Message: "edge version conflict"}
		}
	}
	return nil
}
