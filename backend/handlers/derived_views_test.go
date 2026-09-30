package handlers

import (
	"database/sql/driver"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"backend/models"
	"github.com/DATA-DOG/go-sqlmock"
)

func TestDerivedViewFilterContract(t *testing.T) {
	for _, facet := range []string{"", "business", "uiux", "frontend", "backend"} {
		if !validDerivedFacet(facet) {
			t.Fatalf("expected %q to be valid", facet)
		}
	}
	if validDerivedFacet("unknown") {
		t.Fatal("unknown facet must be rejected")
	}
	marks, args := placeholders([]string{"wi_1", "wi_2"}, 3)
	if marks != "$3,$4" || len(args) != 2 {
		t.Fatalf("unexpected placeholders: %q %#v", marks, args)
	}
}

var derivedViewWorkItemColumns = []string{"id", "work_key", "project_id", "module_id", "node_id", "facet_key", "parent_id", "type", "title", "description", "priority", "points", "status", "assignee_id", "reporter_id", "start_date", "due_date", "blocked_reason", "resolution", "row_version", "created_at", "updated_at"}

func derivedViewWorkItemRow() []driver.Value {
	created := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	due := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	return []driver.Value{"work-1", "FLOW-1", "project-1", "module-1", "node-1", "backend", nil, "Task", "Ship payment", "Delivery scope", "high", 3, "Done", nil, tenantTestUser, start, due, nil, "released", 1, created, created}
}

func expectDerivedViewAccess(mock sqlmock.Sqlmock) {
	mock.ExpectQuery("SELECT COALESCE").WithArgs(tenantTestUser, "project-1", tenantTestOrg).
		WillReturnRows(sqlmock.NewRows([]string{"project_role"}).AddRow("viewer"))
}

func TestDerivedViewFixtureContractAppliesProjectModuleAndFacetFilters(t *testing.T) {
	mock := isolatedTenantDB(t)
	expectDerivedViewAccess(mock)
	mock.ExpectQuery("SELECT EXISTS").WithArgs("module-1", "project-1").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery("SELECT id, work_key").WithArgs("project-1", "module-1", "backend").
		WillReturnRows(sqlmock.NewRows(derivedViewWorkItemColumns).AddRow(derivedViewWorkItemRow()...))
	mock.ExpectQuery("SELECT work_item_id,from_status,to_status,note,created_at").WithArgs("work-1").
		WillReturnRows(sqlmock.NewRows([]string{"work_item_id", "from_status", "to_status", "note", "created_at"}).
			AddRow("work-1", nil, "Backlog", nil, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)).
			AddRow("work-1", "Backlog", "Blocked", "provider unavailable", time.Date(2026, 1, 1, 2, 0, 0, 0, time.UTC)).
			AddRow("work-1", "Blocked", "In Progress", "provider restored", time.Date(2026, 1, 1, 5, 0, 0, 0, time.UTC)).
			AddRow("work-1", "In Progress", "Done", "released", time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)))
	mock.ExpectQuery("SELECT id,work_item_id,body,resolved_at,created_at").WithArgs("work-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "work_item_id", "body", "resolved_at", "created_at"}).AddRow("comment-1", "work-1", "Decision recorded", nil, time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC)))
	mock.ExpectQuery("SELECT id,file_name,work_item_id,node_id,created_at").WithArgs("work-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "file_name", "work_item_id", "node_id", "created_at"}).AddRow("evidence-1", "payment-proof.pdf", "work-1", "node-1", time.Date(2026, 1, 1, 11, 0, 0, 0, time.UTC)))
	mock.ExpectQuery("SELECT n.id,n.label,r.role_key,r.readiness,r.status,r.due_date").WithArgs("module-1", "backend").
		WillReturnRows(sqlmock.NewRows([]string{"id", "label", "role_key", "readiness", "status", "due_date"}).AddRow("node-1", "Charge payment", "backend", "review", "review", time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC)))
	mock.ExpectQuery("SELECT module_id,version,created_at").WithArgs("module-1").
		WillReturnRows(sqlmock.NewRows([]string{"module_id", "version", "created_at"}).AddRow("module-1", 4, time.Date(2026, 1, 4, 0, 0, 0, 0, time.UTC)))

	router := tenantTestRouter(http.MethodGet, "/projects/:id/derived-view-data", GetDerivedViewDataHandler, "viewer")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/projects/project-1/derived-view-data?module_id=module-1&facet_key=backend", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", response.Code, response.Body.String())
	}
	var payload DerivedViewData
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.WorkItems) != 1 || payload.WorkItems[0].Key != "FLOW-1" || len(payload.History) != 4 || len(payload.FacetReviews) != 1 || len(payload.Baselines) != 1 || len(payload.Comments) != 1 || len(payload.Evidence) != 1 {
		t.Fatalf("unexpected derived fixture payload: %#v", payload)
	}
	if strings.Contains(response.Body.String(), "response-secret") || strings.Contains(response.Body.String(), "api_key") {
		t.Fatalf("derived evidence must not expose API response or environment data: %s", response.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestDerivedViewFixtureContractReturnsEmptyCollections(t *testing.T) {
	mock := isolatedTenantDB(t)
	expectDerivedViewAccess(mock)
	mock.ExpectQuery("SELECT id, work_key").WithArgs("project-1").
		WillReturnRows(sqlmock.NewRows(derivedViewWorkItemColumns))
	router := tenantTestRouter(http.MethodGet, "/projects/:id/derived-view-data", GetDerivedViewDataHandler, "viewer")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/projects/project-1/derived-view-data", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", response.Code, response.Body.String())
	}
	var payload DerivedViewData
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.WorkItems) != 0 || len(payload.History) != 0 || len(payload.FacetReviews) != 0 || len(payload.Baselines) != 0 || len(payload.Comments) != 0 || len(payload.Evidence) != 0 {
		t.Fatalf("empty project should retain empty collections: %#v", payload)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestDerivedViewFixturePaginationKeepsFilteredRecordsDistinct(t *testing.T) {
	mock := isolatedTenantDB(t)
	fields := append([]string{"sequence"}, derivedViewWorkItemColumns...)
	row := func(sequence int64, key string) []driver.Value {
		values := append([]driver.Value{sequence}, derivedViewWorkItemRow()...)
		values[2] = key
		return values
	}
	page := func(cursor string, rows *sqlmock.Rows) (models.WorkItem, string) {
		expectDerivedViewAccess(mock)
		if cursor == "" {
			mock.ExpectQuery("SELECT COALESCE\\(MAX\\(sequence\\),0\\)").WithArgs("project-1").WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(2))
		}
		mock.ExpectQuery("SELECT sequence,").WillReturnRows(rows)
		path := "/projects/project-1/work-items?page=1&limit=1&status=Done&node_id=node-1"
		if cursor != "" {
			path += "&cursor=" + cursor
		}
		router := tenantTestRouter(http.MethodGet, "/projects/:id/work-items", ListWorkItemsHandler, "viewer")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200: %s", response.Code, response.Body.String())
		}
		var payload struct {
			Items      []models.WorkItem `json:"items"`
			NextCursor string            `json:"next_cursor"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil || len(payload.Items) != 1 {
			t.Fatalf("invalid page: %v %s", err, response.Body.String())
		}
		return payload.Items[0], payload.NextCursor
	}
	first, next := page("", sqlmock.NewRows(fields).AddRow(row(2, "FLOW-2")...).AddRow(row(1, "FLOW-1")...))
	second, end := page(next, sqlmock.NewRows(fields).AddRow(row(1, "FLOW-1")...))
	if first.Key != "FLOW-2" || second.Key != "FLOW-1" || first.Key == second.Key || next == "" || end != "" {
		t.Fatalf("unexpected cursor pages: first=%s second=%s next=%q end=%q", first.Key, second.Key, next, end)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
