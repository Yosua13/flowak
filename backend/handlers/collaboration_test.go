package handlers

import (
	"encoding/json"
	"testing"
)

func TestSafeEventPayloadRemovesSensitiveFields(t *testing.T) {
	payload := map[string]any{"module_id": "mod_1", "token": "secret", "response": map[string]any{"body": "private"}}
	var decoded map[string]any
	if err := json.Unmarshal(safeEventPayload(payload), &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["module_id"] != "mod_1" {
		t.Fatal("public event context was lost")
	}
	if _, ok := decoded["token"]; ok {
		t.Fatal("token must not reach activity or event payload")
	}
	if _, ok := decoded["response"]; ok {
		t.Fatal("response must not reach activity or event payload")
	}
}
