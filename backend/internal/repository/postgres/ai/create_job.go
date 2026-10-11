package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	domainAI "backend/internal/domain/ai"
)

// CreateJob persists a new AI job record into PostgreSQL.
func (r *PostgresAIJobRepository) CreateJob(ctx context.Context, job *domainAI.AIJob) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("database connection is nil")
	}

	payloadJSON, err := json.Marshal(job.Payload)
	if err != nil {
		payloadJSON = []byte("{}")
	}

	now := time.Now().UTC()
	if job.CreatedAt.IsZero() {
		job.CreatedAt = now
	}
	job.UpdatedAt = now

	query := `
		INSERT INTO ai_jobs (id, job_type, status, prompt, payload, result, error_message, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NULL, $6, $7, $8)
	`
	_, err = r.db.ExecContext(ctx, query,
		job.ID,
		job.Type,
		job.Status,
		job.Prompt,
		string(payloadJSON),
		job.ErrorMessage,
		job.CreatedAt,
		job.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert ai_job: %w", err)
	}

	return nil
}
