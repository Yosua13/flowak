package handlers

import (
	"context"
	"regexp"
	"testing"

	"backend/db"
	"github.com/DATA-DOG/go-sqlmock"
)

func TestReconcileModuleGraphReportsKeyDifferencesWithoutWrites(t *testing.T) {
	database, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	previousDB := db.DB
	db.DB = database
	defer func() { db.DB = previousDB }()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT project_id, nodes, edges FROM modules WHERE id = $1")).
		WithArgs("mod_1").
		WillReturnRows(sqlmock.NewRows([]string{"project_id", "nodes", "edges"}).AddRow("project_1", `[
			{"id":"node_changed","type":"process","label":"Old label","x":10,"y":5},
			{"id":"node_removed","type":"decision","label":"Removed","x":0,"y":0}
		]`, `[
			{"id":"edge_1","from":"node_changed","to":"node_new","label":"Old path","condition":"yes"},
			{"id":"edge_removed","from":"node_removed","to":"node_changed"}
		]`))
	mock.ExpectQuery("SELECT id, type, label, x, y").WithArgs("mod_1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "type", "label", "x", "y"}).
			AddRow("node_changed", "process", "Current label", 10, 5).
			AddRow("node_new", "actor", "New", 20, 10))
	mock.ExpectQuery("SELECT id, from_node_id, to_node_id, label, condition_text").WithArgs("mod_1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "from_node_id", "to_node_id", "label", "condition_text"}).
			AddRow("edge_1", "node_changed", "node_new", "Current path", "yes"))
	mock.ExpectCommit()

	report, err := ReconcileModuleGraph(context.Background(), "mod_1")
	if err != nil {
		t.Fatalf("reconcile module graph: %v", err)
	}
	if report.Matches {
		t.Fatal("expected a mismatch report")
	}
	if report.ProjectID != "project_1" || report.Nodes.SnapshotCount != 2 || report.Nodes.NormalizedCount != 2 {
		t.Fatalf("unexpected node counts or project: %#v", report)
	}
	if got := report.Nodes.MissingFromSnapshot; len(got) != 1 || got[0] != "node_new" {
		t.Fatalf("missing normalized node was not reported: %#v", got)
	}
	if got := report.Nodes.MissingFromNormalized; len(got) != 1 || got[0] != "node_removed" {
		t.Fatalf("orphaned snapshot node was not reported: %#v", got)
	}
	if got := report.Nodes.Changed; len(got) != 1 || got[0].ID != "node_changed" || len(got[0].Fields) != 1 || got[0].Fields[0] != "label" {
		t.Fatalf("node field difference was not reported: %#v", got)
	}
	if got := report.Edges.Changed; len(got) != 1 || got[0].ID != "edge_1" || len(got[0].Fields) != 1 || got[0].Fields[0] != "label" {
		t.Fatalf("edge field difference was not reported: %#v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expected only read-only reconciliation queries: %v", err)
	}
}

func TestReconcileModuleGraphReportsInvalidSnapshotsAndMatchingNormalizedGraph(t *testing.T) {
	database, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	previousDB := db.DB
	db.DB = database
	defer func() { db.DB = previousDB }()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT project_id, nodes, edges FROM modules WHERE id = $1")).
		WithArgs("mod_2").
		WillReturnRows(sqlmock.NewRows([]string{"project_id", "nodes", "edges"}).AddRow("project_1", `invalid`, `[]`))
	mock.ExpectQuery("SELECT id, type, label, x, y").WithArgs("mod_2").
		WillReturnRows(sqlmock.NewRows([]string{"id", "type", "label", "x", "y"}))
	mock.ExpectQuery("SELECT id, from_node_id, to_node_id, label, condition_text").WithArgs("mod_2").
		WillReturnRows(sqlmock.NewRows([]string{"id", "from_node_id", "to_node_id", "label", "condition_text"}))
	mock.ExpectCommit()

	report, err := ReconcileModuleGraph(context.Background(), "mod_2")
	if err != nil {
		t.Fatalf("reconcile module graph: %v", err)
	}
	if report.Matches || len(report.Nodes.SnapshotIssues) != 1 || report.Nodes.SnapshotIssues[0] != "node snapshot is not a JSON array" {
		t.Fatalf("invalid snapshot must produce a deterministic mismatch: %#v", report.Nodes)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expected only read-only reconciliation queries: %v", err)
	}
}

func TestReconcileProjectGraphsReturnsModuleReportsInIDOrder(t *testing.T) {
	database, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	previousDB := db.DB
	db.DB = database
	defer func() { db.DB = previousDB }()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM modules WHERE project_id = $1 ORDER BY id")).
		WithArgs("project_1").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("mod_a").AddRow("mod_b"))
	for _, moduleID := range []string{"mod_a", "mod_b"} {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT project_id, nodes, edges FROM modules WHERE id = $1")).
			WithArgs(moduleID).WillReturnRows(sqlmock.NewRows([]string{"project_id", "nodes", "edges"}).AddRow("project_1", `[]`, `[]`))
		mock.ExpectQuery("SELECT id, type, label, x, y").WithArgs(moduleID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "type", "label", "x", "y"}))
		mock.ExpectQuery("SELECT id, from_node_id, to_node_id, label, condition_text").WithArgs(moduleID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "from_node_id", "to_node_id", "label", "condition_text"}))
	}
	mock.ExpectCommit()

	reports, err := ReconcileProjectGraphs(context.Background(), "project_1")
	if err != nil {
		t.Fatalf("reconcile project graphs: %v", err)
	}
	if len(reports) != 2 || reports[0].ModuleID != "mod_a" || reports[1].ModuleID != "mod_b" || !reports[0].Matches || !reports[1].Matches {
		t.Fatalf("unexpected project reports: %#v", reports)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expected only read-only reconciliation queries: %v", err)
	}
}
