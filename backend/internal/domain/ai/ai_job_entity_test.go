package ai_test

import (
	"testing"
	"time"

	domainAI "backend/internal/domain/ai"
)

func TestAIJobEntity_ConstantsAndFields(t *testing.T) {
	if domainAI.StatusPending != "pending" {
		t.Errorf("expected StatusPending to be 'pending', got: %s", domainAI.StatusPending)
	}
	if domainAI.StatusProcessing != "processing" {
		t.Errorf("expected StatusProcessing to be 'processing', got: %s", domainAI.StatusProcessing)
	}
	if domainAI.StatusCompleted != "completed" {
		t.Errorf("expected StatusCompleted to be 'completed', got: %s", domainAI.StatusCompleted)
	}
	if domainAI.StatusFailed != "failed" {
		t.Errorf("expected StatusFailed to be 'failed', got: %s", domainAI.StatusFailed)
	}

	job := domainAI.AIJob{
		ID:        "job_1",
		Type:      "generate_flow",
		Status:    domainAI.StatusPending,
		Prompt:    "Test",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if job.ID != "job_1" || job.Type != "generate_flow" || job.Status != domainAI.StatusPending {
		t.Fatalf("unexpected job fields: %+v", job)
	}
}
