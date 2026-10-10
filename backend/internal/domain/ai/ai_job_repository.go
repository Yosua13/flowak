package ai

import "context"

// AIJobRepository defines the persistence contract for AI background jobs.
type AIJobRepository interface {
	CreateJob(ctx context.Context, job *AIJob) error
	GetJobByID(ctx context.Context, id string) (*AIJob, error)
	UpdateJobStatus(ctx context.Context, id string, status string, result any, errorMessage string) error
}
