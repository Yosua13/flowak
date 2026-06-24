package handlers

import (
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

func syncModuleGraph(tx *sql.Tx, moduleID string, nodesValue any, edgesValue any) error {
	nodes, err := normalizeJSONArray(nodesValue)
	if err != nil {
		return fmt.Errorf("invalid nodes payload: %w", err)
	}
	edges, err := normalizeJSONArray(edgesValue)
	if err != nil {
		return fmt.Errorf("invalid edges payload: %w", err)
	}

	if _, err := tx.Exec("DELETE FROM workflow_edges WHERE module_id = $1", moduleID); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM workflow_nodes WHERE module_id = $1", moduleID); err != nil {
		return err
	}

	nodeIDs := map[string]bool{}
	for idx, node := range nodes {
		nodeID := stringField(node, "id")
		if nodeID == "" {
			nodeID = "node_" + GenerateUUID()
		}
		nodeIDs[nodeID] = true

		doc := mapField(node, "doc")
		slaValue, slaUnit := parseSLA(stringField(doc, "sla"))
		metadata := jsonText(map[string]any{"source": "frontend_graph", "raw": node})

		_, err := tx.Exec(`
			INSERT INTO workflow_nodes (
				id, module_id, type, label, x, y, actor, trigger, input_desc, process_desc,
				output_desc, business_rules, exception_path, system_context, sla_value, sla_unit,
				priority, risk_level, acceptance_criteria, metadata, sort_order, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
				$11, $12, $13, $14, $15, $16, $17, $18, $19, $20::jsonb, $21, CURRENT_TIMESTAMP
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
			metadata,
			idx,
		)
		if err != nil {
			return err
		}

		roles := mapField(node, "roles")
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
			continue
		}

		edgeID := stringField(edge, "id")
		if edgeID == "" {
			edgeID = "edge_" + GenerateUUID()
		}

		_, err := tx.Exec(`
			INSERT INTO workflow_edges (
				id, module_id, from_node_id, to_node_id, label, condition_text, metadata, sort_order, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8, CURRENT_TIMESTAMP)
			ON CONFLICT (module_id, from_node_id, to_node_id) DO NOTHING
		`,
			edgeID,
			moduleID,
			fromID,
			toID,
			nullableString(stringField(edge, "label")),
			nullableString(firstString(edge, "condition", "conditionText", "condition_text")),
			jsonText(map[string]any{"source": "frontend_graph", "raw": edge}),
			idx,
		)
		if err != nil {
			return err
		}
	}

	return nil
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

	_, err = tx.Exec(`
		INSERT INTO node_role_tasks (
			id, node_id, role_key, assignee_id, assignee_label, status, due_date, notes, metadata, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb, CURRENT_TIMESTAMP)
		ON CONFLICT (node_id, role_key) DO UPDATE SET
			assignee_id = EXCLUDED.assignee_id,
			assignee_label = EXCLUDED.assignee_label,
			status = EXCLUDED.status,
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
		normalizeStatus(stringField(facet, "status")),
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
				accessibility_notes, metadata, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8::jsonb, CURRENT_TIMESTAMP)
			ON CONFLICT (node_id) DO UPDATE SET
				screen_name = EXCLUDED.screen_name,
				persona = EXCLUDED.persona,
				prototype_url = EXCLUDED.prototype_url,
				wireframe_url = EXCLUDED.wireframe_url,
				state_notes = EXCLUDED.state_notes,
				accessibility_notes = EXCLUDED.accessibility_notes,
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
			jsonText(map[string]any{"raw": facet}),
		)
	case "frontend":
		_, err = tx.Exec(`
			INSERT INTO node_frontend_specs (
				node_id, page_name, route_path, interaction_notes, validation_notes,
				state_handling, handoff_url, metadata, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8::jsonb, CURRENT_TIMESTAMP)
			ON CONFLICT (node_id) DO UPDATE SET
				page_name = EXCLUDED.page_name,
				route_path = EXCLUDED.route_path,
				interaction_notes = EXCLUDED.interaction_notes,
				validation_notes = EXCLUDED.validation_notes,
				state_handling = EXCLUDED.state_handling,
				handoff_url = EXCLUDED.handoff_url,
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
			jsonText(map[string]any{"raw": facet}),
		)
	case "backend":
		requestJSON := jsonbFromText(stringField(facet, "request"))
		responseJSON := jsonbFromText(stringField(facet, "response"))
		errorJSON := jsonbFromText(firstString(facet, "errorCodes", "error_codes"))
		_, err = tx.Exec(`
			INSERT INTO node_api_contracts (
				node_id, method, endpoint_path, auth_policy, request_example, response_example,
				status_code, error_codes, curl_example, metadata, updated_at
			) VALUES ($1, $2, $3, $4, $5::jsonb, $6::jsonb, $7, $8::jsonb, $9, $10::jsonb, CURRENT_TIMESTAMP)
			ON CONFLICT (node_id) DO UPDATE SET
				method = EXCLUDED.method,
				endpoint_path = EXCLUDED.endpoint_path,
				auth_policy = EXCLUDED.auth_policy,
				request_example = EXCLUDED.request_example,
				response_example = EXCLUDED.response_example,
				status_code = EXCLUDED.status_code,
				error_codes = EXCLUDED.error_codes,
				curl_example = EXCLUDED.curl_example,
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
			jsonText(map[string]any{"raw": facet}),
		)
	}

	return err
}

func hydrateModuleGraph(module *models.Module) error {
	rows, err := db.DB.Query(`
		SELECT id, type, label, x, y, actor, trigger, input_desc, process_desc,
			output_desc, business_rules, exception_path, system_context, sla_value,
			sla_unit, priority, risk_level, acceptance_criteria
		FROM workflow_nodes
		WHERE module_id = $1
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
		var slaUnit, priority, riskLevel, acceptanceCriteria sql.NullString
		if err := rows.Scan(&id, &nodeType, &label, &x, &y, &actor, &trigger, &input, &process, &output, &rules, &exceptionPath, &systemContext, &slaValue, &slaUnit, &priority, &riskLevel, &acceptanceCriteria); err != nil {
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

		node := map[string]any{
			"id":    id,
			"type":  nodeType,
			"label": label,
			"x":     x,
			"y":     y,
			"doc":   doc,
			"roles": map[string]any{},
		}
		nodes = append(nodes, node)
		nodeIndex[id] = node
	}

	if len(nodes) == 0 {
		return nil
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
		SELECT id, from_node_id, to_node_id, label, condition_text
		FROM workflow_edges
		WHERE module_id = $1
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
		if err := edgeRows.Scan(&id, &fromID, &toID, &label, &condition); err != nil {
			return err
		}
		edge := map[string]any{
			"id":   id,
			"from": fromID,
			"to":   toID,
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
			COALESCE(s.wireframe_url, ''), COALESCE(s.state_notes, ''), COALESCE(s.accessibility_notes, '')
		FROM node_uiux_specs s
		JOIN workflow_nodes n ON n.id = s.node_id
		WHERE n.module_id = $1
	`, moduleID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var nodeID, screen, prototype, wireframe, stateNotes, accessibility string
		if err := rows.Scan(&nodeID, &screen, &prototype, &wireframe, &stateNotes, &accessibility); err != nil {
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
	}
	return nil
}

func hydrateFrontendSpecs(moduleID string, nodeIndex map[string]map[string]any) error {
	rows, err := db.DB.Query(`
		SELECT s.node_id, COALESCE(s.page_name, ''), COALESCE(s.route_path, ''),
			COALESCE(s.interaction_notes, ''), COALESCE(s.validation_notes, ''),
			COALESCE(s.state_handling, ''), COALESCE(s.handoff_url, '')
		FROM node_frontend_specs s
		JOIN workflow_nodes n ON n.id = s.node_id
		WHERE n.module_id = $1
	`, moduleID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var nodeID, page, route, interaction, validation, stateHandling, handoff string
		if err := rows.Scan(&nodeID, &page, &route, &interaction, &validation, &stateHandling, &handoff); err != nil {
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
	}
	return nil
}

func hydrateBackendContracts(moduleID string, nodeIndex map[string]map[string]any) error {
	rows, err := db.DB.Query(`
		SELECT c.node_id, COALESCE(c.method, ''), COALESCE(c.endpoint_path, ''),
			COALESCE(c.auth_policy, ''), COALESCE(c.request_example::text, ''),
			COALESCE(c.response_example::text, ''), COALESCE(c.status_code, ''),
			COALESCE(c.error_codes::text, '')
		FROM node_api_contracts c
		JOIN workflow_nodes n ON n.id = c.node_id
		WHERE n.module_id = $1
	`, moduleID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var nodeID, method, endpoint, auth, requestBody, responseBody, statusCode, errorCodes string
		if err := rows.Scan(&nodeID, &method, &endpoint, &auth, &requestBody, &responseBody, &statusCode, &errorCodes); err != nil {
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
