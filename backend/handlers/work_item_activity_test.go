package handlers

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestWorkItemActivityRequiresProjectMembership(t *testing.T) {
	mock := isolatedTenantDB(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id,project_id FROM work_items WHERE work_key=$1 AND deleted_at IS NULL")).WithArgs("FLOW-2").
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id"}).AddRow("item-2", "project-2"))
	expectProjectDenied(mock, "project-2")
	router := tenantTestRouter(http.MethodGet, "/work-items/:key/activity", GetWorkItemActivityHandler, "viewer")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/work-items/FLOW-2/activity", nil))
	assertAPIError(t, response, http.StatusForbidden, "Project access denied")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestWorkItemActivityReturnsSafeTimeline(t *testing.T) {
	mock := isolatedTenantDB(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id,project_id FROM work_items WHERE work_key=$1 AND deleted_at IS NULL")).WithArgs("FLOW-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id"}).AddRow("item-1", "project-1"))
	mock.ExpectQuery("SELECT COALESCE").WithArgs(tenantTestUser, "project-1", tenantTestOrg).WillReturnRows(sqlmock.NewRows([]string{"project_role"}).AddRow("viewer"))
	mock.ExpectQuery("SELECT id,action,actor_id,created_at FROM activity_logs").WithArgs("project-1", "item-1", nil, nil, 51).
		WillReturnRows(sqlmock.NewRows([]string{"id", "action", "actor_id", "created_at"}).AddRow("act-1", "updated", "user-1", time.Date(2026, 9, 28, 10, 0, 0, 0, time.UTC)))
	router := tenantTestRouter(http.MethodGet, "/work-items/:key/activity", GetWorkItemActivityHandler, "viewer")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/work-items/FLOW-1/activity", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", response.Code, response.Body.String())
	}
	if !regexp.MustCompile(`"action":"updated"`).Match(response.Body.Bytes()) {
		t.Fatalf("missing activity: %s", response.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
