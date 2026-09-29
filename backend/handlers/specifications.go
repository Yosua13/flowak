package handlers

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var variableReference = regexp.MustCompile(`^\{\{[A-Z][A-Z0-9_]*\}\}$`)
var secretField = regexp.MustCompile(`(?i)(authorization|cookie|password|secret|token|api[_-]?key)`)
var bearerValue = regexp.MustCompile(`(?i)\bBearer\s+([^\s"']+)`)
var exactSLA = regexp.MustCompile(`(?i)^\s*\d+(?:[.,]\d+)?\s*(menit|minute|minutes|jam|hour|hours|hari|day|days|minggu|week|weeks)\s*$`)
var safeEndpoint = regexp.MustCompile(`^/[A-Za-z0-9/_{}.:%-]*$`)

func specError(field, reason string) error {
	return &graphSyncError{Code: graphInvalidCode, Message: field + " " + reason}
}

func validReference(value string) bool {
	return variableReference.MatchString(strings.TrimSpace(value))
}

func sanitizeSpecValue(value any, path string) (any, error) {
	switch typed := value.(type) {
	case map[string]any:
		clean := make(map[string]any, len(typed))
		for key, child := range typed {
			field := key
			if path != "" {
				field = path + "." + key
			}
			if secretField.MatchString(key) {
				if text, ok := child.(string); !ok || (strings.TrimSpace(text) != "" && !validReference(text)) {
					return nil, specError("specification", "must contain a variable reference, not a secret")
				}
			}
			next, err := sanitizeSpecValue(child, field)
			if err != nil {
				return nil, err
			}
			clean[key] = next
		}
		return clean, nil
	case []any:
		clean := make([]any, len(typed))
		for i, child := range typed {
			next, err := sanitizeSpecValue(child, fmt.Sprintf("%s[%d]", path, i))
			if err != nil {
				return nil, err
			}
			clean[i] = next
		}
		return clean, nil
	case string:
		trimmed := strings.TrimSpace(typed)
		if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
			var nested any
			if json.Unmarshal([]byte(trimmed), &nested) == nil {
				clean, err := sanitizeSpecValue(nested, path)
				if err != nil {
					return nil, err
				}
				encoded, _ := json.Marshal(clean)
				return string(encoded), nil
			}
		}
		if bearerValue.MatchString(typed) && !strings.Contains(typed, "Bearer {{") {
			return nil, specError("specification", "must not contain a raw Bearer credential")
		}
		return typed, nil
	default:
		return value, nil
	}
}

// prepareSpecificationNodes is the save boundary for both normalized rows and
// the deprecated graph snapshot. Never persist the unsanitized request nodes.
func prepareSpecificationNodes(value any) ([]map[string]any, error) {
	nodes, err := normalizeJSONArray(value)
	if err != nil {
		return nil, err
	}
	for _, node := range nodes {
		roles := mapField(node, "roles")
		backend := mapField(roles, "backend")
		if raw := firstString(backend, "auth", "authPolicy", "auth_policy"); raw != "" {
			if validReference(raw) || oneOf(raw, "none", "inherit") {
				backend["auth"] = raw
			} else {
				backend["auth"] = "{{API_TOKEN}}"
			}
			legacy := mapField(node, "legacyNotes")
			if backend["auth"] != raw {
				legacy["backend.auth"] = "Legacy authorization removed; choose a variable reference."
				node["legacyNotes"] = legacy
			}
		}
		delete(backend, "authPolicy")
		delete(backend, "auth_policy")
		delete(backend, "curl")
		if len(backend) > 0 {
			roles["backend"] = backend
			node["roles"] = roles
		}
		clean, err := sanitizeSpecValue(node, "node")
		if err != nil {
			return nil, err
		}
		for key := range node {
			delete(node, key)
		}
		for key, item := range clean.(map[string]any) {
			node[key] = item
		}
		if err := validateSpecificationNode(node); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func prepareSpecificationEdges(value any) ([]map[string]any, error) {
	edges, err := normalizeJSONArray(value)
	if err != nil {
		return nil, err
	}
	for _, edge := range edges {
		if _, err := sanitizeSpecValue(edge, "edge"); err != nil {
			return nil, err
		}
	}
	return edges, nil
}

func validSpecURL(raw string) bool {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || !u.IsAbs() || (u.Scheme != "https" && u.Scheme != "http") || u.Hostname() == "" || u.User != nil || u.Fragment != "" {
		return false
	}
	for key := range u.Query() {
		if secretField.MatchString(key) {
			return false
		}
	}
	return true
}

func validateSpecificationNode(node map[string]any) error {
	if strings.TrimSpace(stringField(node, "label")) == "" {
		return specError("label", "is required")
	}
	doc := mapField(node, "doc")
	if err := validateBusinessFacet(doc); err != nil {
		return err
	}
	for _, key := range []string{"referenceLinks", "reference_links"} {
		if raw := stringField(doc, key); raw != "" {
			for _, item := range strings.Split(raw, "\n") {
				if strings.TrimSpace(item) != "" && !validSpecURL(item) {
					return specError("doc."+key, "must contain absolute HTTP(S) URLs without credentials")
				}
			}
		}
	}
	for _, rule := range arrayMaps(doc["rules"]) {
		if strings.TrimSpace(stringField(rule, "description")) == "" {
			return specError("doc.rules.description", "is required")
		}
		if severity := stringField(rule, "severity"); severity != "" && !validLevel(severity) {
			return specError("doc.rules.severity", "is invalid")
		}
	}
	roles := mapField(node, "roles")
	for key := range roles {
		if key != "uiux" && key != "frontend" && key != "backend" {
			return specError("roles", "contains an invalid facet key")
		}
	}
	if err := validateRoleURLs(roles); err != nil {
		return err
	}
	for _, key := range []string{"uiux", "frontend", "backend"} {
		facet := mapField(roles, key)
		for _, stateKey := range []string{"status", "readiness"} {
			if state := stringField(facet, stateKey); state != "" && !oneOf(state, "planned", "in_progress", "review", "done") {
				return specError("roles."+key+"."+stateKey, "is invalid")
			}
		}
		if sla := stringField(facet, "sla"); sla != "" {
			value, _ := parseSLA(sla)
			if !exactSLA.MatchString(sla) || value == nil || value.(float64) <= 0 {
				return specError("roles."+key+".sla", "must use a positive number and supported unit")
			}
		}
		if oneOf(firstString(facet, "readiness", "status"), "done") {
			required := map[string][]string{"uiux": {"screen", "userGoal"}, "frontend": {"page", "validation"}, "backend": {"serviceCapability", "endpoint"}}[key]
			for _, field := range required {
				if stringField(facet, field) == "" {
					return specError("roles."+key+"."+field, "is required when done")
				}
			}
		}
		if key == "backend" {
			if method := stringField(facet, "method"); method != "" && !oneOf(method, "GET", "POST", "PUT", "PATCH", "DELETE") {
				return specError("roles.backend.method", "is invalid")
			}
			if auth := stringField(facet, "auth"); auth != "" && !validReference(auth) && !oneOf(auth, "none", "inherit") {
				return specError("roles.backend.auth", "must be a variable reference")
			}
			if endpoint := stringField(facet, "endpoint"); endpoint != "" {
				if strings.HasPrefix(endpoint, "/") {
					if strings.HasPrefix(endpoint, "//") || !safeEndpoint.MatchString(endpoint) {
						return specError("roles.backend.endpoint", "must be a relative path without query credentials")
					}
				} else if !validSpecURL(endpoint) || strings.ContainsAny(endpoint, "'\"`$\\\r\n") {
					return specError("roles.backend.endpoint", "must be a relative path or absolute HTTP(S) URL")
				}
			}
		}
	}
	return nil
}

func specificationCompleteness(node map[string]any) int {
	doc, roles := mapField(node, "doc"), mapField(node, "roles")
	checks := []bool{
		stringField(node, "label") != "", stringField(doc, "outcome") != "", stringField(doc, "actor") != "", stringField(doc, "trigger") != "", stringField(doc, "process") != "", len(arrayMaps(doc["rules"])) > 0,
		firstString(mapField(roles, "uiux"), "userGoal", "user_goal") != "", firstString(mapField(roles, "uiux"), "screen", "screenName") != "",
		firstString(mapField(roles, "frontend"), "experienceName", "page") != "", stringField(mapField(roles, "frontend"), "validation") != "",
		firstString(mapField(roles, "backend"), "serviceCapability", "service_capability") != "", stringField(mapField(roles, "backend"), "endpoint") != "",
	}
	completed := 0
	for _, done := range checks {
		if done {
			completed++
		}
	}
	return completed * 100 / len(checks)
}

func scrubLegacySpecification(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		clean := make(map[string]any, len(typed))
		for key, item := range typed {
			if secretField.MatchString(key) {
				clean[key] = "{{REDACTED}}"
			} else if key == "auth" {
				if raw, ok := item.(string); ok && (validReference(raw) || oneOf(raw, "none", "inherit")) {
					clean[key] = raw
				} else {
					clean[key] = "{{API_TOKEN}}"
				}
			} else if key == "curl" {
				clean[key] = "{{REDACTED}}"
			} else {
				clean[key] = scrubLegacySpecification(item)
			}
		}
		return clean
	case []any:
		clean := make([]any, len(typed))
		for i, item := range typed {
			clean[i] = scrubLegacySpecification(item)
		}
		return clean
	case string:
		trimmed := strings.TrimSpace(typed)
		if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
			var nested any
			if json.Unmarshal([]byte(trimmed), &nested) == nil {
				encoded, _ := json.Marshal(scrubLegacySpecification(nested))
				return string(encoded)
			}
		}
		return bearerValue.ReplaceAllString(typed, "Bearer {{REDACTED}}")
	default:
		return value
	}
}

func scrubLegacyDisplay(raw string) string {
	var parsed any
	if json.Unmarshal([]byte(raw), &parsed) != nil {
		return bearerValue.ReplaceAllString(raw, "Bearer {{REDACTED}}")
	}
	encoded, _ := json.Marshal(scrubLegacySpecification(parsed))
	return string(encoded)
}
