package apicontract_test

import (
	"context"
	"testing"

	"backend/internal/infra/rabbitmq/producer"
	usecaseAPI "backend/internal/usecase/apicontract"
	"backend/models"
	"github.com/DATA-DOG/go-sqlmock"
)

func TestExecuteEnqueueAPIRun_AsyncMode(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	pub := producer.NewAPIRunnerJobPublisher(nil) // nil channel falls back gracefully

	input := models.APIRunRequest{
		EnvironmentID: "env_123",
		Method:        "GET",
		RelativePath:  "/v1/users",
	}

	mock.ExpectQuery("SELECT r.method, r.relative_path, e.approved_base_url FROM api_requests").
		WithArgs("req_123", "env_123").
		WillReturnRows(sqlmock.NewRows([]string{"method", "relative_path", "approved_base_url"}).
			AddRow("GET", "/v1/users", "https://api.example.com"))

	mock.ExpectExec("INSERT INTO api_runs").
		WithArgs(sqlmock.AnyArg(), "req_123", "env_123", "usr_1", sqlmock.AnyArg(), "api.example.com").
		WillReturnResult(sqlmock.NewResult(1, 1))

	result, err := usecaseAPI.ExecuteEnqueueAPIRun(ctx, db, pub, "req_123", input, "usr_1", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil || result.ID == "" {
		t.Fatal("expected valid result with non-empty ID")
	}

	if result.Status != "pending" {
		t.Fatalf("expected status 'pending', got: %s", result.Status)
	}
}
