package module

import (
	"context"
	"database/sql"
	"fmt"

	"backend/internal/domain/module"
)

// UpsertEdges synchronizes workflow edges between active module nodes.
func UpsertEdges(ctx context.Context, tx *sql.Tx, moduleID string, edges []module.Edge) error {
	// 1. Fetch active node IDs for graph reference validation
	rows, err := tx.QueryContext(ctx, "SELECT id FROM workflow_nodes WHERE module_id = $1 AND deleted_at IS NULL", moduleID)
	if err != nil {
		return fmt.Errorf("failed to fetch active nodes: %w", err)
	}
	defer rows.Close()

	activeNodes := make(map[string]bool)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return err
		}
		activeNodes[id] = true
	}
	if err := rows.Err(); err != nil {
		return err
	}

	// 2. Process each edge
	for idx, e := range edges {
		fromID := e.From
		toID := e.To
		if fromID == "" || toID == "" || fromID == toID || !activeNodes[fromID] || !activeNodes[toID] {
			return module.ErrInvalidGraphReference
		}

		edgeID := e.ID
		if edgeID == "" {
			edgeID = "edge_" + generateUUID()
			edges[idx].ID = edgeID
		}

		expectedVersion := e.RowVersion
		hasVersion := expectedVersion > 0

		metadata := jsonText(map[string]any{"source": "frontend_graph", "raw": map[string]any{"id": edgeID, "from": fromID, "to": toID, "label": e.Label, "condition": e.Condition}})
		values := []any{edgeID, moduleID, fromID, toID, nullableString(e.Label), nullableString(e.Condition), metadata, idx}

		var currentVersion int
		err = tx.QueryRowContext(ctx, "SELECT row_version FROM workflow_edges WHERE id = $1 AND module_id = $2 AND deleted_at IS NULL", edgeID, moduleID).Scan(&currentVersion)
		if err == sql.ErrNoRows {
			var foreignModule string
			if err = tx.QueryRowContext(ctx, "SELECT module_id FROM workflow_edges WHERE id = $1", edgeID).Scan(&foreignModule); err == nil {
				return module.ErrInvalidGraphReference
			}

			var existingID string
			if err = tx.QueryRowContext(ctx, "SELECT id FROM workflow_edges WHERE module_id = $1 AND from_node_id = $2 AND to_node_id = $3 AND deleted_at IS NULL", moduleID, fromID, toID).Scan(&existingID); err == nil {
				return module.ErrInvalidGraphReference
			}

			_, err = tx.ExecContext(ctx, `INSERT INTO workflow_edges (
				id, module_id, from_node_id, to_node_id, label, condition_text, metadata, sort_order, row_version, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8, 1, CURRENT_TIMESTAMP)`, values...)
			if err != nil {
				return fmt.Errorf("failed to insert workflow edge: %w", err)
			}
		} else if err != nil {
			return err
		} else {
			if !hasVersion || expectedVersion != currentVersion {
				return module.ErrGraphConflict
			}

			values = append(values, expectedVersion)
			result, err := tx.ExecContext(ctx, `UPDATE workflow_edges SET
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
				if err != nil {
					return err
				}
			} else {
				if err := tx.QueryRowContext(ctx, "SELECT row_version FROM workflow_edges WHERE id = $1 AND module_id = $2 AND deleted_at IS NULL", edgeID, moduleID).Scan(&currentVersion); err != nil {
					return err
				}
				if currentVersion != expectedVersion {
					return module.ErrGraphConflict
				}
			}
		}
	}

	return nil
}
