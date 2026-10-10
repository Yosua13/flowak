package ai

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	domainAI "backend/internal/domain/ai"
)

// GetJobByID queries an AI job by its unique identifier.
func (r *PostgresAIJobRepository) GetJobByID(ctx context.Context, id string) (*domainAI.AIJob, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	query := `
		SELECT id, job_type, status, prompt, payload, result, error_message, created_at, updated_at
		FROM ai_jobs
		WHERE id = $1
	`
	var job domainAI.AIJob
	var rawPayload, rawResult []byte
	var errMsg sql.NullString
	var prompt sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&job.ID,
		&job.Type,
		&job.Status,
		&prompt,
		&rawPayload,
		&rawResult,
		&errMsg,
		&job.CreatedAt,
		&job.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query ai_job %s: %w", id, err)
	}

	if prompt.Valid {
		job.Prompt = prompt.String
	}
	if errMsg.Valid {
		job.ErrorMessage = errMsg.String
	}

	if len(rawPayload) > 0 {
		var payload any
		_ = json.Unmarshal(rawPayload, &payload)
		job.Payload = payload
	}
	if len(rawResult) > 0 {
		var result any
		_ = json.Unmarshal(rawResult, &result)
		job.Result = result
	}

	return &job, nil
}
