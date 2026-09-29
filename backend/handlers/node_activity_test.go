package handlers

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestListNodeActivityIsProjectScoped(t *testing.T) {
	mock := isolatedTenantDB(t)
	mock.ExpectQuery("SELECT m.project_id FROM workflow_nodes").WithArgs("node_b").
		WillReturnRows(sqlmock.NewRows([]string{"project_id"}).AddRow("project_b"))
	expectProjectDenied(mock, "project_b")

	router := tenantTestRouter(http.MethodGet, "/nodes/:id/activity", ListNodeActivityHandler, "viewer")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/nodes/node_b/activity", nil))
	assertAPIError(t, response, http.StatusForbidden, "Project access denied")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestListNodeActivityReturnsOnlyNodeEntries(t *testing.T) {
	mock := isolatedTenantDB(t)
	mock.ExpectQuery("SELECT m.project_id FROM workflow_nodes").WithArgs("node_a").
		WillReturnRows(sqlmock.NewRows([]string{"project_id"}).AddRow("project_a"))
	mock.ExpectQuery("SELECT COALESCE").WithArgs(tenantTestUser, "project_a", tenantTestOrg).
		WillReturnRows(sqlmock.NewRows([]string{"project_role"}).AddRow("viewer"))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, action, actor_id, created_at FROM (")).WithArgs("node_a", "project_a").
		WillReturnRows(sqlmock.NewRows([]string{"id", "action", "actor_id", "created_at"}).
			AddRow("comment:comment_a", "comment_created", "user_a", time.Date(2026, 9, 25, 10, 0, 0, 0, time.UTC)).
			AddRow("activity_a", "delivery_fields_changed", nil, time.Date(2026, 9, 25, 9, 0, 0, 0, time.UTC)))

	router := tenantTestRouter(http.MethodGet, "/nodes/:id/activity", ListNodeActivityHandler, "viewer")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/nodes/node_a/activity", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", response.Code, response.Body.String())
	}
	if !regexp.MustCompile(`"action":"comment_created"`).Match(response.Body.Bytes()) || !regexp.MustCompile(`"action":"delivery_fields_changed"`).Match(response.Body.Bytes()) {
		t.Fatalf("activity response does not contain expected entries: %s", response.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
