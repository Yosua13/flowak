package ai_test

import (
	"context"
	"testing"
	"time"

	domainAI "backend/internal/domain/ai"
	repoAI "backend/internal/repository/postgres/ai"
	"github.com/DATA-DOG/go-sqlmock"
)

func TestPostgresAIJobRepository_CreateAndGet(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	repo := repoAI.NewAIJobRepository(db)
	ctx := context.Background()

	job := &domainAI.AIJob{
		ID:        "job_test_123",
		Type:      "generate_flow",
		Status:    domainAI.StatusPending,
		Prompt:    "Test Prompt",
		Payload:   map[string]any{"foo": "bar"},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	mock.ExpectExec("INSERT INTO ai_jobs").
		WithArgs(job.ID, job.Type, job.Status, job.Prompt, `{"foo":"bar"}`, job.ErrorMessage, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repo.CreateJob(ctx, job); err != nil {
		t.Fatalf("CreateJob failed: %v", err)
	}

	mock.ExpectQuery("SELECT id, job_type, status, prompt, payload, result, error_message, created_at, updated_at FROM ai_jobs WHERE id = \\$1").
		WithArgs(job.ID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "job_type", "status", "prompt", "payload", "result", "error_message", "created_at", "updated_at",
		}).AddRow(job.ID, job.Type, job.Status, job.Prompt, []byte(`{"foo":"bar"}`), nil, nil, job.CreatedAt, job.UpdatedAt))

	fetched, err := repo.GetJobByID(ctx, job.ID)
	if err != nil {
		t.Fatalf("GetJobByID failed: %v", err)
	}
	if fetched == nil || fetched.ID != job.ID {
		t.Fatalf("expected job ID %s, got %v", job.ID, fetched)
	}

	mock.ExpectExec("UPDATE ai_jobs SET status = \\$2").
		WithArgs(job.ID, domainAI.StatusCompleted, []byte(`{"status":"complete"}`), "", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repo.UpdateJobStatus(ctx, job.ID, domainAI.StatusCompleted, map[string]any{"status": "complete"}, ""); err != nil {
		t.Fatalf("UpdateJobStatus failed: %v", err)
	}
}
