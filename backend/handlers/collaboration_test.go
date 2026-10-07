package handlers

import (
	"backend/db"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
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

func TestReplayProjectEventsUsesTenantScopedLastEventID(t *testing.T) {
	mock := isolatedTenantDB(t)
	created := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	mock.ExpectQuery("SELECT created_at FROM event_outbox").WithArgs("project-1", "evt-1").
		WillReturnRows(sqlmock.NewRows([]string{"created_at"}).AddRow(created))
	mock.ExpectQuery("SELECT id,COALESCE\\(module_id,''\\),event_name,payload FROM event_outbox").WithArgs("project-1", created, "evt-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "module_id", "event_name", "payload"}).AddRow("evt-2", "module-1", "work_item.updated", []byte(`{"work_item_id":"item-1"}`)))
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	if err := replayProjectEvents(context, "project-1", "evt-1"); err != nil {
		t.Fatal(err)
	}
	if body := recorder.Body.String(); !containsAll(body, "id: evt-2", "event: change", "work_item.updated") {
		t.Fatalf("unexpected replay: %s", body)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestTimelineCursorIsOpaqueAndRejectsInvalidValues(t *testing.T) {
	valid := encodeTimelineCursor(time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC), "act-1")
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest("GET", "/?cursor="+valid+"&limit=2", nil)
	cursor, ok := decodeTimelineCursor(context)
	if !ok || cursor.ID != "act-1" {
		t.Fatalf("valid cursor was not decoded: %#v", cursor)
	}
	if limit, ok := timelineLimit(context); !ok || limit != 2 {
		t.Fatalf("invalid limit result: %d %v", limit, ok)
	}
	invalid, _ := gin.CreateTestContext(httptest.NewRecorder())
	invalid.Request = httptest.NewRequest("GET", "/?cursor=not-a-cursor", nil)
	if _, ok := decodeTimelineCursor(invalid); ok {
		t.Fatal("invalid cursor was accepted")
	}
}

func TestNotificationDeliveryIsIdempotentAcrossRetries(t *testing.T) {
	mock := isolatedTenantDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO notifications").WithArgs(sqlmock.AnyArg(), "user-2", "project-1", "Mention", "Review", "comment.mentioned", sqlmock.AnyArg(), "mention:comment-1:user-2").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("ntf-1"))
	mock.ExpectExec("INSERT INTO notification_outbox").WithArgs(sqlmock.AnyArg(), "ntf-1", "user-2", "mention:comment-1:user-2").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("INSERT INTO notifications").WithArgs(sqlmock.AnyArg(), "user-2", "project-1", "Mention", "Review", "comment.mentioned", sqlmock.AnyArg(), "mention:comment-1:user-2").
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectCommit()
	tx, err := db.DB.Begin()
	if err != nil {
		t.Fatal(err)
	}
	if err = createNotification(tx, "user-2", "project-1", "mention:comment-1:user-2", "Mention", "Review", "comment.mentioned", map[string]any{"comment_id": "comment-1"}); err != nil {
		t.Fatal(err)
	}
	if err = createNotification(tx, "user-2", "project-1", "mention:comment-1:user-2", "Mention", "Review", "comment.mentioned", map[string]any{"comment_id": "comment-1"}); err != nil {
		t.Fatal(err)
	}
	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestWatcherNotificationSkipsActorAndUsesTransitionDeduplication(t *testing.T) {
	mock := isolatedTenantDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT user_id FROM work_item_watchers").WithArgs("item-1").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow("actor-1").AddRow("watcher-2"))
	mock.ExpectQuery("INSERT INTO notifications").WithArgs(sqlmock.AnyArg(), "watcher-2", "project-1", "Work item yang diikuti diperbarui", "Buka work item untuk meninjau perubahan terbaru.", "work_item.watcher_updated", sqlmock.AnyArg(), "watcher:transition:item-1:3:watcher-2").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("ntf-2"))
	mock.ExpectExec("INSERT INTO notification_outbox").WithArgs(sqlmock.AnyArg(), "ntf-2", "watcher-2", "watcher:transition:item-1:3:watcher-2").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	tx, err := db.DB.Begin()
	if err != nil {
		t.Fatal(err)
	}
	if err = CreateWatcherNotifications(tx, "project-1", "item-1", "actor-1", "transition:item-1:3"); err != nil {
		t.Fatal(err)
	}
	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func containsAll(value string, values ...string) bool {
	for _, expected := range values {
		if !strings.Contains(value, expected) {
			return false
		}
	}
	return true
}
