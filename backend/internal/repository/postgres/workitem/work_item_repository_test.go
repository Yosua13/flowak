package workitem_test

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"backend/internal/domain/workitem"
	repo "backend/internal/repository/postgres/workitem"
	"github.com/DATA-DOG/go-sqlmock"
)

var testColumns = []string{
	"id", "work_key", "project_id", "module_id", "node_id", "facet_key", "parent_id",
	"type", "title", "description", "priority", "points", "status", "assignee_id",
	"reporter_id", "start_date", "due_date", "blocked_reason", "resolution",
	"row_version", "created_at", "updated_at",
}

func TestPostgresWorkItemRepository_FindByKey(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	r := repo.NewPostgresWorkItemRepository(db)
	now := time.Now()

	// 1. Success case
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, work_key, project_id, module_id, node_id, facet_key, parent_id, type, title, description, priority, points, status, assignee_id, reporter_id, start_date, due_date, blocked_reason, resolution, row_version, created_at, updated_at FROM work_items WHERE work_key=$1 AND deleted_at IS NULL")).
		WithArgs("PROJ-1").
		WillReturnRows(sqlmock.NewRows(testColumns).AddRow(
			"wi_1", "PROJ-1", "proj_1", nil, nil, nil, nil,
			"Task", "Test Task", "Desc", "medium", 3, "Ready", nil,
			"usr_1", nil, nil, nil, nil,
			1, now, now,
		))

	item, err := r.FindByKey(context.Background(), "PROJ-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Key != "PROJ-1" || item.Title != "Test Task" || item.Status != workitem.StatusReady {
		t.Fatalf("mismatched work item: %+v", item)
	}

	// 2. Not found case
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, work_key, project_id, module_id, node_id, facet_key, parent_id, type, title, description, priority, points, status, assignee_id, reporter_id, start_date, due_date, blocked_reason, resolution, row_version, created_at, updated_at FROM work_items WHERE work_key=$1 AND deleted_at IS NULL")).
		WithArgs("NONEXISTENT").
		WillReturnError(sql.ErrNoRows)

	_, err = r.FindByKey(context.Background(), "NONEXISTENT")
	if !errors.Is(err, workitem.ErrWorkItemNotFound) {
		t.Fatalf("expected ErrWorkItemNotFound, got: %v", err)
	}
}

func TestPostgresWorkItemRepository_UpdateStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	r := repo.NewPostgresWorkItemRepository(db)
	now := time.Now()

	// 1. Optimistic Lock Conflict (0 rows affected)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT status, project_id FROM work_items WHERE id=$1 AND deleted_at IS NULL")).
		WithArgs("wi_1").
		WillReturnRows(sqlmock.NewRows([]string{"status", "project_id"}).AddRow("Ready", "proj_1"))
	mock.ExpectExec("UPDATE work_items SET status=").
		WithArgs("In Progress", nil, nil, "wi_1", 1).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	_, err = r.UpdateStatus(context.Background(), "wi_1", workitem.StatusInProgress, 1, "", "", "usr_1")
	if !errors.Is(err, workitem.ErrOptimisticLockConflict) {
		t.Fatalf("expected ErrOptimisticLockConflict, got: %v", err)
	}

	// 2. Successful Status Update
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT status, project_id FROM work_items WHERE id=$1 AND deleted_at IS NULL")).
		WithArgs("wi_1").
		WillReturnRows(sqlmock.NewRows([]string{"status", "project_id"}).AddRow("Ready", "proj_1"))
	mock.ExpectExec("UPDATE work_items SET status=").
		WithArgs("In Progress", nil, nil, "wi_1", 1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO work_item_status_history").
		WithArgs(sqlmock.AnyArg(), "wi_1", "Ready", "In Progress", "usr_1", nil).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT user_id FROM work_item_watchers WHERE work_item_id=$1")).
		WithArgs("wi_1").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}))
	mock.ExpectCommit()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, work_key, project_id, module_id, node_id, facet_key, parent_id, type, title, description, priority, points, status, assignee_id, reporter_id, start_date, due_date, blocked_reason, resolution, row_version, created_at, updated_at FROM work_items WHERE id=$1 AND deleted_at IS NULL")).
		WithArgs("wi_1").
		WillReturnRows(sqlmock.NewRows(testColumns).AddRow(
			"wi_1", "PROJ-1", "proj_1", nil, nil, nil, nil,
			"Task", "Test Task", "Desc", "medium", 3, "In Progress", nil,
			"usr_1", nil, nil, nil, nil,
			2, now, now,
		))

	updated, err := r.UpdateStatus(context.Background(), "wi_1", workitem.StatusInProgress, 1, "", "", "usr_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Status != workitem.StatusInProgress || updated.RowVersion != 2 {
		t.Fatalf("unexpected updated entity: %+v", updated)
	}
}
