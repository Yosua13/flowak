package ai

import (
	"net/http"

	"backend/db"
	domainAI "backend/internal/domain/ai"
	repoAI "backend/internal/repository/postgres/ai"
	usecaseAI "backend/internal/usecase/ai"
	"github.com/gin-gonic/gin"
)

// HandleGetJobStatus handles GET /api/ai/jobs/:id.
// It allows frontend/clients to poll or query the current status and results of an enqueued AI job.
func HandleGetJobStatus(c *gin.Context) {
	jobID := c.Param("id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "job ID is required"})
		return
	}

	var repo domainAI.AIJobRepository
	if db.DB != nil {
		repo = repoAI.NewAIJobRepository(db.DB)
	}

	job, err := usecaseAI.ExecuteGetAIJobStatus(c.Request.Context(), repo, jobID)
	if err != nil || job == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "AI job not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         job.ID,
		"type":       job.Type,
		"status":     job.Status,
		"prompt":     job.Prompt,
		"result":     job.Result,
		"error":      job.ErrorMessage,
		"created_at": job.CreatedAt,
		"updated_at": job.UpdatedAt,
	})
}
