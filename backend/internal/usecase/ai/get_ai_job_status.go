package ai

import (
	"context"
	"fmt"

	domainAI "backend/internal/domain/ai"
)

// ExecuteGetAIJobStatus queries the status and results of an enqueued AI background job.
func ExecuteGetAIJobStatus(ctx context.Context, repo domainAI.AIJobRepository, jobID string) (*domainAI.AIJob, error) {
	if repo == nil {
		return nil, fmt.Errorf("ai job repository is not configured")
	}

	job, err := repo.GetJobByID(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve AI job %s: %w", jobID, err)
	}

	return job, nil
}
