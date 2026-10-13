package module_test

import (
	"context"
	"errors"
	"regexp"
	"testing"

	domainModule "backend/internal/domain/module"
	pgModule "backend/internal/repository/postgres/module"
	"github.com/DATA-DOG/go-sqlmock"
)

func TestFindGraphByModuleID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()
	moduleID := "mod_1"

	// Mock modules query
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, project_id, name, description, schema_version, version FROM modules WHERE id = $1 AND status = 'active'")).
		WithArgs(moduleID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "name", "description", "schema_version", "version"}).
			AddRow("mod_1", "proj_1", "Checkout Flow", "Desc", 1, 2))

	// Mock workflow_nodes query
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, type, label, x, y, actor, trigger, input_desc, process_desc, output_desc, business_rules, exception_path, system_context, sla_value, sla_unit, priority, risk_level, acceptance_criteria, outcome, trigger_type, preconditions, reference_links, metadata, row_version FROM workflow_nodes WHERE module_id = $1 AND deleted_at IS NULL")).
		WithArgs(moduleID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "type", "label", "x", "y", "actor", "trigger", "input_desc", "process_desc",
			"output_desc", "business_rules", "exception_path", "system_context", "sla_value",
			"sla_unit", "priority", "risk_level", "acceptance_criteria", "outcome", "trigger_type",
			"preconditions", "reference_links", "metadata", "row_version",
		}).AddRow("node_1", "process", "Start", 10.0, 20.0, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, []byte("{}"), 1))

	// Mock business rules & decision outcomes queries
	mock.ExpectQuery(regexp.QuoteMeta("SELECT r.node_id, COALESCE(r.rule_code, ''), r.severity, r.description FROM node_business_rules r")).
		WithArgs(moduleID).
		WillReturnRows(sqlmock.NewRows([]string{"node_id", "rule_code", "severity", "description"}))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT o.node_id, o.outcome, COALESCE(o.edge_id, '') FROM node_decision_outcomes o")).
		WithArgs(moduleID).
		WillReturnRows(sqlmock.NewRows([]string{"node_id", "outcome", "edge_id"}))

	// Mock role tasks and specs queries
	mock.ExpectQuery(regexp.QuoteMeta("SELECT t.node_id, t.role_key, COALESCE(u.name, t.assignee_label, '') AS assignee")).
		WithArgs(moduleID).
		WillReturnRows(sqlmock.NewRows([]string{"node_id", "role_key", "assignee", "status", "due_date", "notes"}))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT s.node_id, COALESCE(s.screen_name, '')")).
		WithArgs(moduleID).
		WillReturnRows(sqlmock.NewRows([]string{"node_id", "screen_name", "prototype_url", "wireframe_url", "state_notes", "accessibility_notes", "user_goal", "surface", "figma_frame_url", "design_version", "screen_states", "interactions", "content_messages", "responsive_intent"}))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT s.node_id, COALESCE(s.page_name, '')")).
		WithArgs(moduleID).
		WillReturnRows(sqlmock.NewRows([]string{"node_id", "page_name", "route_path", "interaction_notes", "validation_notes", "state_handling", "handoff_url", "experience_name", "entry_exit_behavior", "input_requirements", "api_references", "analytics_intent", "feature_availability"}))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT c.node_id, COALESCE(c.method, '')")).
		WithArgs(moduleID).
		WillReturnRows(sqlmock.NewRows([]string{"node_id", "method", "endpoint_path", "auth_policy", "request_example", "response_example", "status_code", "error_codes", "service_capability", "api_references", "business_validation", "dependency_references", "idempotency_notes", "caching_notes", "security_notes", "observability_intent", "sla_value", "sla_unit"}))

	// Mock workflow_edges query
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, from_node_id, to_node_id, label, condition_text, row_version FROM workflow_edges WHERE module_id = $1 AND deleted_at IS NULL")).
		WithArgs(moduleID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "from_node_id", "to_node_id", "label", "condition_text", "row_version"}).
			AddRow("edge_1", "node_1", "node_2", "next", "", 1))

	repo := pgModule.NewModulePostgresRepo(db)
	g, err := repo.FindGraphByModuleID(ctx, moduleID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if g.ID != "mod_1" || len(g.Nodes) != 1 || len(g.Edges) != 1 {
		t.Fatalf("unexpected graph result: %+v", g)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteNodes_RejectsConnectedEdges(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()
	mock.ExpectBegin()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectQuery("SELECT EXISTS").WithArgs("mod_1", "node_1").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectRollback()

	repo := pgModule.NewModulePostgresRepo(db)
	err = repo.DeleteNodes(ctx, tx, "mod_1", []domainModule.GraphDelete{{ID: "node_1", RowVersion: 1}})
	if !errors.Is(err, domainModule.ErrInvalidGraphReference) {
		t.Fatalf("expected ErrInvalidGraphReference for node with edges, got: %v", err)
	}
	_ = tx.Rollback()
}
