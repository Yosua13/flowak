package ai

import (
	"net/http"
	"strings"

	"backend/config"
	"backend/db"
	domainAI "backend/internal/domain/ai"
	"backend/internal/infra/rabbitmq/producer"
	repoAI "backend/internal/repository/postgres/ai"
	usecaseAI "backend/internal/usecase/ai"
	"backend/middleware"
	"github.com/gin-gonic/gin"
)

// GlobalAIJobPublisher holds the optional RabbitMQ publisher instance for AI jobs.
var GlobalAIJobPublisher *producer.AIJobPublisher

// SetAIJobPublisher configures the active RabbitMQ publisher for AI background jobs.
func SetAIJobPublisher(pub *producer.AIJobPublisher) {
	GlobalAIJobPublisher = pub
}

// HandleGenerateFlow handles POST /api/ai/generate-flow.
// If asynchronous processing is requested via header (Prefer: respond-async) or query (?async=true)
// and RabbitMQ is available, it returns 202 Accepted with job_id immediately.
// Otherwise, it executes the flow synchronously as a fallback without breaking clients.
func HandleGenerateFlow(c *gin.Context) {
	var req struct {
		Prompt string `json:"prompt"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Prompt) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Permintaan prompt tidak boleh kosong"})
		return
	}

	// Verify API key when running without active custom mock
	if config.ActiveConfig.GeminiAPIKey == "" && usecaseAI.DefaultAICaller == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AI service is not configured; create the draft specification manually."})
		return
	}

	userID, _ := middleware.GetUserID(c)
	isAsync := c.Query("async") == "true" ||
		c.GetHeader("Prefer") == "respond-async" ||
		c.GetHeader("X-Async") == "true"

	var repo domainAI.AIJobRepository
	if db.DB != nil {
		repo = repoAI.NewAIJobRepository(db.DB)
	}

	job, err := usecaseAI.ExecuteEnqueueAIFlow(c.Request.Context(), repo, GlobalAIJobPublisher, req.Prompt, userID, isAsync)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Async enqueued response
	if isAsync && job.Status == domainAI.StatusPending {
		c.JSON(http.StatusAccepted, gin.H{
			"job_id":  job.ID,
			"status":  domainAI.StatusPending,
			"message": "AI workflow generation task enqueued",
		})
		return
	}

	// Synchronous / Fallback response: match the exact JSON schema expected by Canvas
	if flow, ok := job.Result.(usecaseAI.GeneratedFlowResult); ok {
		c.JSON(http.StatusOK, gin.H{
			"job_id":      job.ID,
			"status":      "complete",
			"name":        flow.Name,
			"description": flow.Description,
			"nodes":       flow.Nodes,
			"edges":       flow.Edges,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"job_id": job.ID,
		"status": job.Status,
		"result": job.Result,
	})
}
