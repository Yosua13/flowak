package common_test

import (
	"errors"
	"fmt"
	"testing"

	"backend/internal/domain/common"
)

func TestDomainSentinelErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		msg  string
	}{
		{"ErrNotFound", common.ErrNotFound, "not found"},
		{"ErrUnauthorized", common.ErrUnauthorized, "unauthorized"},
		{"ErrForbidden", common.ErrForbidden, "forbidden"},
		{"ErrConflict", common.ErrConflict, "conflict"},
		{"ErrValidation", common.ErrValidation, "validation failed"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err == nil {
				t.Fatalf("expected %s to be non-nil", tc.name)
			}
			if tc.err.Error() != tc.msg {
				t.Errorf("expected error message %q, got %q", tc.msg, tc.err.Error())
			}

			// Verify errors.Is compatibility with wrapped errors
			wrapped := fmt.Errorf("wrapped context: %w", tc.err)
			if !errors.Is(wrapped, tc.err) {
				t.Errorf("expected wrapped error to match %s", tc.name)
			}
		})
	}
}
