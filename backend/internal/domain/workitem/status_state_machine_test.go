package workitem_test

import (
	"testing"

	"backend/internal/domain/workitem"
)

func TestCanTransitionStatus(t *testing.T) {
	tests := []struct {
		name     string
		from     workitem.Status
		to       workitem.Status
		expected bool
	}{
		// Identity transitions (always allowed)
		{"Backlog to Backlog", workitem.StatusBacklog, workitem.StatusBacklog, true},
		{"In Progress to In Progress", workitem.StatusInProgress, workitem.StatusInProgress, true},
		{"Done to Done", workitem.StatusDone, workitem.StatusDone, true},

		// From Backlog
		{"Backlog to Ready", workitem.StatusBacklog, workitem.StatusReady, true},
		{"Backlog to Canceled", workitem.StatusBacklog, workitem.StatusCanceled, true},
		{"Backlog to In Progress", workitem.StatusBacklog, workitem.StatusInProgress, false},
		{"Backlog to Done", workitem.StatusBacklog, workitem.StatusDone, false},
		{"Backlog to Blocked", workitem.StatusBacklog, workitem.StatusBlocked, false},

		// From Ready
		{"Ready to In Progress", workitem.StatusReady, workitem.StatusInProgress, true},
		{"Ready to Blocked", workitem.StatusReady, workitem.StatusBlocked, true},
		{"Ready to Canceled", workitem.StatusReady, workitem.StatusCanceled, true},
		{"Ready to Backlog", workitem.StatusReady, workitem.StatusBacklog, false},
		{"Ready to Done", workitem.StatusReady, workitem.StatusDone, false},

		// From In Progress
		{"In Progress to In Review", workitem.StatusInProgress, workitem.StatusInReview, true},
		{"In Progress to Blocked", workitem.StatusInProgress, workitem.StatusBlocked, true},
		{"In Progress to Ready", workitem.StatusInProgress, workitem.StatusReady, true},
		{"In Progress to Canceled", workitem.StatusInProgress, workitem.StatusCanceled, true},
		{"In Progress to Done", workitem.StatusInProgress, workitem.StatusDone, false},
		{"In Progress to Backlog", workitem.StatusInProgress, workitem.StatusBacklog, false},

		// From In Review
		{"In Review to Done", workitem.StatusInReview, workitem.StatusDone, true},
		{"In Review to In Progress", workitem.StatusInReview, workitem.StatusInProgress, true},
		{"In Review to Blocked", workitem.StatusInReview, workitem.StatusBlocked, true},
		{"In Review to Canceled", workitem.StatusInReview, workitem.StatusCanceled, true},
		{"In Review to Backlog", workitem.StatusInReview, workitem.StatusBacklog, false},
		{"In Review to Ready", workitem.StatusInReview, workitem.StatusReady, false},

		// From Blocked
		{"Blocked to Ready", workitem.StatusBlocked, workitem.StatusReady, true},
		{"Blocked to In Progress", workitem.StatusBlocked, workitem.StatusInProgress, true},
		{"Blocked to Canceled", workitem.StatusBlocked, workitem.StatusCanceled, true},
		{"Blocked to Done", workitem.StatusBlocked, workitem.StatusDone, false},
		{"Blocked to Backlog", workitem.StatusBlocked, workitem.StatusBacklog, false},

		// From Done (terminal status)
		{"Done to In Progress", workitem.StatusDone, workitem.StatusInProgress, false},
		{"Done to Ready", workitem.StatusDone, workitem.StatusReady, false},
		{"Done to Canceled", workitem.StatusDone, workitem.StatusCanceled, false},

		// From Canceled (terminal status)
		{"Canceled to Backlog", workitem.StatusCanceled, workitem.StatusBacklog, false},
		{"Canceled to In Progress", workitem.StatusCanceled, workitem.StatusInProgress, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := workitem.CanTransitionStatus(tc.from, tc.to)
			if result != tc.expected {
				t.Errorf("CanTransitionStatus(%s, %s) = %v; want %v", tc.from, tc.to, result, tc.expected)
			}
		})
	}
}
