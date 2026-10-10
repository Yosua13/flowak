package ai_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	handlerAI "backend/internal/transport/http/handler/ai"
	usecaseAI "backend/internal/usecase/ai"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupAITestRouter() *gin.Engine {
	r := gin.New()
	r.POST("/api/ai/generate-flow", handlerAI.HandleGenerateFlow)
	r.GET("/api/ai/jobs/:id", handlerAI.HandleGetJobStatus)
	return r
}

func TestHandleGenerateFlow_EmptyPrompt(t *testing.T) {
	router := setupAITestRouter()

	body, _ := json.Marshal(map[string]string{"prompt": ""})
	req, _ := http.NewRequest(http.MethodPost, "/api/ai/generate-flow", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request, got: %d", w.Code)
	}
}

func TestHandleGenerateFlow_AsyncRequest(t *testing.T) {
	router := setupAITestRouter()

	// Mock caller to avoid external Gemini API call
	originalCaller := usecaseAI.DefaultAICaller
	usecaseAI.DefaultAICaller = func(systemInstruction, prompt string, schema any) (string, error) {
		return `{"name":"Test","description":"Desc","nodes":[],"edges":[]}`, nil
	}
	defer func() { usecaseAI.DefaultAICaller = originalCaller }()

	body, _ := json.Marshal(map[string]string{"prompt": "Buatkan alur verifikasi email"})
	req, _ := http.NewRequest(http.MethodPost, "/api/ai/generate-flow?async=true", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Since GlobalAIJobPublisher is nil in test, it falls back smoothly to synchronous 200 OK
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK fallback, got: %d (%s)", w.Code, w.Body.String())
	}

	var res map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if res["name"] != "Test" {
		t.Fatalf("expected flow name 'Test', got: %v", res["name"])
	}
}

func TestHandleGetJobStatus_NotFound(t *testing.T) {
	router := setupAITestRouter()

	req, _ := http.NewRequest(http.MethodGet, "/api/ai/jobs/non_existent_job", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 Not Found, got: %d", w.Code)
	}
}
