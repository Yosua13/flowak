package response_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/domain/common"
	"backend/internal/transport/http/response"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestWriteSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	payload := map[string]string{"message": "success"}
	response.WriteSuccess(c, http.StatusOK, payload)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var res map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if res["message"] != "success" {
		t.Errorf("expected message 'success', got %q", res["message"])
	}
}

func TestWriteSuccess_NilData(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	response.WriteSuccess(c, http.StatusNoContent, nil)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected status code %d, got %d", http.StatusNoContent, w.Code)
	}
}

func TestWriteError_Mappings(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:           "ErrNotFound",
			err:            common.ErrNotFound,
			expectedStatus: http.StatusNotFound,
			expectedMsg:    "not found",
		},
		{
			name:           "Wrapped ErrNotFound",
			err:            fmt.Errorf("project %w", common.ErrNotFound),
			expectedStatus: http.StatusNotFound,
			expectedMsg:    "project not found",
		},
		{
			name:           "ErrUnauthorized",
			err:            common.ErrUnauthorized,
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "unauthorized",
		},
		{
			name:           "ErrForbidden",
			err:            common.ErrForbidden,
			expectedStatus: http.StatusForbidden,
			expectedMsg:    "forbidden",
		},
		{
			name:           "ErrConflict",
			err:            common.ErrConflict,
			expectedStatus: http.StatusConflict,
			expectedMsg:    "conflict",
		},
		{
			name:           "ErrValidation",
			err:            common.ErrValidation,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "validation failed",
		},
		{
			name:           "Unknown internal error",
			err:            errors.New("unexpected database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "unexpected database error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			response.WriteError(c, tc.err)

			if w.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, w.Code)
			}

			var body map[string]string
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}
			if body["error"] != tc.expectedMsg {
				t.Errorf("expected error message %q, got %q", tc.expectedMsg, body["error"])
			}
		})
	}
}

func TestWriteError_NilError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	response.WriteError(c, nil)
	if w.Code != http.StatusOK { // default unwritten recorder status is 200
		t.Errorf("expected status 200 for nil error (no write), got %d", w.Code)
	}
	if w.Body.Len() != 0 {
		t.Errorf("expected empty body for nil error, got %s", w.Body.String())
	}
}

func TestWriteValidationError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	response.WriteValidationError(c, "Field email is required")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["error"] != "Field email is required" {
		t.Errorf("expected error %q, got %q", "Field email is required", body["error"])
	}
}

func TestWriteValidationError_DefaultMessage(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	response.WriteValidationError(c, "")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["error"] != "Invalid request body" {
		t.Errorf("expected error %q, got %q", "Invalid request body", body["error"])
	}
}
