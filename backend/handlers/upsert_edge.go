package handlers

import (
	"database/sql"
)

func upsertEdge(tx *sql.Tx, moduleID, edgeID string, idx int, edge map[string]any, expectedVersion int, hasVersion bool) error {
	metadata := jsonText(map[string]any{"source": "frontend_graph", "raw": graphPayload(edge)})
	values := []any{edgeID, moduleID, stringField(edge, "from"), stringField(edge, "to"), nullableString(stringField(edge, "label")), nullableString(firstString(edge, "condition", "conditionText", "condition_text")), metadata, idx}
	var currentVersion int
	err := tx.QueryRow("SELECT row_version FROM workflow_edges WHERE id = $1 AND module_id = $2 AND deleted_at IS NULL", edgeID, moduleID).Scan(&currentVersion)
	if err == sql.ErrNoRows {
		var foreignModule string
		err = tx.QueryRow("SELECT module_id FROM workflow_edges WHERE id = $1", edgeID).Scan(&foreignModule)
		if err == nil {
			return &graphSyncError{Code: graphInvalidCode, Message: "edge belongs to another module or is deleted"}
		}
		if err != sql.ErrNoRows {
			return err
		}
		var existingID string
		err = tx.QueryRow("SELECT id FROM workflow_edges WHERE module_id = $1 AND from_node_id = $2 AND to_node_id = $3 AND deleted_at IS NULL", moduleID, values[2], values[3]).Scan(&existingID)
		if err == nil {
			return &graphSyncError{Code: graphInvalidCode, Message: "duplicate edge must use its existing id"}
		}
		if err != sql.ErrNoRows {
			return err
		}
		_, err = tx.Exec(`INSERT INTO workflow_edges (
			id, module_id, from_node_id, to_node_id, label, condition_text, metadata, sort_order, row_version, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8, 1, CURRENT_TIMESTAMP)`, values...)
		return err
	}
	if !hasVersion || expectedVersion != currentVersion {
		return &graphSyncError{Code: graphConflictCode, Message: "edge version conflict"}
	}
	values = append(values, expectedVersion)
	result, err := tx.Exec(`UPDATE workflow_edges SET
		from_node_id = $3, to_node_id = $4, label = $5, condition_text = $6, metadata = $7::jsonb, sort_order = $8,
		row_version = row_version + 1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND module_id = $2 AND deleted_at IS NULL AND row_version = $9
		AND (from_node_id, to_node_id, label, condition_text, metadata, sort_order)
		IS DISTINCT FROM ($3, $4, $5, $6, $7::jsonb, $8)`, values...)
	if err != nil {
		return err
	}
	changed, err := result.RowsAffected()
	if err != nil || changed > 0 {
		return err
	}
	if err := tx.QueryRow("SELECT row_version FROM workflow_edges WHERE id = $1 AND module_id = $2 AND deleted_at IS NULL", edgeID, moduleID).Scan(&currentVersion); err != nil {
		return err
	}
	if currentVersion != expectedVersion {
		return &graphSyncError{Code: graphConflictCode, Message: "edge version conflict"}
	}
	return nil
}
