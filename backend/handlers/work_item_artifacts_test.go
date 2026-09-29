package handlers

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"backend/models"
	"github.com/DATA-DOG/go-sqlmock"
)

func expectArtifactAuth(mock sqlmock.Sqlmock, role string) {
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id,project_id FROM work_items WHERE work_key=$1 AND deleted_at IS NULL")).WithArgs("FLOW-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id"}).AddRow("item-1", "project-1"))
	mock.ExpectQuery("SELECT COALESCE").WithArgs(tenantTestUser, "project-1", tenantTestOrg).
		WillReturnRows(sqlmock.NewRows([]string{"project_role"}).AddRow(role))
}

func TestArtifactReadRejectsCrossProject(t *testing.T) {
	mock := isolatedTenantDB(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id,project_id FROM work_items WHERE work_key=$1 AND deleted_at IS NULL")).WithArgs("FLOW-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id"}).AddRow("item-1", "foreign-project"))
	expectProjectDenied(mock, "foreign-project")
	router := tenantTestRouter(http.MethodGet, "/work-items/:key/artifacts", GetWorkItemArtifactsHandler, "viewer")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/work-items/FLOW-1/artifacts", nil))
	assertAPIError(t, response, http.StatusForbidden, "Project access denied")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestViewerCannotCreateChecklist(t *testing.T) {
	mock := isolatedTenantDB(t)
	expectArtifactAuth(mock, "viewer")
	router := tenantTestRouter(http.MethodPost, "/work-items/:key/artifacts/:kind", MutateWorkItemArtifactHandler, "viewer")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/work-items/FLOW-1/artifacts/checklist", bytes.NewBufferString(`{"body":"Verify delivery"}`)))
	assertAPIError(t, response, http.StatusForbidden, "Project access denied")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestChecklistCreateAuditedInSameTransaction(t *testing.T) {
	mock := isolatedTenantDB(t)
	expectArtifactAuth(mock, "editor")
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO work_item_acceptance_criteria").WithArgs(sqlmock.AnyArg(), "item-1", "Verify delivery", 0).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO activity_logs").WithArgs(sqlmock.AnyArg(), "project-1", tenantTestUser, "artifact_created", "item-1", sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	router := tenantTestRouter(http.MethodPost, "/work-items/:key/artifacts/:kind", MutateWorkItemArtifactHandler, "editor")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/work-items/FLOW-1/artifacts/checklist", bytes.NewBufferString(`{"body":"Verify delivery"}`)))
	if response.Code != http.StatusCreated {
		t.Fatalf("status %d: %s", response.Code, response.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAttachmentMetadataRejectsCredentialURL(t *testing.T) {
	mock := isolatedTenantDB(t)
	expectArtifactAuth(mock, "editor")
	router := tenantTestRouter(http.MethodPost, "/work-items/:key/artifacts/:kind", MutateWorkItemArtifactHandler, "editor")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/work-items/FLOW-1/artifacts/attachments", bytes.NewBufferString(`{"file_name":"proof.pdf","storage_key":"https://files.example.test/proof.pdf?token=secret"}`)))
	assertAPIError(t, response, http.StatusBadRequest, "storage_key must be an HTTPS artifact URL without credentials or query")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestLinkCannotTargetAnotherProject(t *testing.T) {
	mock := isolatedTenantDB(t)
	expectArtifactAuth(mock, "editor")
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM work_items WHERE work_key=").WithArgs("OTHER-1", "project-1").WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()
	router := tenantTestRouter(http.MethodPost, "/work-items/:key/artifacts/:kind", MutateWorkItemArtifactHandler, "editor")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/work-items/FLOW-1/artifacts/links", bytes.NewBufferString(`{"linked_key":"OTHER-1"}`)))
	assertAPIError(t, response, http.StatusBadRequest, "linked item must be distinct and in the same project")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestArtifactDeleteIsScopedToWorkItemAndAudited(t *testing.T) {
	mock := isolatedTenantDB(t)
	expectArtifactAuth(mock, "editor")
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM attachments WHERE id=").WithArgs("attachment-1", "item-1", "project-1").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO activity_logs").WithArgs(sqlmock.AnyArg(), "project-1", tenantTestUser, "artifact_deleted", "item-1", sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	router := tenantTestRouter(http.MethodDelete, "/work-items/:key/artifacts/:kind/:artifactId", MutateWorkItemArtifactHandler, "editor")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodDelete, "/work-items/FLOW-1/artifacts/attachments/attachment-1", nil))
	if response.Code != http.StatusNoContent {
		t.Fatalf("status %d: %s", response.Code, response.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestWorkItemDeliveryChangesOnlyIncludesChangedFields(t *testing.T) {
	points := 3
	due := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	before := models.WorkItem{Points: &points, DueDate: &due}
	newDue := "2026-09-02"
	after := models.WorkItemRequest{Points: &points, DueDate: &newDue}
	oldValues, newValues := workItemDeliveryChanges(before, after)
	if len(newValues) != 1 || *oldValues["due_date"].(*string) != "2026-09-01" || *newValues["due_date"].(*string) != "2026-09-02" {
		t.Fatalf("unexpected activity diff: %#v %#v", oldValues, newValues)
	}
}

func testWorkItemRow(sequence int64) []driver.Value {
	return []driver.Value{sequence, "wi-" + string(rune('0'+sequence)), "FLOW-" + string(rune('0'+sequence)), "project-1", nil, nil, nil, nil, "Task", "Task", nil, "medium", nil, "Backlog", nil, tenantTestUser, nil, nil, nil, nil, 1, time.Now(), time.Now()}
}

func TestListWorkItemsCursorKeepsSequenceOrderWithoutDuplicates(t *testing.T) {
	mock := isolatedTenantDB(t)
	fields := []string{"sequence", "id", "work_key", "project_id", "module_id", "node_id", "facet_key", "parent_id", "type", "title", "description", "priority", "points", "status", "assignee_id", "reporter_id", "start_date", "due_date", "blocked_reason", "resolution", "row_version", "created_at", "updated_at"}
	page := func(cursor string, rows *sqlmock.Rows) (int, string, string) {
		mock.ExpectQuery("SELECT COALESCE").WithArgs(tenantTestUser, "project-1", tenantTestOrg).WillReturnRows(sqlmock.NewRows([]string{"project_role"}).AddRow("viewer"))
		if cursor == "" {
			mock.ExpectQuery("SELECT COALESCE\\(MAX\\(sequence\\),0\\)").WithArgs("project-1").WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(3))
		}
		mock.ExpectQuery("SELECT sequence,").WillReturnRows(rows)
		path := "/projects/project-1/work-items?page=1&limit=1&sort=newest"
		if cursor != "" {
			path += "&cursor=" + cursor
		}
		router := tenantTestRouter(http.MethodGet, "/projects/:id/work-items", ListWorkItemsHandler, "viewer")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		var payload struct {
			Items      []models.WorkItem `json:"items"`
			NextCursor string            `json:"next_cursor"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
			t.Fatal(err)
		}
		if len(payload.Items) != 1 {
			t.Fatalf("unexpected page: %s", response.Body.String())
		}
		return response.Code, payload.Items[0].Key, payload.NextCursor
	}
	status, first, next := page("", sqlmock.NewRows(fields).AddRow(testWorkItemRow(3)...).AddRow(testWorkItemRow(2)...))
	if status != 200 || first != "FLOW-3" || next == "" {
		t.Fatalf("first page: %d %s %s", status, first, next)
	}
	status, second, finalCursor := page(next, sqlmock.NewRows(fields).AddRow(testWorkItemRow(2)...).AddRow(testWorkItemRow(1)...))
	if status != 200 || second != "FLOW-2" || second == first {
		t.Fatalf("second page: %d %s", status, second)
	}
	status, third, end := page(finalCursor, sqlmock.NewRows(fields).AddRow(testWorkItemRow(1)...))
	if status != 200 || third != "FLOW-1" || end != "" {
		t.Fatalf("third page: %d %s %s", status, third, end)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestListWorkItemsRejectsCursorFromDifferentSort(t *testing.T) {
	mock := isolatedTenantDB(t)
	mock.ExpectQuery("SELECT COALESCE").WithArgs(tenantTestUser, "project-1", tenantTestOrg).
		WillReturnRows(sqlmock.NewRows([]string{"project_role"}).AddRow("viewer"))
	cursor := encodeWorkItemCursor(workItemCursor{Project: "project-1", Sort: "newest", Filter: workItemFilterKey("", "", ""), Sequence: 2, Ceiling: 3})
	router := tenantTestRouter(http.MethodGet, "/projects/:id/work-items", ListWorkItemsHandler, "viewer")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/projects/project-1/work-items?page=1&sort=oldest&cursor="+cursor, nil))
	assertAPIError(t, response, http.StatusBadRequest, "invalid cursor for sort or filters")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
