package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func validateBusinessFacet(doc map[string]any) error {
	for _, key := range []string{"priority", "riskLevel", "risk_level"} {
		if value := stringField(doc, key); value != "" && !validLevel(value) {
			return &graphSyncError{Code: graphInvalidCode, Message: key + " must be low, medium, high, or critical"}
		}
	}
	if triggerType := firstString(doc, "triggerType", "trigger_type"); triggerType != "" && !oneOf(triggerType, "manual", "event", "schedule", "api") {
		return &graphSyncError{Code: graphInvalidCode, Message: "triggerType is invalid"}
	}
	if sla := stringField(doc, "sla"); sla != "" {
		value, _ := parseSLA(sla)
		if !exactSLA.MatchString(sla) || value == nil || value.(float64) <= 0 {
			return specError("doc.sla", "must use a positive number and supported unit")
		}
	}
	return nil
}

func validateRoleURLs(roles map[string]any) error {
	for _, roleKey := range []string{"uiux", "frontend", "backend"} {
		facet := mapField(roles, roleKey)
		for _, key := range []string{"link", "figmaFrameUrl", "wireframeUrl", "handoffLink", "prototypeUrl"} {
			value := stringField(facet, key)
			if value != "" && !validSpecURL(value) {
				return &graphSyncError{Code: graphInvalidCode, Message: key + " must be an absolute HTTP(S) URL without credentials"}
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

// ModuleGraphSnapshotMismatch retains the previous boolean API for callers.
func ModuleGraphSnapshotMismatch(moduleID string) (bool, error) {
	report, err := ReconcileModuleGraph(context.Background(), moduleID)
	if err != nil {
		return false, err
	}
	return !report.Matches, nil
}
