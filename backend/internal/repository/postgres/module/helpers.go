package module

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var nonIDChars = regexp.MustCompile(`[^a-zA-Z0-9_]+`)

func generateUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func roleTaskID(nodeID, roleKey string) string {
	base := nonIDChars.ReplaceAllString(nodeID+"_"+roleKey, "_")
	if len(base) > 145 {
		base = base[:145]
	}
	return "task_" + base
}

func stringField(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
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
		if val := stringField(m, key); val != "" {
			return val
		}
	}
	return ""
}

func mapField(m map[string]any, key string) map[string]any {
	if m == nil {
		return map[string]any{}
	}
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

func floatField(m map[string]any, key string) float64 {
	if m == nil {
		return 0
	}
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

func formatSLA(value sql.NullFloat64, unit sql.NullString) string {
	if !value.Valid || !unit.Valid || unit.String == "" {
		return ""
	}
	if value.Float64 == float64(int64(value.Float64)) {
		return fmt.Sprintf("%d %s", int64(value.Float64), unit.String)
	}
	return fmt.Sprintf("%.1f %s", value.Float64, unit.String)
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

func normalizeSeverity(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "low", "medium", "high", "critical":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return "medium"
	}
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
