package db

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkItemTransitionHistoryCompatibilityMigrationIsAdditive(t *testing.T) {
	path := filepath.Join("migrations", "12_work_item_transition_history_compatibility.sql")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read compatibility migration: %v", err)
	}
	statement := strings.ToUpper(string(content))
	for _, fragment := range []string{"ALTER TABLE WORK_ITEM_STATUS_HISTORY", "ADD COLUMN IF NOT EXISTS FROM_STATUS"} {
		if !strings.Contains(statement, fragment) {
			t.Fatalf("migration must contain %q", fragment)
		}
	}
}
