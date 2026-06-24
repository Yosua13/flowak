package handlers

import "testing"

func TestParseSLA(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantValue float64
		wantUnit  string
		wantNil   bool
	}{
		{name: "hours indonesian", input: "24 jam", wantValue: 24, wantUnit: "jam"},
		{name: "days english", input: "3 days", wantValue: 3, wantUnit: "hari"},
		{name: "decimal comma", input: "1,5 jam", wantValue: 1.5, wantUnit: "jam"},
		{name: "empty", input: "Instan", wantNil: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, unit := parseSLA(tt.input)
			if tt.wantNil {
				if value != nil || unit != nil {
					t.Fatalf("expected nil SLA, got value=%v unit=%v", value, unit)
				}
				return
			}
			if value != tt.wantValue {
				t.Fatalf("value mismatch: got %v want %v", value, tt.wantValue)
			}
			if unit != tt.wantUnit {
				t.Fatalf("unit mismatch: got %v want %v", unit, tt.wantUnit)
			}
		})
	}
}

func TestJSONBFromTextAndDisplay(t *testing.T) {
	valid := jsonbFromText(`{"success": true}`)
	if valid != `{"success":true}` {
		t.Fatalf("unexpected compact JSON: %s", valid)
	}

	raw := jsonbFromText("not-json")
	display := jsonbDisplay(raw)
	if display != "not-json" {
		t.Fatalf("expected raw text round-trip, got %q from %s", display, raw)
	}
}

func TestRoleTaskIDIsStableAndBounded(t *testing.T) {
	id := roleTaskID("node_abc-123", "frontend")
	if id != "task_node_abc_123_frontend" {
		t.Fatalf("unexpected task id: %s", id)
	}

	longID := roleTaskID("node_"+string(make([]byte, 300)), "backend")
	if len(longID) > 150 {
		t.Fatalf("task id too long: %d", len(longID))
	}
}
