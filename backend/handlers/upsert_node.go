package handlers

import (
	"database/sql"
)

func upsertNode(tx *sql.Tx, moduleID, nodeID string, idx int, node, doc map[string]any, slaValue, slaUnit any, expectedVersion int, hasVersion bool) error {
	metadata := jsonText(map[string]any{"source": "frontend_graph", "legacy_notes": mapField(node, "legacyNotes")})
	var currentVersion int
	err := tx.QueryRow("SELECT row_version FROM workflow_nodes WHERE id = $1 AND module_id = $2 AND deleted_at IS NULL", nodeID, moduleID).Scan(&currentVersion)
	if err == sql.ErrNoRows {
		var foreignModule string
		err = tx.QueryRow("SELECT module_id FROM workflow_nodes WHERE id = $1", nodeID).Scan(&foreignModule)
		if err == nil {
			return &graphSyncError{Code: graphInvalidCode, Message: "node belongs to another module or is deleted"}
		}
		if err != sql.ErrNoRows {
			return err
		}
		_, err = tx.Exec(`
			INSERT INTO workflow_nodes (
				id, module_id, type, label, x, y, actor, trigger, input_desc, process_desc,
				output_desc, business_rules, exception_path, system_context, sla_value, sla_unit,
				priority, risk_level, acceptance_criteria, outcome, trigger_type, preconditions, reference_links, metadata, sort_order, row_version, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
				$11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24::jsonb, $25, 1, CURRENT_TIMESTAMP
			)
		`,
			nodeID,
			moduleID,
			normalizeNodeType(stringField(node, "type")),
			defaultString(stringField(node, "label"), "Langkah Alur"),
			floatField(node, "x"),
			floatField(node, "y"),
			nullableString(stringField(doc, "actor")),
			nullableString(firstString(doc, "trigger", "event")),
			nullableString(stringField(doc, "input")),
			nullableString(stringField(doc, "process")),
			nullableString(stringField(doc, "output")),
			nullableString(rulesText(doc)),
			nullableString(firstString(doc, "exceptionPath", "exception_path")),
			nullableString(stringField(doc, "system")),
			slaValue,
			slaUnit,
			defaultString(firstString(doc, "priority"), "medium"),
			defaultString(firstString(doc, "riskLevel", "risk_level"), "medium"),
			nullableString(firstString(doc, "acceptanceCriteria", "acceptance_criteria")),
			nullableString(stringField(doc, "outcome")),
			nullableString(firstString(doc, "triggerType", "trigger_type")),
			nullableString(stringField(doc, "preconditions")),
			nullableString(firstString(doc, "referenceLinks", "reference_links")),
			metadata,
			idx,
		)
		return err
	}
	if !hasVersion || expectedVersion != currentVersion {
		return &graphSyncError{Code: graphConflictCode, Message: "node version conflict"}
	}
	result, err := tx.Exec(`UPDATE workflow_nodes SET
		type = $3, label = $4, x = $5, y = $6, actor = $7, trigger = $8, input_desc = $9, process_desc = $10,
		output_desc = $11, business_rules = $12, exception_path = $13, system_context = $14, sla_value = $15, sla_unit = $16,
		priority = $17, risk_level = $18, acceptance_criteria = $19, outcome = $20, trigger_type = $21, preconditions = $22, reference_links = $23, metadata = $24::jsonb, sort_order = $25,
		row_version = row_version + 1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND module_id = $2 AND deleted_at IS NULL AND row_version = $26
		AND (type, label, x, y, actor, trigger, input_desc, process_desc, output_desc, business_rules, exception_path, system_context, sla_value, sla_unit, priority, risk_level, acceptance_criteria, outcome, trigger_type, preconditions, reference_links, metadata, sort_order)
		IS DISTINCT FROM ($3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24::jsonb, $25)`,
		nodeValues(nodeID, moduleID, idx, node, doc, slaValue, slaUnit, metadata, expectedVersion)...)
	if err != nil {
		return err
	}
	changed, err := result.RowsAffected()
	if err != nil || changed > 0 {
		return err
	}
	if err := tx.QueryRow("SELECT row_version FROM workflow_nodes WHERE id = $1 AND module_id = $2 AND deleted_at IS NULL", nodeID, moduleID).Scan(&currentVersion); err != nil {
		return err
	}
	if currentVersion != expectedVersion {
		return &graphSyncError{Code: graphConflictCode, Message: "node version conflict"}
	}
	return nil
}

func nodeValues(nodeID, moduleID string, idx int, node, doc map[string]any, slaValue, slaUnit any, metadata string, expectedVersion int) []any {
	return []any{
		nodeID, moduleID, normalizeNodeType(stringField(node, "type")), defaultString(stringField(node, "label"), "Langkah Alur"), floatField(node, "x"), floatField(node, "y"),
		nullableString(stringField(doc, "actor")), nullableString(firstString(doc, "trigger", "event")), nullableString(stringField(doc, "input")), nullableString(stringField(doc, "process")),
		nullableString(stringField(doc, "output")), nullableString(rulesText(doc)), nullableString(firstString(doc, "exceptionPaths", "exceptionPath", "exception_path")), nullableString(stringField(doc, "system")),
		slaValue, slaUnit, defaultString(firstString(doc, "priority"), "medium"), defaultString(firstString(doc, "riskLevel", "risk_level"), "medium"), nullableString(firstString(doc, "acceptanceCriteria", "acceptance_criteria")),
		nullableString(stringField(doc, "outcome")), nullableString(firstString(doc, "triggerType", "trigger_type")), nullableString(stringField(doc, "preconditions")), nullableString(firstString(doc, "referenceLinks", "reference_links")), metadata, idx, expectedVersion,
	}
}
