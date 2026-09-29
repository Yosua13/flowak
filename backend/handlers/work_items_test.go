package handlers

import (
	"strings"
	"testing"

	"backend/models"
)

func stringPtr(value string) *string { return &value }

func TestWorkItemValidationEnforcesExecutionRules(t *testing.T) {
	points := 5
	valid := models.WorkItemRequest{Type: "Bug", Title: "Cannot submit", Description: stringPtr("Checkout scope"), Points: &points}
	if err := validWorkItemInput(valid, true); err != nil {
		t.Fatalf("valid project work item rejected: %v", err)
	}

	blocked := valid
	blocked.Status = "Blocked"
	if err := validWorkItemInput(blocked, true); err == nil {
		t.Fatal("blocked item without a reason was accepted")
	}
	blocked.BlockedReason = stringPtr("Waiting for provider")
	if err := validWorkItemInput(blocked, true); err != nil {
		t.Fatalf("blocked item with a reason rejected: %v", err)
	}

	invalidPoints := valid
	badPoints := 4
	invalidPoints.Points = &badPoints
	if err := validWorkItemInput(invalidPoints, true); err == nil {
		t.Fatal("unsupported story points were accepted")
	}
}

func TestWorkItemTransitionsAreExplicit(t *testing.T) {
	if !statusTransitionAllowed("Ready", "In Progress") {
		t.Fatal("expected Ready -> In Progress to be valid")
	}
	if statusTransitionAllowed("Backlog", "Done") {
		t.Fatal("Backlog -> Done must be rejected")
	}
	if statusTransitionAllowed("Done", "In Progress") {
		t.Fatal("Done items must not reopen without an explicit workflow policy")
	}
}

func TestTransitionStatementUsesStableStatusParameterType(t *testing.T) {
	if !strings.Contains(transitionWorkItemSQL, "status=$1::varchar") || !strings.Contains(transitionWorkItemSQL, "CASE WHEN $1::varchar") {
		t.Fatal("transition SQL must cast the reused status parameter for PostgreSQL prepared statements")
	}
}
