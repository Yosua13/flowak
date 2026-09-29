package handlers

import (
	"strings"
	"testing"
)

func TestPrepareSpecificationNodesReplacesLegacyAuthorization(t *testing.T) {
	nodes, err := prepareSpecificationNodes([]map[string]any{{
		"id": "node_1", "label": "Review", "doc": map[string]any{"priority": "high"},
		"roles": map[string]any{"backend": map[string]any{"authPolicy": "Bearer old-private-value", "curl": "curl -H 'Authorization: Bearer old-private-value'", "endpoint": "/review"}},
	}})
	if err != nil {
		t.Fatal(err)
	}
	backend := mapField(mapField(nodes[0], "roles"), "backend")
	if backend["auth"] != "{{API_TOKEN}}" || mapField(nodes[0], "legacyNotes")["backend.auth"] == nil {
		t.Fatalf("unsafe legacy auth migration: %#v", nodes)
	}
	if strings.Contains(jsonText(nodes), "old-private-value") {
		t.Fatal("raw secret survived graph preparation")
	}
	if _, exists := backend["curl"]; exists {
		t.Fatal("legacy cURL must not enter the graph")
	}
}

func TestSpecificationRejectsNestedSecretsAndInvalidFields(t *testing.T) {
	cases := []struct {
		name  string
		node  map[string]any
		field string
	}{
		{"required label", map[string]any{"doc": map[string]any{}}, "label"},
		{"risk", map[string]any{"label": "Step", "doc": map[string]any{"riskLevel": "extreme"}}, "riskLevel"},
		{"sla", map[string]any{"label": "Step", "doc": map[string]any{"sla": "about 2 hours"}}, "sla"},
		{"url", map[string]any{"label": "Step", "roles": map[string]any{"uiux": map[string]any{"link": "javascript:alert(1)"}}}, "link"},
		{"facet key", map[string]any{"label": "Step", "roles": map[string]any{"unknown": map[string]any{}}}, "facet"},
		{"required done field", map[string]any{"label": "Step", "roles": map[string]any{"backend": map[string]any{"status": "done"}}}, "serviceCapability"},
		{"endpoint", map[string]any{"label": "Step", "roles": map[string]any{"backend": map[string]any{"endpoint": "//foreign.test/path"}}}, "endpoint"},
		{"nested token", map[string]any{"label": "Step", "roles": map[string]any{"backend": map[string]any{"request": `{"token":"private"}`}}}, "variable reference"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			_, err := prepareSpecificationNodes([]map[string]any{tt.node})
			if err == nil || !strings.Contains(err.Error(), tt.field) {
				t.Fatalf("expected %s error, got %v", tt.field, err)
			}
		})
	}
}

func TestSpecificationCompletenessIsDeterministic(t *testing.T) {
	node := map[string]any{"label": "Step", "doc": map[string]any{"outcome": "Approved"}}
	first := specificationCompleteness(node)
	if first != specificationCompleteness(node) || first <= 0 || first >= 100 {
		t.Fatalf("unexpected incomplete score: %d", first)
	}
	node["doc"] = map[string]any{"outcome": "Approved", "actor": "Agent", "trigger": "Submit", "process": "Review", "rules": []map[string]any{{"description": "Check"}}}
	node["roles"] = map[string]any{"uiux": map[string]any{"userGoal": "Submit", "screen": "Review"}, "frontend": map[string]any{"experienceName": "Review", "validation": "Required"}, "backend": map[string]any{"serviceCapability": "Review", "endpoint": "/reviews"}}
	if got := specificationCompleteness(node); got != 100 {
		t.Fatalf("complete spec scored %d", got)
	}
}

func TestLegacyDisplayRedactsSensitiveExamples(t *testing.T) {
	raw := `{"token":"private","nested":{"authorization":"Bearer private"},"safe":"ok"}`
	got := scrubLegacyDisplay(raw)
	if strings.Contains(got, "private") || !strings.Contains(got, `"safe":"ok"`) {
		t.Fatalf("unsafe display: %s", got)
	}
}

func TestSpecificationEdgesRejectCredentialFields(t *testing.T) {
	_, err := prepareSpecificationEdges([]map[string]any{{"from": "a", "to": "b", "token": "private"}})
	if err == nil || strings.Contains(err.Error(), "private") {
		t.Fatalf("credential must be rejected without echoing its value: %v", err)
	}
}
