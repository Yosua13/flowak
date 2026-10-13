package handlers

import (
	"database/sql"
	"fmt"
)

func syncBusinessDetails(tx *sql.Tx, moduleID, nodeID string, doc map[string]any) error {
	if _, err := tx.Exec("DELETE FROM node_business_rules WHERE node_id = $1", nodeID); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM node_decision_outcomes WHERE node_id = $1", nodeID); err != nil {
		return err
	}
	for idx, rule := range arrayMaps(doc["rules"]) {
		if _, err := tx.Exec(`INSERT INTO node_business_rules (id,node_id,rule_code,severity,description,sort_order) VALUES ($1,$2,$3,$4,$5,$6)`,
			fmt.Sprintf("rule_%s_%d", nodeID, idx), nodeID, nullableString(stringField(rule, "code")), normalizeSeverity(stringField(rule, "severity")), defaultString(stringField(rule, "description"), "Rule"), idx); err != nil {
			return err
		}
	}
	for idx, outcome := range arrayMaps(doc["decisionOutcomes"]) {
		edgeID := firstString(outcome, "edgeId", "edge_id")
		if edgeID != "" {
			var exists bool
			if err := tx.QueryRow("SELECT EXISTS(SELECT 1 FROM workflow_edges WHERE id = $1 AND module_id = $2 AND deleted_at IS NULL)", edgeID, moduleID).Scan(&exists); err != nil {
				return err
			}
			if !exists {
				return &graphSyncError{Code: graphInvalidCode, Message: "decision outcome references an invalid edge"}
			}
		}
		if _, err := tx.Exec(`INSERT INTO node_decision_outcomes (id,node_id,edge_id,outcome,sort_order) VALUES ($1,$2,$3,$4,$5)`,
			fmt.Sprintf("outcome_%s_%d", nodeID, idx), nodeID, nullableString(edgeID), defaultString(stringField(outcome, "outcome"), "Outcome"), idx); err != nil {
			return err
		}
	}
	return nil
}
