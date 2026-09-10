package handlers

import (
	"errors"
	"regexp"
	"testing"

	"backend/models"
	"github.com/DATA-DOG/go-sqlmock"
)

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

func TestUpsertNodeIsIncrementalAndIdempotent(t *testing.T) {
	database, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	mock.ExpectBegin()
	tx, err := database.Begin()
	if err != nil {
		t.Fatal(err)
	}

	node := map[string]any{"id": "node_1", "type": "process", "label": "Review", "x": 12, "y": 8, "rowVersion": 3}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT row_version FROM workflow_nodes WHERE id = $1 AND module_id = $2 AND deleted_at IS NULL")).
		WithArgs("node_1", "mod_1").WillReturnRows(sqlmock.NewRows([]string{"row_version"}).AddRow(3))
	// A zero affected update means the normalized values were unchanged. No DELETE is
	// expected, which is what preserves comments and other node children.
	mock.ExpectExec("UPDATE workflow_nodes SET").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	if err := upsertNode(tx, "mod_1", "node_1", 0, node, map[string]any{}, nil, nil, 3, true); err != nil {
		t.Fatalf("idempotent upsert failed: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUpsertNodeRejectsStaleVersion(t *testing.T) {
	database, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	mock.ExpectBegin()
	tx, err := database.Begin()
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT row_version FROM workflow_nodes WHERE id = $1 AND module_id = $2 AND deleted_at IS NULL")).
		WithArgs("node_1", "mod_1").WillReturnRows(sqlmock.NewRows([]string{"row_version"}).AddRow(2))
	mock.ExpectRollback()

	err = upsertNode(tx, "mod_1", "node_1", 0, map[string]any{"id": "node_1"}, map[string]any{}, nil, nil, 1, true)
	var graphErr *graphSyncError
	if !errors.As(err, &graphErr) || graphErr.Code != graphConflictCode {
		t.Fatalf("expected stable conflict, got %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSyncModuleGraphRejectsInvalidEdge(t *testing.T) {
	database, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	mock.ExpectBegin()
	tx, err := database.Begin()
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM workflow_nodes WHERE module_id = $1 AND deleted_at IS NULL")).
		WithArgs("mod_1").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("node_1"))
	mock.ExpectRollback()

	err = syncModuleGraph(tx, "mod_1", []map[string]any{}, []map[string]any{{"id": "edge_1", "from": "node_1", "to": "other_module_node"}}, nil, nil)
	var graphErr *graphSyncError
	if !errors.As(err, &graphErr) || graphErr.Code != graphInvalidCode {
		t.Fatalf("expected invalid edge error, got %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestTombstoneNodeRequiresExplicitSafeDelete(t *testing.T) {
	database, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	mock.ExpectBegin()
	tx, err := database.Begin()
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery("SELECT EXISTS").WithArgs("mod_1", "node_1").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectExec("UPDATE workflow_nodes SET deleted_at").WithArgs("node_1", "mod_1", 4).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := tombstoneNodes(tx, "mod_1", []models.GraphDelete{{ID: "node_1", RowVersion: 4}}); err != nil {
		t.Fatalf("tombstone failed: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRowVersion(t *testing.T) {
	if version, ok := rowVersion(map[string]any{"rowVersion": 2.0}); !ok || version != 2 {
		t.Fatalf("expected row version 2, got %d, %t", version, ok)
	}
	if _, ok := rowVersion(map[string]any{}); ok {
		t.Fatal("missing row version must not be accepted for an existing record")
	}
}
