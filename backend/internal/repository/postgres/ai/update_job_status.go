package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// UpdateJobStatus updates the state, result, and optional error of an existing AI job.
func (r *PostgresAIJobRepository) UpdateJobStatus(ctx context.Context, id string, status string, result any, errorMessage string) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("database connection is nil")
	}

	var resultJSON []byte
	if result != nil {
		var err error
		resultJSON, err = json.Marshal(result)
		if err != nil {
			return fmt.Errorf("failed to marshal job result: %w", err)
		}
	}

	now := time.Now().UTC()
	query := `
		UPDATE ai_jobs
		SET status = $2,
		    result = COALESCE($3, result),
		    error_message = $4,
		    updated_at = $5
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query,
		id,
		status,
		resultJSON,
		errorMessage,
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to update ai_job status: %w", err)
	}

	return nil
}
