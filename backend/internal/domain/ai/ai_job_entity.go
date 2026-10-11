package ai

import "time"

// AI Job status constants
const (
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusCompleted  = "completed"
	StatusFailed     = "failed"
)

// AIJob represents an asynchronous AI generation or audit work task.
type AIJob struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"`
	Status       string    `json:"status"`
	Prompt       string    `json:"prompt,omitempty"`
	Payload      any       `json:"payload,omitempty"`
	Result       any       `json:"result,omitempty"`
	ErrorMessage string    `json:"error_message,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
