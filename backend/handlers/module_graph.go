package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"backend/db"
	"backend/models"
)

var nonIDChars = regexp.MustCompile(`[^a-zA-Z0-9_]+`)
var slaPattern = regexp.MustCompile(`(?i)(\d+(?:[.,]\d+)?)\s*(menit|minute|minutes|jam|hour|hours|hari|day|days|minggu|week|weeks)`)

const (
	graphConflictCode = "GRAPH_VERSION_CONFLICT"
	graphInvalidCode  = "INVALID_GRAPH_REFERENCE"
)

type graphSyncError struct {
	Code    string
	Message string
}

func (e *graphSyncError) Error() string { return e.Message }

func syncModuleGraph(tx *sql.Tx, moduleID string, nodesValue any, edgesValue any, deletedNodes, deletedEdges []models.GraphDelete) error {
	nodes, err := normalizeJSONArray(nodesValue)
	if err != nil {
		return fmt.Errorf("invalid nodes payload: %w", err)
	}
	edges, err := normalizeJSONArray(edgesValue)
	if err != nil {
		return fmt.Errorf("invalid edges payload: %w", err)
	}

	nodeIDs, err := activeNodeIDs(tx, moduleID)
	if err != nil {
		return err
	}
	for idx, node := range nodes {
		nodeID := stringField(node, "id")
		if nodeID == "" {
			nodeID = "node_" + GenerateUUID()
			node["id"] = nodeID
		}
		nodeIDs[nodeID] = true

		doc := mapField(node, "doc")
		if err := validateBusinessFacet(doc); err != nil {
			return err
		}
		slaValue, slaUnit := parseSLA(stringField(doc, "sla"))
		expectedVersion, hasVersion := rowVersion(node)
		if err := upsertNode(tx, moduleID, nodeID, idx, node, doc, slaValue, slaUnit, expectedVersion, hasVersion); err != nil {
			return err
		}
		roles := mapField(node, "roles")
		if err := validateRoleURLs(roles); err != nil {
			return err
		}
		if err := syncRoleFacet(tx, nodeID, "uiux", mapField(roles, "uiux")); err != nil {
			return err
		}
		if err := syncRoleFacet(tx, nodeID, "frontend", mapField(roles, "frontend")); err != nil {
			return err
		}
		if err := syncRoleFacet(tx, nodeID, "backend", mapField(roles, "backend")); err != nil {
			return err
		}
	}

	for idx, edge := range edges {
		fromID := stringField(edge, "from")
		toID := stringField(edge, "to")
		if fromID == "" || toID == "" || fromID == toID || !nodeIDs[fromID] || !nodeIDs[toID] {
			return &graphSyncError{Code: graphInvalidCode, Message: "edge source and target must be active nodes in the same module"}
		}

		edgeID := stringField(edge, "id")
		if edgeID == "" {
			edgeID = "edge_" + GenerateUUID()
		}
		expectedVersion, hasVersion := rowVersion(edge)
		if err := upsertEdge(tx, moduleID, edgeID, idx, edge, expectedVersion, hasVersion); err != nil {
			return err
		}
	}
	for _, node := range nodes {
		if err := syncBusinessDetails(tx, moduleID, stringField(node, "id"), mapField(node, "doc")); err != nil {
			return err
		}
	}

	if err := tombstoneEdges(tx, moduleID, deletedEdges); err != nil {
		return err
	}
	if err := tombstoneNodes(tx, moduleID, deletedNodes); err != nil {
		return err
	}

	return nil
}

func validateBusinessFacet(doc map[string]any) error {
	for _, key := range []string{"priority", "riskLevel", "risk_level"} {
		if value := stringField(doc, key); value != "" && !validLevel(value) {
			return &graphSyncError{Code: graphInvalidCode, Message: key + " must be low, medium, high, or critical"}
		}
	}
	if triggerType := firstString(doc, "triggerType", "trigger_type"); triggerType != "" && !oneOf(triggerType, "manual", "event", "schedule", "api") {
		return &graphSyncError{Code: graphInvalidCode, Message: "triggerType is invalid"}
	}
	return nil
}

func validateRoleURLs(roles map[string]any) error {
	for _, roleKey := range []string{"uiux", "frontend", "backend"} {
		facet := mapField(roles, roleKey)
		for _, key := range []string{"link", "figmaFrameUrl", "wireframeUrl", "handoffLink"} {
			value := stringField(facet, key)
			if value != "" && !(strings.HasPrefix(value, "https://") || strings.HasPrefix(value, "http://")) {
				return &graphSyncError{Code: graphInvalidCode, Message: key + " must be an absolute URL"}
			}
		}
	}
	return nil
}

func validLevel(value string) bool { return oneOf(value, "low", "medium", "high", "critical") }
func oneOf(value string, allowed ...string) bool {
	for _, item := range allowed {
		if strings.EqualFold(strings.TrimSpace(value), item) {
			return true
		}
	}
	return false
}

// syncBusinessDetails keeps repeatable business rules and decision outcomes normalized.
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

func arrayMaps(value any) []map[string]any {
	bytes, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	var values []map[string]any
	if json.Unmarshal(bytes, &values) != nil {
		return nil
	}
	return values
}

func rulesText(doc map[string]any) string {
	if raw, ok := doc["rules"].(string); ok && strings.TrimSpace(raw) != "" {
		return strings.TrimSpace(raw)
	}
	rules := arrayMaps(doc["rules"])
	parts := make([]string, 0, len(rules))
	for _, rule := range rules {
		if description := stringField(rule, "description"); description != "" {
			parts = append(parts, description)
		}
	}
	return strings.Join(parts, "\n")
}

func normalizeSeverity(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "low", "medium", "high", "critical":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return "medium"
	}
}

func activeNodeIDs(tx *sql.Tx, moduleID string) (map[string]bool, error) {
	rows, err := tx.Query("SELECT id FROM workflow_nodes WHERE module_id = $1 AND deleted_at IS NULL", moduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := map[string]bool{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids[id] = true
	}
	return ids, rows.Err()
}

func upsertNode(tx *sql.Tx, moduleID, nodeID string, idx int, node, doc map[string]any, slaValue, slaUnit any, expectedVersion int, hasVersion bool) error {
	metadata := jsonText(map[string]any{"source": "frontend_graph", "raw": graphPayload(node)})
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
			nullableString(stringField(doc, "rules")),
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

func rowVersion(value map[string]any) (int, bool) {
	raw, ok := value["rowVersion"]
	if !ok || raw == nil {
		return 0, false
	}
	version := int(floatField(map[string]any{"version": raw}, "version"))
	return version, version > 0
}

func graphPayload(value map[string]any) map[string]any {
	payload := make(map[string]any, len(value))
	for key, item := range value {
		if key != "rowVersion" && key != "deletedAt" {
			payload[key] = item
		}
	}
	return payload
}

func syncRoleFacet(tx *sql.Tx, nodeID, roleKey string, facet map[string]any) error {
	if len(facet) == 0 {
		return nil
	}

	assigneeLabel := stringField(facet, "assignee")
	assigneeID, err := resolveAssigneeID(tx, assigneeLabel)
	if err != nil {
		return err
	}
	reviewerID, err := resolveAssigneeID(tx, stringField(facet, "reviewer"))
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO node_role_tasks (
			id, node_id, role_key, assignee_id, assignee_label, reviewer_id, status, readiness, due_date, notes, metadata, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11::jsonb, CURRENT_TIMESTAMP)
		ON CONFLICT (node_id, role_key) DO UPDATE SET
			assignee_id = EXCLUDED.assignee_id,
			assignee_label = EXCLUDED.assignee_label,
			reviewer_id = EXCLUDED.reviewer_id,
			status = EXCLUDED.status,
			readiness = EXCLUDED.readiness,
			due_date = EXCLUDED.due_date,
			notes = EXCLUDED.notes,
			metadata = EXCLUDED.metadata,
			updated_at = CURRENT_TIMESTAMP
	`,
		roleTaskID(nodeID, roleKey),
		nodeID,
		roleKey,
		assigneeID,
		nullableString(assigneeLabel),
		reviewerID,
		normalizeStatus(stringField(facet, "status")),
		normalizeStatus(firstString(facet, "readiness", "status")),
		nullableDate(firstString(facet, "dueDate", "due_date")),
		nullableString(firstString(facet, "notes", "note")),
		jsonText(map[string]any{"raw": facet}),
	)
	if err != nil {
		return err
	}

	switch roleKey {
	case "uiux":
		_, err = tx.Exec(`
			INSERT INTO node_uiux_specs (
				node_id, screen_name, persona, prototype_url, wireframe_url, state_notes,
				accessibility_notes, user_goal, surface, figma_frame_url, design_version, screen_states, interactions, content_messages, responsive_intent, metadata, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16::jsonb, CURRENT_TIMESTAMP)
			ON CONFLICT (node_id) DO UPDATE SET
				screen_name = EXCLUDED.screen_name,
				persona = EXCLUDED.persona,
				prototype_url = EXCLUDED.prototype_url,
				wireframe_url = EXCLUDED.wireframe_url,
				state_notes = EXCLUDED.state_notes,
				accessibility_notes = EXCLUDED.accessibility_notes,
				user_goal = EXCLUDED.user_goal, surface = EXCLUDED.surface, figma_frame_url = EXCLUDED.figma_frame_url, design_version = EXCLUDED.design_version,
				screen_states = EXCLUDED.screen_states, interactions = EXCLUDED.interactions, content_messages = EXCLUDED.content_messages, responsive_intent = EXCLUDED.responsive_intent,
				metadata = EXCLUDED.metadata,
				updated_at = CURRENT_TIMESTAMP
		`,
			nodeID,
			nullableString(firstString(facet, "screen", "screenName", "screen_name")),
			nullableString(firstString(facet, "persona")),
			nullableString(firstString(facet, "prototypeUrl", "prototype_url", "link")),
			nullableString(firstString(facet, "wireframeUrl", "wireframe_url")),
			nullableString(firstString(facet, "stateNotes", "state_notes")),
			nullableString(firstString(facet, "accessibilityNotes", "accessibility_notes")),
			nullableString(firstString(facet, "userGoal", "user_goal")), nullableString(stringField(facet, "surface")), nullableString(firstString(facet, "figmaFrameUrl", "figma_frame_url", "link")), nullableString(firstString(facet, "designVersion", "design_version")),
			nullableString(firstString(facet, "screenStates", "screen_states", "stateNotes")), nullableString(stringField(facet, "interactions")), nullableString(firstString(facet, "contentMessages", "content_messages")), nullableString(firstString(facet, "responsiveIntent", "responsive_intent")),
			jsonText(map[string]any{"raw": facet}),
		)
	case "frontend":
		_, err = tx.Exec(`
			INSERT INTO node_frontend_specs (
				node_id, page_name, route_path, interaction_notes, validation_notes,
				state_handling, handoff_url, experience_name, entry_exit_behavior, input_requirements, api_references, analytics_intent, feature_availability, metadata, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14::jsonb, CURRENT_TIMESTAMP)
			ON CONFLICT (node_id) DO UPDATE SET
				page_name = EXCLUDED.page_name,
				route_path = EXCLUDED.route_path,
				interaction_notes = EXCLUDED.interaction_notes,
				validation_notes = EXCLUDED.validation_notes,
				state_handling = EXCLUDED.state_handling,
				handoff_url = EXCLUDED.handoff_url,
				experience_name = EXCLUDED.experience_name, entry_exit_behavior = EXCLUDED.entry_exit_behavior, input_requirements = EXCLUDED.input_requirements,
				api_references = EXCLUDED.api_references, analytics_intent = EXCLUDED.analytics_intent, feature_availability = EXCLUDED.feature_availability,
				metadata = EXCLUDED.metadata,
				updated_at = CURRENT_TIMESTAMP
		`,
			nodeID,
			nullableString(firstString(facet, "page", "pageName", "page_name", "component")),
			nullableString(firstString(facet, "route", "routePath", "route_path")),
			nullableString(firstString(facet, "interaction", "interactionNotes", "interaction_notes")),
			nullableString(firstString(facet, "validation", "validationNotes", "validation_notes")),
			nullableString(firstString(facet, "state", "stateHandling", "state_handling")),
			nullableString(firstString(facet, "handoffLink", "handoffUrl", "handoff_url", "link")),
			nullableString(firstString(facet, "experienceName", "experience_name", "page")), nullableString(firstString(facet, "entryExitBehavior", "entry_exit_behavior")), nullableString(firstString(facet, "inputRequirements", "input_requirements")),
			nullableString(firstString(facet, "apiReferences", "api_references")), nullableString(firstString(facet, "analyticsIntent", "analytics_intent")), nullableString(firstString(facet, "featureAvailability", "feature_availability")),
			jsonText(map[string]any{"raw": facet}),
		)
	case "backend":
		requestJSON := jsonbFromText(stringField(facet, "request"))
		responseJSON := jsonbFromText(stringField(facet, "response"))
		errorJSON := jsonbFromText(firstString(facet, "errorCodes", "error_codes"))
		_, err = tx.Exec(`
			INSERT INTO node_api_contracts (
				node_id, method, endpoint_path, auth_policy, request_example, response_example,
				status_code, error_codes, curl_example, service_capability, api_references, business_validation, dependency_references, idempotency_notes, caching_notes, security_notes, observability_intent, sla_value, sla_unit, metadata, updated_at
			) VALUES ($1, $2, $3, $4, $5::jsonb, $6::jsonb, $7, $8::jsonb, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21::jsonb, CURRENT_TIMESTAMP)
			ON CONFLICT (node_id) DO UPDATE SET
				method = EXCLUDED.method,
				endpoint_path = EXCLUDED.endpoint_path,
				auth_policy = EXCLUDED.auth_policy,
				request_example = EXCLUDED.request_example,
				response_example = EXCLUDED.response_example,
				status_code = EXCLUDED.status_code,
				error_codes = EXCLUDED.error_codes,
				curl_example = EXCLUDED.curl_example,
				service_capability = EXCLUDED.service_capability, api_references = EXCLUDED.api_references, business_validation = EXCLUDED.business_validation, dependency_references = EXCLUDED.dependency_references,
				idempotency_notes = EXCLUDED.idempotency_notes, caching_notes = EXCLUDED.caching_notes, security_notes = EXCLUDED.security_notes, observability_intent = EXCLUDED.observability_intent,
				sla_value = EXCLUDED.sla_value, sla_unit = EXCLUDED.sla_unit,
				metadata = EXCLUDED.metadata,
				updated_at = CURRENT_TIMESTAMP
		`,
			nodeID,
			nullableString(normalizeMethod(stringField(facet, "method"))),
			nullableString(stringField(facet, "endpoint")),
			nullableString(firstString(facet, "auth", "authPolicy", "auth_policy")),
			requestJSON,
			responseJSON,
			nullableString(firstString(facet, "statusCode", "status_code")),
			errorJSON,
			nullableString(stringField(facet, "curl")),
			nullableString(firstString(facet, "serviceCapability", "service_capability")), nullableString(firstString(facet, "apiReferences", "api_references")), nullableString(firstString(facet, "businessValidation", "business_validation")), nullableString(firstString(facet, "dependencyReferences", "dependency_references")),
			nullableString(firstString(facet, "idempotencyNotes", "idempotency_notes")), nullableString(firstString(facet, "cachingNotes", "caching_notes")), nullableString(firstString(facet, "securityNotes", "security_notes")), nullableString(firstString(facet, "observabilityIntent", "observability_intent")),
			func() any { value, _ := parseSLA(stringField(facet, "sla")); return value }(), func() any { _, unit := parseSLA(stringField(facet, "sla")); return unit }(),
			jsonText(map[string]any{"raw": facet}),
		)
	}

	return err
}

func hydrateModuleGraph(module *models.Module) error {
	rows, err := db.DB.Query(`
		SELECT id, type, label, x, y, actor, trigger, input_desc, process_desc,
			output_desc, business_rules, exception_path, system_context, sla_value,
			sla_unit, priority, risk_level, acceptance_criteria, outcome, trigger_type, preconditions, reference_links, row_version
		FROM workflow_nodes
		WHERE module_id = $1 AND deleted_at IS NULL
		ORDER BY sort_order ASC, created_at ASC
	`, module.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	nodes := []map[string]any{}
	nodeIndex := map[string]map[string]any{}
	for rows.Next() {
		var id, nodeType, label string
		var x, y float64
		var actor, trigger, input, process, output, rules, exceptionPath, systemContext sql.NullString
		var slaValue sql.NullFloat64
		var slaUnit, priority, riskLevel, acceptanceCriteria, outcome, triggerType, preconditions, referenceLinks sql.NullString
		var rowVersion int
		if err := rows.Scan(&id, &nodeType, &label, &x, &y, &actor, &trigger, &input, &process, &output, &rules, &exceptionPath, &systemContext, &slaValue, &slaUnit, &priority, &riskLevel, &acceptanceCriteria, &outcome, &triggerType, &preconditions, &referenceLinks, &rowVersion); err != nil {
			return err
		}

		doc := map[string]any{
			"actor":   nullStringValue(actor),
			"input":   nullStringValue(input),
			"process": nullStringValue(process),
			"output":  nullStringValue(output),
			"rules":   nullStringValue(rules),
			"system":  nullStringValue(systemContext),
			"sla":     formatSLA(slaValue, slaUnit),
		}
		putIfString(doc, "trigger", trigger)
		putIfString(doc, "exceptionPath", exceptionPath)
		putIfString(doc, "priority", priority)
		putIfString(doc, "riskLevel", riskLevel)
		putIfString(doc, "acceptanceCriteria", acceptanceCriteria)
		putIfString(doc, "outcome", outcome)
		putIfString(doc, "triggerType", triggerType)
		putIfString(doc, "preconditions", preconditions)
		putIfString(doc, "referenceLinks", referenceLinks)

		node := map[string]any{
			"id":         id,
			"type":       nodeType,
			"label":      label,
			"x":          x,
			"y":          y,
			"rowVersion": rowVersion,
			"doc":        doc,
			"roles":      map[string]any{},
		}
		nodes = append(nodes, node)
		nodeIndex[id] = node
	}

	if len(nodes) == 0 {
		return nil
	}
	if err := hydrateBusinessDetails(module.ID, nodeIndex); err != nil {
		return err
	}

	if err := hydrateRoleTasks(module.ID, nodeIndex); err != nil {
		return err
	}
	if err := hydrateUiuxSpecs(module.ID, nodeIndex); err != nil {
		return err
	}
	if err := hydrateFrontendSpecs(module.ID, nodeIndex); err != nil {
		return err
	}
	if err := hydrateBackendContracts(module.ID, nodeIndex); err != nil {
		return err
	}

	edgeRows, err := db.DB.Query(`
		SELECT id, from_node_id, to_node_id, label, condition_text, row_version
		FROM workflow_edges
		WHERE module_id = $1 AND deleted_at IS NULL
		ORDER BY sort_order ASC, created_at ASC
	`, module.ID)
	if err != nil {
		return err
	}
	defer edgeRows.Close()

	edges := []map[string]any{}
	for edgeRows.Next() {
		var id, fromID, toID string
		var label, condition sql.NullString
		var rowVersion int
		if err := edgeRows.Scan(&id, &fromID, &toID, &label, &condition, &rowVersion); err != nil {
			return err
		}
		edge := map[string]any{
			"id":         id,
			"from":       fromID,
			"to":         toID,
			"rowVersion": rowVersion,
		}
		putIfString(edge, "label", label)
		putIfString(edge, "condition", condition)
		edges = append(edges, edge)
	}

	nodesJSON, _ := json.Marshal(nodes)
	edgesJSON, _ := json.Marshal(edges)
	module.Nodes = string(nodesJSON)
	module.Edges = string(edgesJSON)
	return nil
}

func hydrateBusinessDetails(moduleID string, nodeIndex map[string]map[string]any) error {
	rules, err := db.DB.Query(`SELECT r.node_id, COALESCE(r.rule_code, ''), r.severity, r.description FROM node_business_rules r JOIN workflow_nodes n ON n.id = r.node_id WHERE n.module_id = $1 ORDER BY r.sort_order`, moduleID)
	if err != nil {
		return err
	}
	defer rules.Close()
	for rules.Next() {
		var nodeID, code, severity, description string
		if err := rules.Scan(&nodeID, &code, &severity, &description); err != nil {
			return err
		}
		node := nodeIndex[nodeID]
		doc, _ := node["doc"].(map[string]any)
		current, _ := doc["rules"].([]map[string]any)
		doc["rules"] = append(current, map[string]any{"code": code, "severity": severity, "description": description})
	}
	if err := rules.Err(); err != nil {
		return err
	}
	outcomes, err := db.DB.Query(`SELECT o.node_id, o.outcome, COALESCE(o.edge_id, '') FROM node_decision_outcomes o JOIN workflow_nodes n ON n.id = o.node_id WHERE n.module_id = $1 ORDER BY o.sort_order`, moduleID)
	if err != nil {
		return err
	}
	defer outcomes.Close()
	for outcomes.Next() {
		var nodeID, outcome, edgeID string
		if err := outcomes.Scan(&nodeID, &outcome, &edgeID); err != nil {
			return err
		}
		node := nodeIndex[nodeID]
		doc, _ := node["doc"].(map[string]any)
		current, _ := doc["decisionOutcomes"].([]map[string]any)
		item := map[string]any{"outcome": outcome}
		if edgeID != "" {
			item["edgeId"] = edgeID
		}
		doc["decisionOutcomes"] = append(current, item)
	}
	return outcomes.Err()
}

// ModuleGraphSnapshotMismatch retains the previous boolean API for callers that
// have not moved to the detailed reconciliation report yet.
func ModuleGraphSnapshotMismatch(moduleID string) (bool, error) {
	report, err := ReconcileModuleGraph(context.Background(), moduleID)
	if err != nil {
		return false, err
	}
	return !report.Matches, nil
}

func hydrateRoleTasks(moduleID string, nodeIndex map[string]map[string]any) error {
	rows, err := db.DB.Query(`
		SELECT t.node_id, t.role_key, COALESCE(u.name, t.assignee_label, '') AS assignee,
			t.status, COALESCE(t.due_date::text, '') AS due_date, COALESCE(t.notes, '') AS notes
		FROM node_role_tasks t
		JOIN workflow_nodes n ON n.id = t.node_id
		LEFT JOIN users u ON u.id = t.assignee_id
		WHERE n.module_id = $1
	`, moduleID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var nodeID, roleKey, assignee, status, dueDate, notes string
		if err := rows.Scan(&nodeID, &roleKey, &assignee, &status, &dueDate, &notes); err != nil {
			return err
		}
		role := roleMap(nodeIndex, nodeID, roleKey)
		role["assignee"] = assignee
		role["status"] = status
		if dueDate != "" {
			role["dueDate"] = dueDate
		}
		if notes != "" {
			role["notes"] = notes
		}
	}
	return nil
}

func hydrateUiuxSpecs(moduleID string, nodeIndex map[string]map[string]any) error {
	rows, err := db.DB.Query(`
		SELECT s.node_id, COALESCE(s.screen_name, ''), COALESCE(s.prototype_url, ''),
			COALESCE(s.wireframe_url, ''), COALESCE(s.state_notes, ''), COALESCE(s.accessibility_notes, ''), COALESCE(s.user_goal, ''), COALESCE(s.surface, ''), COALESCE(s.figma_frame_url, ''), COALESCE(s.design_version, ''), COALESCE(s.screen_states, ''), COALESCE(s.interactions, ''), COALESCE(s.content_messages, ''), COALESCE(s.responsive_intent, '')
		FROM node_uiux_specs s
		JOIN workflow_nodes n ON n.id = s.node_id
		WHERE n.module_id = $1
	`, moduleID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var nodeID, screen, prototype, wireframe, stateNotes, accessibility, userGoal, surface, figmaFrame, designVersion, screenStates, interactions, contentMessages, responsiveIntent string
		if err := rows.Scan(&nodeID, &screen, &prototype, &wireframe, &stateNotes, &accessibility, &userGoal, &surface, &figmaFrame, &designVersion, &screenStates, &interactions, &contentMessages, &responsiveIntent); err != nil {
			return err
		}
		role := roleMap(nodeIndex, nodeID, "uiux")
		role["screen"] = screen
		role["link"] = prototype
		if wireframe != "" {
			role["wireframeUrl"] = wireframe
		}
		if stateNotes != "" {
			role["stateNotes"] = stateNotes
		}
		if accessibility != "" {
			role["accessibilityNotes"] = accessibility
		}
		role["userGoal"], role["surface"], role["figmaFrameUrl"], role["designVersion"] = userGoal, surface, figmaFrame, designVersion
		role["screenStates"], role["interactions"], role["contentMessages"], role["responsiveIntent"] = screenStates, interactions, contentMessages, responsiveIntent
	}
	return nil
}

func hydrateFrontendSpecs(moduleID string, nodeIndex map[string]map[string]any) error {
	rows, err := db.DB.Query(`
		SELECT s.node_id, COALESCE(s.page_name, ''), COALESCE(s.route_path, ''),
			COALESCE(s.interaction_notes, ''), COALESCE(s.validation_notes, ''),
			COALESCE(s.state_handling, ''), COALESCE(s.handoff_url, ''), COALESCE(s.experience_name, ''), COALESCE(s.entry_exit_behavior, ''), COALESCE(s.input_requirements, ''), COALESCE(s.api_references, ''), COALESCE(s.analytics_intent, ''), COALESCE(s.feature_availability, '')
		FROM node_frontend_specs s
		JOIN workflow_nodes n ON n.id = s.node_id
		WHERE n.module_id = $1
	`, moduleID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var nodeID, page, route, interaction, validation, stateHandling, handoff, experience, entryExit, inputs, apiRefs, analytics, availability string
		if err := rows.Scan(&nodeID, &page, &route, &interaction, &validation, &stateHandling, &handoff, &experience, &entryExit, &inputs, &apiRefs, &analytics, &availability); err != nil {
			return err
		}
		role := roleMap(nodeIndex, nodeID, "frontend")
		role["page"] = page
		role["component"] = page
		role["route"] = route
		role["interaction"] = interaction
		role["validation"] = validation
		role["state"] = stateHandling
		role["handoffLink"] = handoff
		role["link"] = handoff
		role["experienceName"], role["entryExitBehavior"], role["inputRequirements"] = experience, entryExit, inputs
		role["apiReferences"], role["analyticsIntent"], role["featureAvailability"] = apiRefs, analytics, availability
	}
	return nil
}

func hydrateBackendContracts(moduleID string, nodeIndex map[string]map[string]any) error {
	rows, err := db.DB.Query(`
		SELECT c.node_id, COALESCE(c.method, ''), COALESCE(c.endpoint_path, ''),
			COALESCE(c.auth_policy, ''), COALESCE(c.request_example::text, ''),
			COALESCE(c.response_example::text, ''), COALESCE(c.status_code, ''),
			COALESCE(c.error_codes::text, ''), COALESCE(c.service_capability, ''), COALESCE(c.api_references, ''), COALESCE(c.business_validation, ''), COALESCE(c.dependency_references, ''), COALESCE(c.idempotency_notes, ''), COALESCE(c.caching_notes, ''), COALESCE(c.security_notes, ''), COALESCE(c.observability_intent, ''), c.sla_value, COALESCE(c.sla_unit, '')
		FROM node_api_contracts c
		JOIN workflow_nodes n ON n.id = c.node_id
		WHERE n.module_id = $1
	`, moduleID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var nodeID, method, endpoint, auth, requestBody, responseBody, statusCode, errorCodes, capability, apiRefs, validation, dependencies, idempotency, caching, security, observability, slaUnit string
		var slaValue sql.NullFloat64
		if err := rows.Scan(&nodeID, &method, &endpoint, &auth, &requestBody, &responseBody, &statusCode, &errorCodes, &capability, &apiRefs, &validation, &dependencies, &idempotency, &caching, &security, &observability, &slaValue, &slaUnit); err != nil {
			return err
		}
		role := roleMap(nodeIndex, nodeID, "backend")
		role["method"] = method
		role["endpoint"] = endpoint
		role["auth"] = auth
		role["request"] = jsonbDisplay(requestBody)
		role["response"] = jsonbDisplay(responseBody)
		role["statusCode"] = statusCode
		if errorCodes != "" {
			role["errorCodes"] = jsonbDisplay(errorCodes)
		}
		role["serviceCapability"], role["apiReferences"], role["businessValidation"], role["dependencyReferences"] = capability, apiRefs, validation, dependencies
		role["idempotencyNotes"], role["cachingNotes"], role["securityNotes"], role["observabilityIntent"] = idempotency, caching, security, observability
		role["sla"] = formatSLA(slaValue, sql.NullString{String: slaUnit, Valid: slaUnit != ""})
	}
	return nil
}

func normalizeJSONArray(value any) ([]map[string]any, error) {
	if value == nil {
		return []map[string]any{}, nil
	}
	bytes, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var items []map[string]any
	if err := json.Unmarshal(bytes, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func resolveAssigneeID(tx *sql.Tx, label string) (any, error) {
	label = strings.TrimSpace(label)
	if label == "" {
		return nil, nil
	}

	var id string
	err := tx.QueryRow("SELECT id FROM users WHERE (id = $1 OR LOWER(name) = LOWER($1) OR LOWER(email) = LOWER($1)) AND status = 'active' ORDER BY created_at ASC LIMIT 1", label).Scan(&id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return id, nil
}

func roleTaskID(nodeID, roleKey string) string {
	base := nonIDChars.ReplaceAllString(nodeID+"_"+roleKey, "_")
	if len(base) > 145 {
		base = base[:145]
	}
	return "task_" + base
}

func roleMap(nodeIndex map[string]map[string]any, nodeID, roleKey string) map[string]any {
	node := nodeIndex[nodeID]
	roles, _ := node["roles"].(map[string]any)
	if roles == nil {
		roles = map[string]any{}
		node["roles"] = roles
	}
	role, _ := roles[roleKey].(map[string]any)
	if role == nil {
		role = map[string]any{}
		roles[roleKey] = role
	}
	return role
}

func mapField(m map[string]any, key string) map[string]any {
	val, ok := m[key]
	if !ok || val == nil {
		return map[string]any{}
	}
	if typed, ok := val.(map[string]any); ok {
		return typed
	}
	bytes, err := json.Marshal(val)
	if err != nil {
		return map[string]any{}
	}
	out := map[string]any{}
	_ = json.Unmarshal(bytes, &out)
	return out
}

func stringField(m map[string]any, key string) string {
	val, ok := m[key]
	if !ok || val == nil {
		return ""
	}
	switch typed := val.(type) {
	case string:
		return strings.TrimSpace(typed)
	case fmt.Stringer:
		return strings.TrimSpace(typed.String())
	default:
		return strings.TrimSpace(fmt.Sprint(typed))
	}
}

func firstString(m map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := stringField(m, key); value != "" {
			return value
		}
	}
	return ""
}

func floatField(m map[string]any, key string) float64 {
	val, ok := m[key]
	if !ok || val == nil {
		return 0
	}
	switch typed := val.(type) {
	case float64:
		return typed
	case float32:
		return float64(typed)
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	case json.Number:
		out, _ := typed.Float64()
		return out
	case string:
		out, _ := strconv.ParseFloat(strings.TrimSpace(typed), 64)
		return out
	default:
		return 0
	}
}

func normalizeNodeType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "terminator", "process", "decision", "actor", "system":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return "process"
	}
}

func normalizeStatus(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "planned", "in_progress", "review", "done":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return "planned"
	}
}

func normalizeMethod(value string) string {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "GET", "POST", "PUT", "PATCH", "DELETE":
		return strings.ToUpper(strings.TrimSpace(value))
	default:
		return "GET"
	}
}

func parseSLA(value string) (any, any) {
	match := slaPattern.FindStringSubmatch(value)
	if len(match) < 3 {
		return nil, nil
	}
	number, err := strconv.ParseFloat(strings.ReplaceAll(match[1], ",", "."), 64)
	if err != nil {
		return nil, nil
	}
	unit := strings.ToLower(match[2])
	switch unit {
	case "minute", "minutes":
		unit = "menit"
	case "hour", "hours":
		unit = "jam"
	case "day", "days":
		unit = "hari"
	case "week", "weeks":
		unit = "minggu"
	}
	return number, unit
}

func formatSLA(value sql.NullFloat64, unit sql.NullString) string {
	if !value.Valid || !unit.Valid || unit.String == "" {
		return ""
	}
	if value.Float64 == float64(int64(value.Float64)) {
		return fmt.Sprintf("%d %s", int64(value.Float64), unit.String)
	}
	return fmt.Sprintf("%.1f %s", value.Float64, unit.String)
}

func nullableString(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return value
}

func nullableDate(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil
	}
	return parsed
}

func defaultString(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func nullStringValue(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func putIfString(target map[string]any, key string, value sql.NullString) {
	if value.Valid && strings.TrimSpace(value.String) != "" {
		target[key] = value.String
	}
}

func jsonText(value any) string {
	bytes, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(bytes)
}

func jsonbFromText(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "null"
	}
	var parsed any
	if err := json.Unmarshal([]byte(value), &parsed); err == nil {
		bytes, _ := json.Marshal(parsed)
		return string(bytes)
	}
	return jsonText(map[string]any{"raw": value})
}

func jsonbDisplay(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || value == "null" {
		return ""
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(value), &parsed); err == nil {
		if raw, ok := parsed["raw"].(string); ok {
			return raw
		}
	}
	var anyValue any
	if err := json.Unmarshal([]byte(value), &anyValue); err != nil {
		return value
	}
	bytes, _ := json.MarshalIndent(anyValue, "", "  ")
	return string(bytes)
}
